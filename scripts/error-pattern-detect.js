#!/usr/bin/env node
// Agent-tutor hook: detects error patterns in Bash output — advisory only, never blocks.
'use strict';

const ERROR_PATTERNS = [
  /panic:/i,
  /\bFAIL\b/,
  /Traceback \(most recent call last\)/i,
  /\bError:/,
  /\bexception\b/i,
  /segfault/i,
  /fatal error/i,
  /command not found/i,
  /\bno such file or directory\b/i,
];

let raw = '';
process.stdin.on('data', chunk => { raw += chunk; });
process.stdin.on('end', () => {
  try {
    const input = JSON.parse(raw);
    if (input?.tool_name !== 'Bash') process.exit(0);

    const stdout = input?.tool_response?.stdout || '';
    const stderr = input?.tool_response?.stderr || '';
    const combined = stdout + '\n' + stderr;

    const hasError = ERROR_PATTERNS.some(p => p.test(combined));
    if (hasError) {
      const result = {
        continue: true,
        hookSpecificOutput: {
          hookEventName: 'PostToolUse',
          additionalContext: [
            'An error was detected in the terminal output.',
            'If coaching intensity is not "silent", consider guiding the student through the error.',
            'Use /atu:debug for a guided debugging session, or /atu:explain to explain the specific error.',
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
