#!/usr/bin/env node
// Agent-tutor hook: warns when a file exceeds 200 LOC — advisory only, never blocks.
'use strict';

let raw = '';
process.stdin.on('data', chunk => { raw += chunk; });
process.stdin.on('end', () => {
  try {
    const input = JSON.parse(raw);
    const toolName = input?.tool_name;
    if (toolName !== 'Write' && toolName !== 'Edit') {
      process.exit(0);
    }

    const filePath = input?.tool_input?.file_path || input?.tool_input?.path;
    if (!filePath) process.exit(0);

    const fs = require('fs');
    let content;
    try {
      content = fs.readFileSync(filePath, 'utf8');
    } catch {
      process.exit(0); // Unreadable — skip silently
    }

    const lines = content.split('\n').length;
    if (lines > 200) {
      const result = {
        continue: true,
        hookSpecificOutput: {
          hookEventName: 'PostToolUse',
          additionalContext: [
            `File ${filePath} is now ${lines} lines — above the 200-line threshold.`,
            'Consider coaching the student on extracting functions or splitting modules.',
            'This is a good moment for /atu:decompose to practice breaking up large files.',
          ],
        },
      };
      process.stdout.write(JSON.stringify(result) + '\n');
    }
  } catch {
    // Malformed input — skip silently
  }
  process.exit(0);
});
