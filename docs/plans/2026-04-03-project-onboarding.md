# Project Onboarding Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add project analysis tools (`scan_project`, `get_project_profile`, `save_project_doc`) and two skills (`/atu:onboard`, `/atu:deep-dive`) that help students understand existing codebases through automated analysis and sub-agent-driven reasoning.

**Architecture:** MCP server does fast mechanical extraction (file tree, manifest parsing, project type detection via CSV). Sub-agents spawned by the `/atu:onboard` skill handle deep analysis per domain (architecture, API, data, etc.) and write markdown docs to `.agent-tutor/docs/`. StateManager gets v1→v2 migration with a `project` key.

**Tech Stack:** Node.js, `node:test`, `@modelcontextprotocol/sdk`, `zod`, CSV parsing (built-in)

**Working directory:** Use a worktree branched from `master`

**Reference:** Design doc at `docs/plans/2026-04-03-project-onboarding-design.md`

---

### Task 1: Create project-types.csv

**Files:**
- Create: `plugin/data/project-types.csv`

**Step 1: Create the data directory**

```bash
mkdir -p plugin/data
```

**Step 2: Write the CSV file**

Create `plugin/data/project-types.csv` with this content:

```csv
type_id,display_name,key_files,critical_dirs,analysis_domains,test_patterns,entry_points
web-app,Web Application,"package.json,next.config.*,vite.config.*,nuxt.config.*,svelte.config.*,angular.json,astro.config.*","src/,app/,pages/,components/","architecture,components,state,routing,testing,dev-setup","*.test.{ts,tsx,js,jsx},*.spec.{ts,tsx,js,jsx}","index.{ts,tsx,js,jsx},main.{ts,tsx,js,jsx},app.{ts,tsx,js,jsx}"
backend-api,Backend API,"server.js,app.js,manage.py,main.go,cmd/,Gemfile,build.gradle,pom.xml","src/,app/,api/,routes/,controllers/,services/,handlers/","architecture,api,data,testing,dev-setup","*.test.{ts,js},*_test.go,test_*.py,*_test.rb,*Test.java","server.{ts,js},app.{ts,js,py},main.{go,py,ts,js},index.{ts,js}"
full-stack,Full-Stack Application,"","client/,server/,frontend/,backend/,apps/,packages/","architecture,api,data,components,testing,dev-setup","","" 
cli-tool,CLI Tool,"bin/,commander,click,cobra,clap","src/,cmd/,lib/","architecture,testing,dev-setup","*.test.{ts,js},*_test.go,test_*.py","cli.{ts,js,py},main.{go,py,ts,js,rs}"
library,Library/Package,"index.{ts,js,mjs},lib.rs,setup.py,setup.cfg","src/,lib/","architecture,api,testing,dev-setup","*.test.{ts,js},*_test.go,test_*.py,*_test.rs","index.{ts,js,mjs},lib.{rs,py},mod.rs"
mobile-app,Mobile Application,"react-native.config.*,expo,pubspec.yaml,*.xcodeproj,AndroidManifest.xml","src/,app/,lib/,screens/,components/","architecture,components,state,data,testing,dev-setup","*.test.{ts,tsx,js,jsx},*_test.dart","App.{tsx,jsx,ts,js},main.dart,index.{tsx,jsx}"
desktop-app,Desktop Application,"electron,tauri.conf.json,*.pro,CMakeLists.txt","src/,src-tauri/","architecture,components,testing,dev-setup","*.test.{ts,js},*_test.*","main.{ts,js,py,cpp},index.{ts,js}"
game,Game,"ProjectSettings/,*.godot,Cargo.toml,pygame","Assets/,src/,scripts/,scenes/","architecture,components,testing,dev-setup","*.test.*,test_*.*","main.{gd,rs,py,cs,cpp}"
data-pipeline,Data Pipeline,"airflow,dbt_project.yml,pipeline,dagster,prefect","dags/,models/,pipelines/,notebooks/","architecture,data,infra,testing,dev-setup","test_*.py,*_test.py","dag*.py,pipeline*.py,main.py"
extension,Extension/Plugin,"manifest.json,package.json,.vscodeignore,plugin.json","src/,extension/,content/,background/","architecture,api,testing,dev-setup","*.test.{ts,js}","extension.{ts,js},background.{ts,js},content.{ts,js}"
infra,Infrastructure,"*.tf,pulumi.*,template.yaml,ansible,playbook","modules/,stacks/,roles/,templates/","architecture,infra,dev-setup","*.tftest.*,test_*.py","main.tf,Pulumi.*,template.yaml"
embedded,Embedded System,"platformio.ini,*.ino,CMakeLists.txt,Zephyr,Makefile","src/,include/,drivers/,hal/","architecture,api,testing,dev-setup","test_*.*,*_test.*","main.{c,cpp,ino},app_main.c"
ai-llm,AI/LLM Application,"langchain,llama_index,transformers,openai,anthropic,mcp,autogen,crewai","agents/,chains/,prompts/,tools/,models/,pipelines/","architecture,api,data,infra,testing,dev-setup","test_*.py,*_test.py,*.test.{ts,js}","main.py,app.py,agent.py,chain.py,server.{py,ts,js}"
devops,DevOps/Platform,".github/workflows/,Dockerfile,docker-compose.*,k8s/,helm/,Jenkinsfile,.gitlab-ci.yml","charts/,k8s/,terraform/,ansible/,.github/","architecture,infra,dev-setup","*.tftest.*,test_*.*","Dockerfile,docker-compose.yml,main.tf"
```

