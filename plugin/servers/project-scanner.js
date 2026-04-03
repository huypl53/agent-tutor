'use strict';

const fs = require('fs');
const path = require('path');

const IGNORE_DIRS = new Set([
  'node_modules', '.git', 'vendor', 'target', '.agent-tutor',
  'dist', 'build', '.next', '__pycache__', '.venv', 'venv',
  '.tox', '.mypy_cache', '.pytest_cache', 'coverage',
]);

const SOURCE_EXTS = new Set([
  '.js', '.ts', '.jsx', '.tsx', '.py', '.go', '.rs', '.java',
  '.rb', '.c', '.cpp', '.h', '.cs', '.dart', '.swift', '.kt',
  '.gd', '.tf', '.yaml', '.yml', '.toml', '.json', '.md',
]);

class ProjectScanner {
  constructor(projectDir) {
    this._dir = projectDir;
    this._csvPath = path.join(__dirname, '..', 'data', 'project-types.csv');
  }

  loadProjectTypes() {
    const raw = fs.readFileSync(this._csvPath, 'utf8');
    const lines = raw.trim().split('\n');
    const headers = this._parseCSVLine(lines[0]);
    return lines.slice(1).filter(l => l.trim()).map(line => {
      const values = this._parseCSVLine(line);
      const obj = {};
      headers.forEach((h, i) => { obj[h] = values[i] || ''; });
      return obj;
    });
  }

  _parseCSVLine(line) {
    const result = [];
    let current = '';
    let inQuotes = false;
    for (const ch of line) {
      if (ch === '"') { inQuotes = !inQuotes; continue; }
      if (ch === ',' && !inQuotes) { result.push(current.trim()); current = ''; continue; }
      current += ch;
    }
    result.push(current.trim());
    return result;
  }

  detectProjectType() {
    const types = this.loadProjectTypes();
    const existingFiles = this._listTopLevel();
    const fileContents = this._readManifestContents();

    // full-stack: check for frontend+backend subdirectories
    const fullStack = types.find(t => t.type_id === 'full-stack');
    if (fullStack) {
      const fsDirs = fullStack.critical_dirs.split(',').map(d => d.trim().replace(/\/$/, ''));
      const matchedDirs = fsDirs.filter(d => existingFiles.dirs.includes(d));
      if (matchedDirs.length >= 2) return fullStack;
    }

    for (const t of types) {
      if (t.type_id === 'full-stack') continue;
      const keyFiles = t.key_files.split(',').map(f => f.trim()).filter(Boolean);
      for (const pattern of keyFiles) {
        if (this._matchesPattern(pattern, existingFiles, fileContents)) {
          return t;
        }
      }
    }
    return null;
  }

  _matchesPattern(pattern, existingFiles, fileContents) {
    // Check as directory name
    if (existingFiles.dirs.includes(pattern.replace(/\/$/, ''))) return true;
    // Check as file (with glob)
    if (pattern.includes('*')) {
      if (pattern.startsWith('*')) {
        // Pattern like *.ext — match by suffix
        const suffix = pattern.slice(1);
        return existingFiles.files.some(f => f.endsWith(suffix));
      }
      const base = pattern.replace(/\*.*$/, '');
      return existingFiles.files.some(f => f.startsWith(base));
    }
    // Check as exact file
    if (existingFiles.files.includes(pattern)) return true;
    // Check as dependency name in manifests
    if (fileContents.allDeps && fileContents.allDeps.includes(pattern)) return true;
    return false;
  }

  _listTopLevel() {
    const entries = fs.readdirSync(this._dir, { withFileTypes: true });
    const files = [];
    const dirs = [];
    for (const e of entries) {
      if (IGNORE_DIRS.has(e.name)) continue;
      if (e.isDirectory()) dirs.push(e.name);
      else files.push(e.name);
    }
    return { files, dirs };
  }

