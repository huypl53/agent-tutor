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
      assert.ok(webApp.key_files.includes('vite.config.*'));
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