**Step 3: Commit**

```bash
git add plugin/data/project-types.csv
git commit -m "feat: add project-types.csv with 14 project type definitions"
```

---

### Task 2: Create ProjectScanner module

**Files:**
- Create: `test/project-scanner.test.js`
- Create: `plugin/servers/project-scanner.js`

**Step 1: Write failing tests for CSV loading and project type detection**

Create `test/project-scanner.test.js`:

```javascript
const { describe, it, beforeEach, afterEach } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('fs');
const path = require('path');
const os = require('os');
const { ProjectScanner } = require('../plugin/servers/project-scanner');

describe('ProjectScanner', () => {
  let tmpDir;
  let scanner;

  beforeEach(() => {
    tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'atu-scan-'));
    scanner = new ProjectScanner(tmpDir);
  });

  afterEach(() => {
    fs.rmSync(tmpDir, { recursive: true, force: true });
  });

  describe('loadProjectTypes', () => {
    it('loads all 14 project types from CSV', () => {
      const types = scanner.loadProjectTypes();
      assert.equal(types.length, 14);
      assert.ok(types.find(t => t.type_id === 'web-app'));
      assert.ok(types.find(t => t.type_id === 'ai-llm'));
      assert.ok(types.find(t => t.type_id === 'devops'));
    });

    it('parses CSV fields correctly', () => {
      const types = scanner.loadProjectTypes();
      const webApp = types.find(t => t.type_id === 'web-app');
      assert.equal(webApp.display_name, 'Web Application');
      assert.ok(webApp.key_files.includes('package.json'));
      assert.ok(webApp.analysis_domains.includes('components'));
    });
  });

  describe('detectProjectType', () => {
    it('detects web-app from package.json + vite config', () => {
      fs.writeFileSync(path.join(tmpDir, 'package.json'), '{}');
      fs.writeFileSync(path.join(tmpDir, 'vite.config.ts'), '');
      const result = scanner.detectProjectType();
      assert.equal(result.type_id, 'web-app');
    });

    it('detects backend-api from server.js', () => {
      fs.writeFileSync(path.join(tmpDir, 'package.json'), '{}');
      fs.writeFileSync(path.join(tmpDir, 'server.js'), '');
      const result = scanner.detectProjectType();
      assert.equal(result.type_id, 'backend-api');
    });

    it('detects cli-tool from bin directory', () => {
      fs.mkdirSync(path.join(tmpDir, 'bin'));
      fs.writeFileSync(path.join(tmpDir, 'bin', 'cli.js'), '');
      fs.writeFileSync(path.join(tmpDir, 'package.json'), '{}');
      const result = scanner.detectProjectType();
      assert.equal(result.type_id, 'cli-tool');
    });

    it('detects ai-llm from langchain dependency', () => {
      fs.writeFileSync(path.join(tmpDir, 'requirements.txt'), 'langchain\nopenai\n');
      const result = scanner.detectProjectType();
      assert.equal(result.type_id, 'ai-llm');
    });

    it('detects devops from Dockerfile + docker-compose', () => {
      fs.writeFileSync(path.join(tmpDir, 'Dockerfile'), 'FROM node:20');
      fs.writeFileSync(path.join(tmpDir, 'docker-compose.yml'), '');
      fs.mkdirSync(path.join(tmpDir, '.github', 'workflows'), { recursive: true });
      const result = scanner.detectProjectType();
      assert.equal(result.type_id, 'devops');
    });

    it('detects infra from terraform files', () => {
      fs.writeFileSync(path.join(tmpDir, 'main.tf'), '');
      fs.mkdirSync(path.join(tmpDir, 'modules'));
      const result = scanner.detectProjectType();
      assert.equal(result.type_id, 'infra');
    });

    it('returns null when no type matches', () => {
      const result = scanner.detectProjectType();
      assert.equal(result, null);
    });
  });

  describe('parseManifests', () => {
    it('parses package.json', () => {
      fs.writeFileSync(path.join(tmpDir, 'package.json'), JSON.stringify({
        name: 'my-app',
        dependencies: { express: '^4.18.0', pg: '^8.0.0' },
        devDependencies: { jest: '^29.0.0' },
      }));
      const result = scanner.parseManifests();
      assert.equal(result.name, 'my-app');
      assert.ok(result.dependencies.express);
      assert.ok(result.devDependencies.jest);
    });

    it('parses requirements.txt', () => {
      fs.writeFileSync(path.join(tmpDir, 'requirements.txt'), 'fastapi==0.100.0\nsqlalchemy>=2.0\npytest\n');
      const result = scanner.parseManifests();
      assert.ok(result.dependencies.fastapi);
    });

    it('parses go.mod', () => {
      fs.writeFileSync(path.join(tmpDir, 'go.mod'), 'module github.com/user/app\n\ngo 1.21\n\nrequire github.com/gin-gonic/gin v1.9.0\n');
      const result = scanner.parseManifests();
      assert.equal(result.goVersion, '1.21');
      assert.ok(result.dependencies['github.com/gin-gonic/gin']);
    });

    it('parses Cargo.toml', () => {
      fs.writeFileSync(path.join(tmpDir, 'Cargo.toml'), '[package]\nname = "myapp"\nversion = "0.1.0"\n\n[dependencies]\ntokio = "1"\n');
      const result = scanner.parseManifests();
      assert.equal(result.name, 'myapp');
      assert.ok(result.dependencies.tokio);
    });

    it('parses pyproject.toml', () => {
      fs.writeFileSync(path.join(tmpDir, 'pyproject.toml'), '[project]\nname = "myapp"\n\n[project.dependencies]\ndjango = ">=4.0"\n');
      const result = scanner.parseManifests();
      assert.equal(result.name, 'myapp');
    });

    it('returns empty object when no manifest exists', () => {
      const result = scanner.parseManifests();
      assert.deepEqual(result, {});
    });
  });

  describe('scanStructure', () => {
    it('returns file tree with directory categorization', () => {
      fs.mkdirSync(path.join(tmpDir, 'src', 'routes'), { recursive: true });
      fs.mkdirSync(path.join(tmpDir, 'test'));
      fs.writeFileSync(path.join(tmpDir, 'src', 'index.js'), '');
      fs.writeFileSync(path.join(tmpDir, 'src', 'routes', 'api.js'), '');
      fs.writeFileSync(path.join(tmpDir, 'test', 'api.test.js'), '');
      const result = scanner.scanStructure();
      assert.ok(result.directories.length > 0);
      assert.ok(result.files.length > 0);
      assert.ok(result.entryPoints.length > 0);
    });

    it('ignores node_modules and .git', () => {
      fs.mkdirSync(path.join(tmpDir, 'node_modules', 'foo'), { recursive: true });
      fs.mkdirSync(path.join(tmpDir, '.git', 'objects'), { recursive: true });
      fs.writeFileSync(path.join(tmpDir, 'index.js'), '');
      const result = scanner.scanStructure();
      const dirs = result.directories.map(d => d.name);
      assert.ok(!dirs.includes('node_modules'));
      assert.ok(!dirs.includes('.git'));
    });
  });

  describe('scan (full)', () => {
    it('returns complete project profile', () => {
      fs.writeFileSync(path.join(tmpDir, 'package.json'), JSON.stringify({
        name: 'test-api',
        dependencies: { express: '^4.18.0' },
      }));
      fs.writeFileSync(path.join(tmpDir, 'server.js'), 'const express = require("express")');
      fs.mkdirSync(path.join(tmpDir, 'src'));
      fs.writeFileSync(path.join(tmpDir, 'src', 'index.js'), '');

      const profile = scanner.scan();
      assert.equal(profile.projectType, 'backend-api');
      assert.ok(profile.stack);
      assert.ok(profile.structure);
      assert.ok(profile.entryPoints);
      assert.ok(profile.scannedAt);
      assert.ok(profile.analysisDomains);
    });

    it('returns generic profile when no type matches', () => {
      fs.writeFileSync(path.join(tmpDir, 'README.md'), '# Hello');
      const profile = scanner.scan();
      assert.equal(profile.projectType, 'unknown');
      assert.ok(profile.structure);
    });
  });
});
```