  _readManifestContents() {
    const allDeps = [];

    // package.json deps
    try {
      const pkg = JSON.parse(fs.readFileSync(path.join(this._dir, 'package.json'), 'utf8'));
      allDeps.push(...Object.keys(pkg.dependencies || {}));
      allDeps.push(...Object.keys(pkg.devDependencies || {}));
    } catch {}

    // requirements.txt
    try {
      const req = fs.readFileSync(path.join(this._dir, 'requirements.txt'), 'utf8');
      req.split('\n').forEach(l => {
        const name = l.trim().split(/[=<>!~\s]/)[0];
        if (name) allDeps.push(name);
      });
    } catch {}

    // go.mod
    try {
      const gomod = fs.readFileSync(path.join(this._dir, 'go.mod'), 'utf8');
      const reqMatch = gomod.match(/require\s+\(([^)]+)\)/s);
      if (reqMatch) {
        reqMatch[1].split('\n').forEach(l => {
          const name = l.trim().split(/\s/)[0];
          if (name) allDeps.push(name);
        });
      }
      const singleReq = gomod.match(/^require\s+(\S+)/gm);
      if (singleReq) {
        singleReq.forEach(r => allDeps.push(r.replace(/^require\s+/, '')));
      }
    } catch {}

    // Cargo.toml [dependencies]
    try {
      const cargo = fs.readFileSync(path.join(this._dir, 'Cargo.toml'), 'utf8');
      const depSection = cargo.match(/\[dependencies\]\n([\s\S]*?)(\n\[|$)/);
      if (depSection) {
        depSection[1].split('\n').forEach(l => {
          const name = l.trim().split(/\s*=/)[0];
          if (name) allDeps.push(name);
        });
      }
    } catch {}

    // pyproject.toml
    try {
      const pyproj = fs.readFileSync(path.join(this._dir, 'pyproject.toml'), 'utf8');
      const depSection = pyproj.match(/\[project\.dependencies\]\n([\s\S]*?)(\n\[|$)/);
      if (depSection) {
        depSection[1].split('\n').forEach(l => {
          const name = l.trim().split(/\s*[=<>!~]/)[0];
          if (name) allDeps.push(name);
        });
      }
    } catch {}

    return { allDeps };
  }

  parseManifests() {
    // Try package.json
    try {
      const pkg = JSON.parse(fs.readFileSync(path.join(this._dir, 'package.json'), 'utf8'));
      return {
        name: pkg.name,
        version: pkg.version,
        dependencies: pkg.dependencies || {},
        devDependencies: pkg.devDependencies || {},
      };
    } catch {}

    // Try requirements.txt
    try {
      const req = fs.readFileSync(path.join(this._dir, 'requirements.txt'), 'utf8');
      const deps = {};
      req.split('\n').forEach(l => {
        const parts = l.trim().split(/([=<>!~]+)/);
        if (parts[0]) deps[parts[0]] = parts.slice(1).join('') || '*';
      });
      return { dependencies: deps };
    } catch {}

    // Try go.mod
    try {
      const gomod = fs.readFileSync(path.join(this._dir, 'go.mod'), 'utf8');
      const modMatch = gomod.match(/^module\s+(\S+)/m);
      const goMatch = gomod.match(/^go\s+(\S+)/m);
      const deps = {};
      const reqBlock = gomod.match(/require\s+\(([^)]+)\)/s);
      if (reqBlock) {
        reqBlock[1].split('\n').forEach(l => {
          const parts = l.trim().split(/\s+/);
          if (parts[0]) deps[parts[0]] = parts[1] || '*';
        });
      }
      // Single-line requires
      for (const m of gomod.matchAll(/^require\s+(\S+)\s+(\S+)/gm)) {
        deps[m[1]] = m[2];
      }
      return {
        name: modMatch ? modMatch[1] : undefined,
        goVersion: goMatch ? goMatch[1] : undefined,
        dependencies: deps,
      };
    } catch {}

    // Try Cargo.toml
    try {
      const cargo = fs.readFileSync(path.join(this._dir, 'Cargo.toml'), 'utf8');
      const nameMatch = cargo.match(/name\s*=\s*"([^"]+)"/);
      const deps = {};
      const depSection = cargo.match(/\[dependencies\]\n([\s\S]*?)(\n\[|$)/);
      if (depSection) {
        depSection[1].split('\n').forEach(l => {
          const m = l.match(/^(\w[\w-]*)\s*=\s*"?([^"\n]+)"?/);
          if (m) deps[m[1]] = m[2];
        });
      }
      return { name: nameMatch ? nameMatch[1] : undefined, dependencies: deps };
    } catch {}

