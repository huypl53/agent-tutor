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

    it('throws on corrupt JSON instead of returning empty state', async () => {
      const atuDir = path.join(tmpDir, '.agent-tutor');
      fs.mkdirSync(atuDir, { recursive: true });
      fs.writeFileSync(path.join(atuDir, 'state.json'), '{corrupt json!!!');
      await assert.rejects(() => sm.readState(), /Unexpected token|Expected/);
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

  describe('plan CRUD', () => {
    it('createPlan stores a learning plan', async () => {
      await sm.createPlan({
        goal: 'Learn async JS',
        steps: [
          { topicId: 'callbacks', order: 1 },
          { topicId: 'promises', order: 2 },
        ],
      });
      const state = await sm.readState();
      assert.equal(state.plan.goal, 'Learn async JS');
      assert.equal(state.plan.steps.length, 2);
      assert.equal(state.plan.steps[0].status, 'pending');
      assert.deepEqual(state.plan.progress, { completed: 0, total: 2 });
    });

    it('createPlan overwrites existing plan', async () => {
      await sm.createPlan({ goal: 'Old', steps: [{ topicId: 'a', order: 1 }] });
      await sm.createPlan({ goal: 'New', steps: [{ topicId: 'b', order: 1 }] });
      const state = await sm.readState();
      assert.equal(state.plan.goal, 'New');
    });

    it('updatePlan marks steps completed and updates progress', async () => {
      await sm.createPlan({
        goal: 'Learn',
        steps: [
          { topicId: 'a', order: 1 },
          { topicId: 'b', order: 2 },
        ],
      });
      await sm.updatePlan([{ topicId: 'a', status: 'mastered' }]);
      const state = await sm.readState();
      assert.equal(state.plan.steps[0].status, 'mastered');
      assert.deepEqual(state.plan.progress, { completed: 1, total: 2 });
    });

    it('updatePlan adds new steps', async () => {
      await sm.createPlan({ goal: 'Learn', steps: [{ topicId: 'a', order: 1 }] });
      await sm.updatePlan([{ topicId: 'c', order: 2, action: 'add' }]);
      const state = await sm.readState();
      assert.equal(state.plan.steps.length, 2);
      assert.deepEqual(state.plan.progress, { completed: 0, total: 2 });
    });

    it('updatePlan rejects when no plan exists', async () => {
      await assert.rejects(
        () => sm.updatePlan([{ topicId: 'a', status: 'mastered' }]),
        /No plan exists/
      );
    });

    it('getPlan returns plan or null', async () => {
      assert.equal(await sm.getPlan(), null);
      await sm.createPlan({ goal: 'G', steps: [{ topicId: 'x', order: 1 }] });
      const plan = await sm.getPlan();
      assert.equal(plan.goal, 'G');
    });
  });

  describe('session', () => {
    it('saveSession stores context', async () => {
      await sm.saveSession({ activeTopicId: 'promises', resumeContext: 'Working on error handling' });
      const state = await sm.readState();
      assert.equal(state.session.activeTopicId, 'promises');
      assert.equal(state.session.resumeContext, 'Working on error handling');
      assert.ok(state.session.lastActivity);
    });

    it('restoreSession returns session or null', async () => {
      assert.equal(await sm.restoreSession(), null);
      await sm.saveSession({ activeTopicId: 'x', resumeContext: 'test' });
      const session = await sm.restoreSession();
      assert.equal(session.activeTopicId, 'x');
    });
  });

  describe('getLearningSummary', () => {
    it('returns aggregate summary', async () => {
      await sm.createTopic({ id: 'a', title: 'A' });
      await sm.createTopic({ id: 'b', title: 'B' });
      await sm.updateTopic('a', { status: 'practicing' });
      await sm.updateTopic('a', { moment: { type: 'struggle', detail: 'hard' } });
      await sm.createPlan({ goal: 'G', steps: [{ topicId: 'a', order: 1 }, { topicId: 'b', order: 2 }] });

      const summary = await sm.getLearningSummary();
      assert.equal(summary.topicsByStatus.introduced, 1);
      assert.equal(summary.topicsByStatus.practicing, 1);
      assert.equal(summary.totalTopics, 2);
      assert.equal(summary.plan.goal, 'G');
      assert.equal(summary.plan.progress.completed, 0);
      assert.equal(summary.plan.progress.total, 2);
      assert.equal(summary.recentMoments.length, 1);
    });

    it('returns summary with no data', async () => {
      const summary = await sm.getLearningSummary();
      assert.equal(summary.totalTopics, 0);
      assert.equal(summary.plan, null);
      assert.equal(summary.recentMoments.length, 0);
    });
  });

  describe('migration', () => {
    it('migrates current-topic.md on first load', async () => {
      const atuDir = path.join(tmpDir, '.agent-tutor');
      fs.mkdirSync(atuDir, { recursive: true });
      fs.writeFileSync(path.join(atuDir, 'current-topic.md'), [
        '# Current Topic',
        '',
        '**Topic:** JavaScript Promises',
        '**Started:** 2026-03-30T10:00:00Z',
        '',
        '## Moments',
        '- struggle: confused about .then chaining',
        '- hint: showed diagram of promise states',
      ].join('\n'));

      await sm.migrateIfNeeded();
      const state = await sm.readState();

      assert.ok(state.topics['javascript-promises']);
      assert.equal(state.topics['javascript-promises'].title, 'JavaScript Promises');
      assert.equal(state.topics['javascript-promises'].status, 'introduced');
      assert.equal(state.topics['javascript-promises'].moments.length, 2);
      assert.ok(fs.existsSync(path.join(atuDir, 'current-topic.md.bak')));
      assert.equal(fs.existsSync(path.join(atuDir, 'current-topic.md')), false);
    });

    it('migrates learning-plan.md on first load', async () => {
      const atuDir = path.join(tmpDir, '.agent-tutor');
      fs.mkdirSync(atuDir, { recursive: true });
      fs.writeFileSync(path.join(atuDir, 'learning-plan.md'), [
        '# Learning Plan: Master Async JS',
        '',
        '**Progress:** 1/3 complete',
        '',
        '## Steps',
        '- [x] 1. **Callbacks** — understand callback patterns',
        '- [ ] 2. **Promises** — learn promise API',
        '- [ ] 3. **Async/Await** — modern async syntax',
      ].join('\n'));

      await sm.migrateIfNeeded();
      const state = await sm.readState();

      assert.equal(state.plan.goal, 'Master Async JS');
      assert.equal(state.plan.steps.length, 3);
      assert.equal(state.plan.steps[0].status, 'mastered');
      assert.equal(state.plan.steps[1].status, 'pending');
      assert.deepEqual(state.plan.progress, { completed: 1, total: 3 });
      assert.ok(fs.existsSync(path.join(atuDir, 'learning-plan.md.bak')));
    });

    it('skips migration if state.json already exists', async () => {
      const atuDir = path.join(tmpDir, '.agent-tutor');
      fs.mkdirSync(atuDir, { recursive: true });
      fs.writeFileSync(path.join(atuDir, 'state.json'), JSON.stringify({ version: 1, topics: {}, plan: null, session: null }));
      fs.writeFileSync(path.join(atuDir, 'current-topic.md'), '# Current Topic\n**Topic:** X\n**Started:** 2026-01-01T00:00:00Z');

      await sm.migrateIfNeeded();
      const state = await sm.readState();
      assert.deepEqual(state.topics, {});
      // original file NOT renamed — migration was skipped
      assert.ok(fs.existsSync(path.join(atuDir, 'current-topic.md')));
    });

    it('handles missing markdown files gracefully', async () => {
      await sm.migrateIfNeeded();
      const state = await sm.readState();
      assert.deepEqual(state.topics, {});
    });
  });
});