**Step 2: Run tests to verify they fail**

```bash
cd <worktree> && npm test
```

Expected: FAIL with `Cannot find module '../plugin/servers/project-scanner'`.

**Step 3: Implement ProjectScanner**

Create `plugin/servers/project-scanner.js`:

```javascript
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
```

**Step 4: Run tests**

```bash
cd <worktree> && npm test
```

Expected: All new ProjectScanner tests pass plus existing 51 tests.

**Step 5: Commit**

```bash
git add plugin/servers/project-scanner.js test/project-scanner.test.js
git commit -m "feat: add ProjectScanner module with type detection and manifest parsing"
```

---

### Task 3: Add project methods to StateManager (v1→v2 migration)

**Files:**
- Modify: `plugin/servers/state-manager.js:16-21` (EMPTY_STATE)
- Modify: `plugin/servers/state-manager.js:29-38` (readState)
- Modify: `test/state-manager.test.js`

**Step 1: Write failing tests for v2 migration and project methods**

Add to `test/state-manager.test.js`, inside the top-level `describe('StateManager')` block, after the `migration` describe block (after line 476):

```javascript
  describe('project profile', () => {
    it('auto-migrates v1 state to v2 on read', async () => {
      const atuDir = path.join(tmpDir, '.agent-tutor');
      fs.mkdirSync(atuDir, { recursive: true });
      fs.writeFileSync(path.join(atuDir, 'state.json'), JSON.stringify({
        version: 1,
        topics: { a: { id: 'a', title: 'A', status: 'introduced', complexity: null, dependencies: [], started: '2026-01-01', moments: [], lessonFile: null } },
        plan: null,
        session: null,
      }));
      const state = await sm.readState();
      assert.equal(state.version, 2);
      assert.equal(state.project, null);
      // existing data preserved
      assert.equal(state.topics.a.title, 'A');
    });

    it('saveProjectProfile stores and getProjectProfile retrieves', async () => {
      const profile = { projectType: 'backend-api', stack: { language: 'Python' }, scannedAt: '2026-04-03T00:00:00Z' };
      await sm.saveProjectProfile(profile);
      const loaded = await sm.getProjectProfile();
      assert.equal(loaded.projectType, 'backend-api');
      assert.equal(loaded.stack.language, 'Python');
    });

    it('getProjectProfile returns null when no profile exists', async () => {
      const profile = await sm.getProjectProfile();
      assert.equal(profile, null);
    });

    it('saveProjectDoc writes file and updates state', async () => {
      await sm.saveProjectDoc('architecture', '# Architecture\n\nMVC pattern.');
      const docPath = path.join(tmpDir, '.agent-tutor', 'docs', 'architecture.md');
      assert.ok(fs.existsSync(docPath));
      assert.equal(fs.readFileSync(docPath, 'utf8'), '# Architecture\n\nMVC pattern.');
      const state = await sm.readState();
      assert.ok(state.project.docs.includes('architecture'));
    });

    it('getProjectDoc returns doc content or null', async () => {
      assert.equal(await sm.getProjectDoc('nope'), null);
      await sm.saveProjectDoc('api', '# API');
      assert.equal(await sm.getProjectDoc('api'), '# API');
    });

    it('listProjectDocs returns doc names', async () => {
      await sm.saveProjectDoc('arch', '# A');
      await sm.saveProjectDoc('data', '# D');
      const docs = await sm.listProjectDocs();
      assert.deepEqual(docs.sort(), ['arch', 'data']);
    });
  });
```

