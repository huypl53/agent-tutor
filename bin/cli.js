#!/usr/bin/env node
'use strict';

const { program } = require('commander');
const fs = require('fs');
const path = require('path');

const BEGIN = '<!-- BEGIN AGENT-TUTOR -->';
const END = '<!-- END AGENT-TUTOR -->';

function getInstructionsContent() {
  const mdPath = path.join(__dirname, '..', 'plugin', 'templates', 'tutor-instructions.md');
  return fs.readFileSync(mdPath, 'utf8');
}

function resolveClaudeMdPath(scope) {
  if (scope === 'global') {
    return path.join(require('os').homedir(), '.claude', 'CLAUDE.md');
  }
  return path.join(process.cwd(), '.claude', 'CLAUDE.md');
}

function resolveAgentsMdPath(scope) {
  if (scope === 'global') {
    return path.join(require('os').homedir(), 'AGENTS.md');
  }
  return path.join(process.cwd(), 'AGENTS.md');
}

function injectSection(filePath, section) {
  fs.mkdirSync(path.dirname(filePath), { recursive: true });
  let content = '';
  try { content = fs.readFileSync(filePath, 'utf8'); } catch {}

  // Remove existing section (idempotent)
  if (content.includes(BEGIN)) {
    content = removeSection(content);
  }

  if (content && !content.endsWith('\n')) content += '\n';
  if (content) content += '\n';
  content += section + '\n';

  fs.writeFileSync(filePath, content);
}

function removeSection(content) {
  const beginIdx = content.indexOf(BEGIN);
  const endIdx = content.indexOf(END);
  if (beginIdx < 0 || endIdx < 0) return content;

  let before = content.slice(0, beginIdx).trimEnd();
  let after = content.slice(endIdx + END.length).trimStart();

  if (!before) return after;
  if (!after) return before + '\n';
  return before + '\n\n' + after;
}

function getPluginDir() {
  return path.join(__dirname, '..', 'plugin');
}

program
  .name('agent-tutor')
  .version(require('../package.json').version)
  .description('Programming tutor plugin for coding agents');

program
  .command('install')
  .description('Install agent-tutor plugin')
  .option('--scope <scope>', 'local or global', 'local')
  .option('--agent <agent>', 'claude or codex', 'claude')
  .action((opts) => {
    const section = getInstructionsContent();

    if (opts.agent === 'codex') {
      const agentsPath = resolveAgentsMdPath(opts.scope);
      injectSection(agentsPath, section);
      console.log(`Injected tutor instructions into ${agentsPath}`);
      console.log(`\nAdd MCP server:\n  codex mcp add agent-tutor -- node ${path.join(getPluginDir(), 'servers', 'tutoring-mcp.js')}`);
    } else {
      const claudePath = resolveClaudeMdPath(opts.scope);
      injectSection(claudePath, section);
      console.log(`Injected tutor instructions into ${claudePath}`);
      console.log(`\nUse as plugin:\n  claude --plugin-dir ${getPluginDir()}`);
    }

    console.log('\nAvailable commands: /atu:check, /atu:hint, /atu:explain, /atu:debug, /atu:review, /atu:plan');
  });

program
  .command('uninstall')
  .description('Remove agent-tutor plugin')
  .option('--scope <scope>', 'local or global', 'local')
  .option('--agent <agent>', 'claude or codex', 'claude')
  .action((opts) => {
    const filePath = opts.agent === 'codex'
      ? resolveAgentsMdPath(opts.scope)
      : resolveClaudeMdPath(opts.scope);

    try {
      const content = fs.readFileSync(filePath, 'utf8');
      fs.writeFileSync(filePath, removeSection(content));
      console.log(`Removed tutor instructions from ${filePath}`);
    } catch {
      console.log('Nothing to uninstall.');
    }
  });

program
  .command('plugin-dir')
  .description('Print the plugin directory path (for --plugin-dir)')
  .action(() => {
    console.log(getPluginDir());
  });

program.parse();