    // Try pyproject.toml
    try {
      const pyproj = fs.readFileSync(path.join(this._dir, 'pyproject.toml'), 'utf8');
      const nameMatch = pyproj.match(/name\s*=\s*"([^"]+)"/);
      return { name: nameMatch ? nameMatch[1] : undefined, dependencies: {} };
    } catch {}

    return {};
  }

  scanStructure() {
    const directories = [];
    const files = [];
    const entryPoints = [];
    this._walkDir(this._dir, '', directories, files, 0, 3);

    // Detect entry points
    const commonEntries = ['index', 'main', 'app', 'server', 'cli'];
    for (const f of files) {
      const base = path.basename(f, path.extname(f));
      if (commonEntries.includes(base) && SOURCE_EXTS.has(path.extname(f))) {
        entryPoints.push(f);
      }
    }

    return { directories, files, entryPoints };
  }

  _walkDir(dir, rel, directories, files, depth, maxDepth) {
    if (depth > maxDepth) return;
    let entries;
    try { entries = fs.readdirSync(dir, { withFileTypes: true }); }
    catch { return; }

    for (const e of entries) {
      if (IGNORE_DIRS.has(e.name)) continue;
      if (e.name.startsWith('.') && e.name !== '.github') continue;
      const relPath = rel ? `${rel}/${e.name}` : e.name;
      if (e.isDirectory()) {
        directories.push({ name: relPath, depth });
        this._walkDir(path.join(dir, e.name), relPath, directories, files, depth + 1, maxDepth);
      } else if (SOURCE_EXTS.has(path.extname(e.name))) {
        files.push(relPath);
      }
    }
  }

  scan() {
    const typeInfo = this.detectProjectType();
    const manifest = this.parseManifests();
    const structure = this.scanStructure();

    const stack = {};
    if (manifest.dependencies) {
      // Infer language/framework from manifest
      const deps = { ...manifest.dependencies, ...(manifest.devDependencies || {}) };
      if (deps.react || deps['react-dom']) stack.framework = 'React';
      else if (deps.vue) stack.framework = 'Vue';
      else if (deps.express) stack.framework = 'Express';
      else if (deps.fastapi) stack.framework = 'FastAPI';
      else if (deps.django) stack.framework = 'Django';
      else if (deps.flask) stack.framework = 'Flask';
      else if (deps['@angular/core']) stack.framework = 'Angular';
      else if (deps.svelte) stack.framework = 'Svelte';
      else if (deps.next) stack.framework = 'Next.js';
    }
    if (manifest.goVersion) stack.language = 'Go';
    else if (manifest.dependencies && (manifest.dependencies.fastapi || manifest.dependencies.django || manifest.dependencies.flask)) stack.language = 'Python';
    else if (manifest.dependencies && manifest.dependencies.express) stack.language = 'JavaScript/TypeScript';

    return {
      projectType: typeInfo ? typeInfo.type_id : 'unknown',
      displayName: typeInfo ? typeInfo.display_name : 'Unknown Project',
      stack,
      structure: structure.directories.length > 10 ? 'large' : 'small',
      entryPoints: structure.entryPoints,
      directories: structure.directories,
      files: structure.files,
      manifest,
      analysisDomains: typeInfo ? typeInfo.analysis_domains.split(',').map(d => d.trim()) : ['architecture', 'dev-setup'],
      scannedAt: new Date().toISOString(),
    };
  }
}

module.exports = { ProjectScanner };