**Step 2: Run tests to verify they fail**

```bash
cd <worktree> && npm test
```

Expected: New tests FAIL with `state.project` undefined or `sm.saveProjectProfile is not a function`.

**Step 3: Update EMPTY_STATE and readState for v2**

In `plugin/servers/state-manager.js`, change `EMPTY_STATE` (lines 16-21):

```javascript
const EMPTY_STATE = {
  version: 2,
  topics: {},
  plan: null,
  session: null,
  project: null,
};
```

Update `readState` (lines 29-38) to auto-migrate v1→v2:

```javascript
  async readState() {
    try {
      const raw = fs.readFileSync(this._file, 'utf8');
      const state = JSON.parse(raw);
      if (state.version === 1) {
        state.version = 2;
        state.project = state.project || null;
        await this.writeState(state);
      }
      return state;
    } catch (err) {
      if (err.code === 'ENOENT') {
        return JSON.parse(JSON.stringify(EMPTY_STATE));
      }
      throw err;
    }
  }
```

**Step 4: Add project methods to StateManager**

Add after `getLearningSummary` method (after line 227), before `migrateIfNeeded`:

```javascript
  async saveProjectProfile(profile) {
    const state = await this.readState();
    state.project = { ...profile, docs: state.project?.docs || [], deepDives: state.project?.deepDives || [] };
    await this.writeState(state);
    return state.project;
  }

  async getProjectProfile() {
    const state = await this.readState();
    return state.project;
  }

  async saveProjectDoc(name, content) {
    const docsDir = path.join(this._dir, 'docs');
    fs.mkdirSync(docsDir, { recursive: true });
    const filePath = path.join(docsDir, `${name}.md`);
    fs.writeFileSync(filePath, content);

    const state = await this.readState();
    if (!state.project) {
      state.project = { docs: [], deepDives: [] };
    }
    if (!state.project.docs.includes(name)) {
      state.project.docs.push(name);
    }
    await this.writeState(state);
    return filePath;
  }

  async getProjectDoc(name) {
    const filePath = path.join(this._dir, 'docs', `${name}.md`);
    try {
      return fs.readFileSync(filePath, 'utf8');
    } catch {
      return null;
    }
  }

  async listProjectDocs() {
    const state = await this.readState();
    return state.project?.docs || [];
  }
```

