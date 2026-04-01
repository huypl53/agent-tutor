const { describe, it, beforeEach, afterEach } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('fs');
const path = require('path');
const os = require('os');
const { StateManager } = require('../plugin/servers/state-manager');

// We test StateManager directly since MCP tool handlers are thin shells.
// This verifies the contract that tools will fulfill.
describe('MCP tool contracts', () => {
  let tmpDir, sm;

  beforeEach(() => {
    tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'atu-mcp-'));
    sm = new StateManager(tmpDir);
  });

  afterEach(() => {
    fs.rmSync(tmpDir, { recursive: true, force: true });
  });

  it('full workflow: create topics → plan → session → summary', async () => {
    // Create topics with dependencies
    await sm.createTopic({ id: 'callbacks', title: 'Callbacks' });
    await sm.createTopic({ id: 'promises', title: 'Promises', dependencies: ['callbacks'], complexity: 6 });
    await sm.createTopic({ id: 'async', title: 'Async/Await', dependencies: ['promises'], complexity: 8 });

    // Progress through a topic
    await sm.updateTopic('callbacks', { status: 'practicing' });
    await sm.updateTopic('callbacks', { moment: { type: 'practice', detail: 'wrote first callback' } });
    await sm.updateTopic('callbacks', { status: 'mastered' });

    // Create plan
    await sm.createPlan({
      goal: 'Master async JS',
      steps: [
        { topicId: 'callbacks', order: 1 },
        { topicId: 'promises', order: 2 },
        { topicId: 'async', order: 3 },
      ],
    });
    await sm.updatePlan([{ topicId: 'callbacks', status: 'mastered' }]);

    // Save session
    await sm.saveSession({ activeTopicId: 'promises', resumeContext: 'Starting promises chapter' });

    // Get graph
    const graph = await sm.getTopicGraph();
    assert.equal(graph.nodes.length, 3);
    assert.equal(graph.edges.length, 2);
    assert.equal(graph.nodes.find(n => n.id === 'callbacks').status, 'mastered');

    // Get summary
    const summary = await sm.getLearningSummary();
    assert.equal(summary.totalTopics, 3);
    assert.equal(summary.topicsByStatus.mastered, 1);
    assert.equal(summary.topicsByStatus.introduced, 2);
    assert.equal(summary.plan.progress.completed, 1);
    assert.equal(summary.plan.progress.total, 3);
    assert.equal(summary.recentMoments.length, 1);
    assert.equal(summary.activeSession.activeTopicId, 'promises');

    // Restore session
    const session = await sm.restoreSession();
    assert.equal(session.activeTopicId, 'promises');
    assert.equal(session.resumeContext, 'Starting promises chapter');
  });
});
