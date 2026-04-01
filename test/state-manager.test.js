const { describe, it, beforeEach, afterEach } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('fs');
const path = require('path');
const os = require('os');
const { StateManager } = require('../plugin/servers/state-manager');

describe('StateManager', () => {
  let tmpDir;
  let sm;

  beforeEach(() => {
    tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'atu-test-'));
    sm = new StateManager(tmpDir);
  });

  afterEach(() => {
    fs.rmSync(tmpDir, { recursive: true, force: true });
  });

  describe('readState / writeState', () => {
    it('returns empty state when no file exists', async () => {
      const state = await sm.readState();
      assert.equal(state.version, 1);
      assert.deepEqual(state.topics, {});
      assert.equal(state.plan, null);
      assert.equal(state.session, null);
    });

    it('persists and reads back state', async () => {
      const state = await sm.readState();
      state.topics['test-topic'] = {
        id: 'test-topic',
        title: 'Test Topic',
        status: 'introduced',
        complexity: null,
        dependencies: [],
        subtopics: [],
        started: new Date().toISOString(),
        moments: [],
        lessonFile: null,
      };
      await sm.writeState(state);

      const loaded = await sm.readState();
      assert.equal(loaded.topics['test-topic'].title, 'Test Topic');
    });

    it('creates .agent-tutor directory if missing', async () => {
      await sm.readState();
      // readState should not create the dir (lazy)
      assert.equal(fs.existsSync(path.join(tmpDir, '.agent-tutor')), false);

      // writeState creates it
      await sm.writeState({ version: 1, topics: {}, plan: null, session: null });
      assert.ok(fs.existsSync(path.join(tmpDir, '.agent-tutor', 'state.json')));
    });
  });

  describe('topic status transitions', () => {
    it('allows introduced → practicing', () => {
      assert.doesNotThrow(() => sm.validateTransition('introduced', 'practicing'));
    });

    it('allows practicing → struggling', () => {
      assert.doesNotThrow(() => sm.validateTransition('practicing', 'struggling'));
    });

    it('allows practicing → breakthrough', () => {
      assert.doesNotThrow(() => sm.validateTransition('practicing', 'breakthrough'));
    });

    it('allows practicing → mastered', () => {
      assert.doesNotThrow(() => sm.validateTransition('practicing', 'mastered'));
    });

    it('allows struggling → practicing', () => {
      assert.doesNotThrow(() => sm.validateTransition('struggling', 'practicing'));
    });

    it('allows struggling → breakthrough', () => {
      assert.doesNotThrow(() => sm.validateTransition('struggling', 'breakthrough'));
    });

    it('allows breakthrough → mastered', () => {
      assert.doesNotThrow(() => sm.validateTransition('breakthrough', 'mastered'));
    });

    it('allows breakthrough → practicing', () => {
      assert.doesNotThrow(() => sm.validateTransition('breakthrough', 'practicing'));
    });

    it('rejects introduced → mastered', () => {
      assert.throws(
        () => sm.validateTransition('introduced', 'mastered'),
        /Invalid transition/
      );
    });

    it('rejects mastered → anything', () => {
      assert.throws(
        () => sm.validateTransition('mastered', 'practicing'),
        /Invalid transition/
      );
    });

    it('rejects introduced → struggling', () => {
      assert.throws(
        () => sm.validateTransition('introduced', 'struggling'),
        /Invalid transition/
      );
    });
  });
});