**Step 5: Run tests**

```bash
cd <worktree> && npm test
```

Expected: All tests pass including new project profile tests. Some existing tests that check `state.version === 1` may need updating — change them to expect `2`.

**Step 6: Fix any existing tests that check version**

In `test/state-manager.test.js`, line 24: change `assert.equal(state.version, 1)` to `assert.equal(state.version, 2)`.

In `test/state-manager.test.js`, line 61: change the manually written state to `version: 2`.

In `test/state-manager.test.js`, line 441: change the manually written state to `version: 2` (in the "skips migration if state.json already exists" test, the state should be v2 since v1 would trigger auto-migration).

**Step 7: Run tests again**

```bash
cd <worktree> && npm test
```

Expected: All tests pass.

**Step 8: Commit**

```bash
git add plugin/servers/state-manager.js test/state-manager.test.js
git commit -m "feat: add project profile methods and v1→v2 state migration"
```

---

### Task 4: Wire 3 new MCP tools in tutoring-mcp.js

**Files:**
- Modify: `plugin/servers/tutoring-mcp.js`
- Modify: `test/mcp-tools.test.js`

**Step 1: Add import for ProjectScanner**

In `plugin/servers/tutoring-mcp.js`, after line 11 (the StateManager require), add:

```javascript
const { ProjectScanner } = require('./project-scanner');
```

**Step 2: Create scanner instance**

After line 155 (`const stateManager = ...`), add:

```javascript
const projectScanner = new ProjectScanner(process.cwd());
```

**Step 3: Add scan_project tool**

After the `get_learning_summary` tool (after line 335), add:

