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

  describe('topics CRUD', () => {
    it('createTopic adds a new topic', async () => {
      await sm.createTopic({ id: 'promises', title: 'JavaScript Promises' });
      const state = await sm.readState();
      assert.equal(state.topics['promises'].title, 'JavaScript Promises');
      assert.equal(state.topics['promises'].status, 'introduced');
      assert.ok(state.topics['promises'].started);
    });

    it('createTopic with optional fields', async () => {
      await sm.createTopic({
        id: 'async',
        title: 'Async/Await',
        complexity: 7,
        dependencies: ['promises'],
      });
      const state = await sm.readState();
      assert.equal(state.topics['async'].complexity, 7);
      assert.deepEqual(state.topics['async'].dependencies, ['promises']);
    });

    it('createTopic rejects duplicate id', async () => {
      await sm.createTopic({ id: 'x', title: 'X' });
      await assert.rejects(
        () => sm.createTopic({ id: 'x', title: 'X again' }),
        /already exists/
      );
    });

    it('updateTopic changes status with validation', async () => {
      await sm.createTopic({ id: 't', title: 'T' });
      await sm.updateTopic('t', { status: 'practicing' });
      const state = await sm.readState();
      assert.equal(state.topics['t'].status, 'practicing');
    });

    it('updateTopic rejects invalid transition', async () => {
      await sm.createTopic({ id: 't', title: 'T' });
      await assert.rejects(
        () => sm.updateTopic('t', { status: 'mastered' }),
        /Invalid transition/
      );
    });

    it('updateTopic adds a moment', async () => {
      await sm.createTopic({ id: 't', title: 'T' });
      await sm.updateTopic('t', {
        moment: { type: 'struggle', detail: 'confused about closures' },
      });
      const state = await sm.readState();
      assert.equal(state.topics['t'].moments.length, 1);
      assert.equal(state.topics['t'].moments[0].type, 'struggle');
      assert.ok(state.topics['t'].moments[0].ts);
    });

    it('updateTopic sets complexity', async () => {
      await sm.createTopic({ id: 't', title: 'T' });
      await sm.updateTopic('t', { complexity: 5 });
      const state = await sm.readState();
      assert.equal(state.topics['t'].complexity, 5);
    });

    it('updateTopic sets lessonFile', async () => {
      await sm.createTopic({ id: 't', title: 'T' });
      await sm.updateTopic('t', { lessonFile: 'lessons/2026-04-01-t.md' });
      const state = await sm.readState();
      assert.equal(state.topics['t'].lessonFile, 'lessons/2026-04-01-t.md');
    });

    it('updateTopic rejects unknown topic', async () => {
      await assert.rejects(
        () => sm.updateTopic('nope', { status: 'practicing' }),
        /not found/
      );
    });

    it('getTopic returns a topic', async () => {
      await sm.createTopic({ id: 't', title: 'T' });
      const topic = await sm.getTopic('t');
      assert.equal(topic.title, 'T');
    });

    it('getTopic returns null for unknown', async () => {
      const topic = await sm.getTopic('nope');
      assert.equal(topic, null);
    });

    it('listTopics returns all topics', async () => {
      await sm.createTopic({ id: 'a', title: 'A' });
      await sm.createTopic({ id: 'b', title: 'B' });
      const list = await sm.listTopics();
      assert.equal(list.length, 2);
    });

    it('listTopics filters by status', async () => {
      await sm.createTopic({ id: 'a', title: 'A' });
      await sm.createTopic({ id: 'b', title: 'B' });
      await sm.updateTopic('a', { status: 'practicing' });
      const list = await sm.listTopics('practicing');
      assert.equal(list.length, 1);
      assert.equal(list[0].id, 'a');
    });
  });

  describe('getTopicGraph', () => {
    it('returns topics with dependency edges', async () => {
      await sm.createTopic({ id: 'callbacks', title: 'Callbacks' });
      await sm.createTopic({ id: 'promises', title: 'Promises', dependencies: ['callbacks'] });
      await sm.createTopic({ id: 'async', title: 'Async/Await', dependencies: ['promises'] });

      const graph = await sm.getTopicGraph();
      assert.equal(graph.nodes.length, 3);
      assert.equal(graph.edges.length, 2);
      assert.deepEqual(graph.edges[0], { from: 'callbacks', to: 'promises' });
      assert.deepEqual(graph.edges[1], { from: 'promises', to: 'async' });
    });

    it('returns empty graph when no topics', async () => {
      const graph = await sm.getTopicGraph();
      assert.equal(graph.nodes.length, 0);
      assert.equal(graph.edges.length, 0);
    });
  });
});