```javascript
// --- Project Analysis ---

server.tool('scan_project',
  'Scan the project structure, detect type, parse manifests, identify entry points. Fast — no source reading.',
  {},
  async () => {
    try {
      const profile = projectScanner.scan();
      await stateManager.saveProjectProfile(profile);
      return { content: [{ type: 'text', text: JSON.stringify(profile, null, 2) }] };
    } catch (err) {
      return { content: [{ type: 'text', text: `Error: ${err.message}` }], isError: true };
    }
  }
);

server.tool('get_project_profile',
  'Get the stored project profile and list of analysis docs',
  {},
  async () => {
    const profile = await stateManager.getProjectProfile();
    if (!profile) return { content: [{ type: 'text', text: 'No project profile. Run scan_project first.' }] };
    const docs = await stateManager.listProjectDocs();
    return { content: [{ type: 'text', text: JSON.stringify({ ...profile, availableDocs: docs }, null, 2) }] };
  }
);

server.tool('save_project_doc',
  'Save a project analysis document (used by sub-agents during onboarding)',
  {
    name: z.string().describe('Document name without extension (e.g. "architecture", "api-contracts")'),
    content: z.string().describe('Markdown content of the analysis document'),
  },
  async ({ name, content }) => {
    try {
      const filePath = await stateManager.saveProjectDoc(name, content);
      return { content: [{ type: 'text', text: `Saved to ${filePath}` }] };
    } catch (err) {
      return { content: [{ type: 'text', text: `Error: ${err.message}` }], isError: true };
    }
  }
);
```

**Step 4: Add MCP tool contract test**

Add to `test/mcp-tools.test.js`, after the existing `it` block (after line 67), inside the `MCP tool contracts` describe:

```javascript
  it('project workflow: scan → save doc → get profile', async () => {
    // Set up a fake project
    fs.writeFileSync(path.join(tmpDir, 'package.json'), JSON.stringify({
      name: 'test-app', dependencies: { express: '^4.18.0' },
    }));
    fs.writeFileSync(path.join(tmpDir, 'server.js'), '');
    fs.mkdirSync(path.join(tmpDir, 'src'));

    // Import ProjectScanner and verify it works with StateManager
    const { ProjectScanner } = require('../plugin/servers/project-scanner');
    const scanner = new ProjectScanner(tmpDir);
    const profile = scanner.scan();
    assert.ok(profile.projectType);
    assert.ok(profile.scannedAt);

    // Save profile via StateManager
    await sm.saveProjectProfile(profile);
    const loaded = await sm.getProjectProfile();
    assert.equal(loaded.projectType, profile.projectType);

    // Save a doc
    await sm.saveProjectDoc('architecture', '# Test Architecture');
    const doc = await sm.getProjectDoc('architecture');
    assert.equal(doc, '# Test Architecture');

    // List docs
    const docs = await sm.listProjectDocs();
    assert.ok(docs.includes('architecture'));
  });
```

**Step 5: Run tests**

```bash
cd <worktree> && npm test
```

Expected: All tests pass.

**Step 6: Commit**

```bash
git add plugin/servers/tutoring-mcp.js test/mcp-tools.test.js
git commit -m "feat: wire scan_project, get_project_profile, save_project_doc MCP tools"
```

---

### Task 5: Create /atu:onboard skill

**Files:**
- Create: `plugin/skills/atu-onboard/SKILL.md`

**Step 1: Create the skill directory and file**

```bash
mkdir -p plugin/skills/atu-onboard
```

**Step 2: Write the skill**

Create `plugin/skills/atu-onboard/SKILL.md`:

```markdown
---
name: atu:onboard
description: Analyze the current project — detect stack, architecture, patterns — and create a learning path
---

Analyze the student's project to help them understand the codebase. Uses a fast MCP scan followed by sub-agent deep analysis.

## Step 1: Quick Scan

1. Call `scan_project` to get the project skeleton (type, stack, structure, entry points)
2. Call `get_project_profile` to check if analysis docs already exist

**If docs already exist**, ask the student:
- "I've already analyzed this project. Would you like to: (a) rescan, (b) deep-dive into a specific area with /atu:deep-dive, or (c) see the summary?"
- If (a), proceed. If (b), suggest `/atu:deep-dive`. If (c), read and present `.agent-tutor/docs/index.md`.

## Step 2: Present Initial Findings

Show the student what you found:
- Project type and tech stack
- Directory structure overview
- Entry points
- Which analysis domains will be covered

Ask: "I'll now do a deep analysis of [N] areas: [domains]. This will take a minute. Ready?"

## Step 3: Spawn Analysis Sub-Agents

For each domain in the project's `analysisDomains`, spawn a sub-agent using the Agent tool. Run them **in parallel** where possible.

Each sub-agent gets this prompt template (fill in `{domain}`, `{projectType}`, `{skeleton}`):

```
You are analyzing a {projectType} project to help a student understand it.

**Your domain:** {domain}
**Project skeleton:** {skeleton}

Read the relevant source files in the project, then write a clear markdown analysis covering:

**For architecture:** Overall pattern (MVC, microservices, etc.), data flow diagram, key abstractions, design decisions
**For api:** All routes/endpoints, request/response shapes, middleware chain, auth flow
**For data:** Database schema, models, migrations, ORM patterns, validation rules
**For components:** UI component tree, state management, routing, styling approach
**For infra:** CI/CD pipeline, Docker setup, deployment config, environment management
**For testing:** Test patterns, frameworks, coverage approach, fixtures
**For dev-setup:** Prerequisites, install steps, build/run/test commands, environment setup

Write your findings as a complete markdown document. Focus on what a student needs to understand, not exhaustive documentation. Explain WHY things are structured this way, not just WHAT they are.

When done, call the `save_project_doc` MCP tool with name="{domain}" and your markdown content.
```

## Step 4: Collect Results

After all sub-agents complete:

1. Read each generated doc from `.agent-tutor/docs/`
2. Write an `index.md` via `save_project_doc` that links all docs with one-line summaries
3. Present a project overview to the student:
   - What the project does
   - Key architectural decisions
   - How the parts fit together
   - Where to start reading code

## Step 5: Gap Analysis and Learning Plan

1. Call `list_topics` to see what the student already knows
2. Compare against what the project uses (from the analysis)
3. Identify gaps — technologies/patterns the project uses that the student hasn't learned
4. Offer to create a learning plan: "Based on this project, you'd benefit from learning: [gaps]. Want me to create a plan with `/atu:plan`?"

ARGUMENTS: $ARGUMENTS
```

**Step 3: Commit**

```bash
git add plugin/skills/atu-onboard/SKILL.md
git commit -m "feat: add /atu:onboard skill for project analysis orchestration"
```

---

### Task 6: Create /atu:deep-dive skill

**Files:**
- Create: `plugin/skills/atu-deep-dive/SKILL.md`

**Step 1: Create the skill directory and file**

```bash
mkdir -p plugin/skills/atu-deep-dive
```

**Step 2: Write the skill**

Create `plugin/skills/atu-deep-dive/SKILL.md`:

```markdown
---
name: atu:deep-dive
description: Exhaustive analysis of a specific module or feature in the project
---

Deep-dive into a specific area of the project for focused learning.

## Step 1: Check Prerequisites

1. Call `get_project_profile` — if no profile exists, call `scan_project` first
2. Present the project's directory structure to the student

## Step 2: Pick Target

**If ARGUMENTS specify a path** (e.g., `/atu:deep-dive src/auth`):
- Use that path as the deep-dive target

**If no ARGUMENTS:**
- Show the top-level directories with brief descriptions
- Ask: "Which area would you like to understand deeply?"

## Step 3: Spawn Deep-Dive Sub-Agent

Spawn a single sub-agent with this prompt (fill in `{target}`, `{projectType}`):

```
You are doing an exhaustive analysis of the `{target}` directory in a {projectType} project to help a student understand it deeply.

Read EVERY file in `{target}` (and subdirectories). For each file, document:
- **Purpose:** What this file does (1-2 sentences)
- **Exports:** What it exposes to other files
- **Imports:** What it depends on
- **Key patterns:** Notable code patterns or techniques used

Then synthesize:
- **Dependency graph:** How files in this area relate to each other (which imports which)
- **Data flow:** How data moves through this module (input → processing → output)
- **Integration points:** How this module connects to the rest of the codebase
- **Key abstractions:** The main concepts/classes/functions a student must understand
- **Contributor guidance:** What someone modifying this code needs to know

Write everything as a clear markdown document. Explain patterns and decisions, not just list facts.

When done, call the `save_project_doc` MCP tool with name="deep-dive-{target-slug}" and your markdown content.
```

## Step 4: Present Findings

After the sub-agent completes:

1. Read the deep-dive doc from `.agent-tutor/docs/`
2. Walk through the findings pedagogically:
   - Start with the big picture: "This module does X"
   - Ask comprehension questions: "Looking at the dependency graph, why do you think Y depends on Z?"
   - Highlight non-obvious patterns: "Notice how they use X here — that's because..."
3. Update the project profile's `deepDives` array
4. Offer to create a topic for concepts discovered: "Want me to track your learning on [pattern/concept]?"

ARGUMENTS: $ARGUMENTS
```

**Step 3: Commit**

```bash
git add plugin/skills/atu-deep-dive/SKILL.md
git commit -m "feat: add /atu:deep-dive skill for focused module analysis"
```

---

### Task 7: Update tutor instructions with Project Awareness

**Files:**
- Modify: `plugin/templates/tutor-instructions.md`

**Step 1: Add Project Awareness section**

In `plugin/templates/tutor-instructions.md`, add a new section after "## Learning Plan Awareness" (after line 92, before "## Session Recovery"):

```markdown
## Project Awareness

Use `get_project_profile` to check if the project has been analyzed.

**When a project profile exists:**
- Reference the student's actual codebase when teaching: "In your project, `src/auth/middleware.js` uses this exact pattern."
- Read relevant docs from `.agent-tutor/docs/` when they connect to the current topic
- When the student asks about a concept, check if their project uses it and point to the specific file

**When no project profile exists:**
- On first interaction, call `scan_project` to get a basic profile (fast, no source reading)
- If the student seems to be exploring an unfamiliar codebase, suggest `/atu:onboard`
- If the student asks about a specific module, suggest `/atu:deep-dive <module>`
```

**Step 2: Update the Commands table**

In `plugin/templates/tutor-instructions.md`, add two rows to the Commands table (after the `/atu:plan` row):

```markdown
| `/atu:onboard` | Analyze the project — detect stack, architecture, patterns |
| `/atu:deep-dive` | Deep-dive into a specific module or feature |
```

**Step 3: Run tests**

```bash
cd <worktree> && npm test
```

Expected: All tests still pass (template changes don't affect code tests).

**Step 4: Commit**

```bash
git add plugin/templates/tutor-instructions.md
git commit -m "docs: add Project Awareness section and new commands to tutor instructions"
```

---

### Task 8: Update architecture docs, README, and CLAUDE.md

**Files:**
- Modify: `docs/architecture.md`
- Modify: `README.md`
- Modify: `CLAUDE.md`

**Step 1: Update docs/architecture.md**

Update the architecture diagram to include `project-scanner.js`. Update tool counts from 18 to 21. Add the 3 new tools to the tools table. Add `project-types.csv` to the package structure. Add a "Project Analysis" subsection under Components describing ProjectScanner and the sub-agent workflow. Update the data flow diagram to include the project analysis path.

Add a new section documenting the 14 supported project types.

**Step 2: Update README.md**

Update tool count from 18 to 21. Add `/atu:onboard` and `/atu:deep-dive` to the commands table. Add a "Project Analysis" section describing the onboarding feature.

**Step 3: Update CLAUDE.md**

Update the tool count reference from 18 to 21. Add `plugin/servers/project-scanner.js` to the project structure list.

**Step 4: Commit**

```bash
git add docs/architecture.md README.md CLAUDE.md
git commit -m "docs: update architecture and README for project onboarding feature"
```

---

### Summary

| Task | What | Files | Tests |
|------|------|-------|-------|
| 1 | project-types.csv | 1 new | — |
| 2 | ProjectScanner module | 2 new | ~15 tests |
| 3 | StateManager v2 + project methods | 1 modified, 1 modified | ~6 tests |
| 4 | 3 MCP tools | 1 modified, 1 modified | ~1 test |
| 5 | /atu:onboard skill | 1 new | — |
| 6 | /atu:deep-dive skill | 1 new | — |
| 7 | Tutor instructions update | 1 modified | — |
| 8 | Docs + README | 3 modified | — |

Total: 8 tasks, 5 new files, 6 modified files, ~22 new tests.
