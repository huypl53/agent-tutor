'use strict';

const fs = require('fs');
const path = require('path');

const VALID_TRANSITIONS = {
  introduced: ['practicing'],
  practicing: ['struggling', 'breakthrough', 'mastered'],
  struggling: ['practicing', 'breakthrough'],
  breakthrough: ['mastered', 'practicing'],
  mastered: [],
};

const TOPIC_STATUSES = Object.keys(VALID_TRANSITIONS);

const EMPTY_STATE = {
  version: 1,
  topics: {},
  plan: null,
  session: null,
};

class StateManager {
  constructor(projectDir) {
    this._dir = path.join(projectDir, '.agent-tutor');
    this._file = path.join(this._dir, 'state.json');
  }

  async readState() {
    try {
      const raw = fs.readFileSync(this._file, 'utf8');
      return JSON.parse(raw);
    } catch {
      return JSON.parse(JSON.stringify(EMPTY_STATE));
    }
  }

  validateTransition(from, to) {
    const allowed = VALID_TRANSITIONS[from];
    if (!allowed || !allowed.includes(to)) {
      throw new Error(`Invalid transition: ${from} → ${to}`);
    }
  }

  async writeState(state) {
    fs.mkdirSync(this._dir, { recursive: true });
    const tmp = this._file + '.tmp';
    fs.writeFileSync(tmp, JSON.stringify(state, null, 2) + '\n');
    fs.renameSync(tmp, this._file);
  }

  async createTopic({ id, title, complexity = null, dependencies = [] }) {
    const state = await this.readState();
    if (state.topics[id]) {
      throw new Error(`Topic "${id}" already exists`);
    }
    state.topics[id] = {
      id,
      title,
      status: 'introduced',
      complexity,
      dependencies,
      subtopics: [],
      started: new Date().toISOString(),
      moments: [],
      lessonFile: null,
    };
    await this.writeState(state);
    return state.topics[id];
  }

  async updateTopic(id, updates) {
    const state = await this.readState();
    const topic = state.topics[id];
    if (!topic) {
      throw new Error(`Topic "${id}" not found`);
    }
    if (updates.status) {
      this.validateTransition(topic.status, updates.status);
      topic.status = updates.status;
    }
    if (updates.moment) {
      topic.moments.push({ ...updates.moment, ts: new Date().toISOString() });
    }
    if (updates.complexity !== undefined) {
      topic.complexity = updates.complexity;
    }
    if (updates.lessonFile !== undefined) {
      topic.lessonFile = updates.lessonFile;
    }
    await this.writeState(state);
    return topic;
  }

  async getTopic(id) {
    const state = await this.readState();
    return state.topics[id] || null;
  }

  async listTopics(status = null) {
    const state = await this.readState();
    let topics = Object.values(state.topics);
    if (status) {
      topics = topics.filter(t => t.status === status);
    }
    return topics;
  }

  async getTopicGraph() {
    const state = await this.readState();
    const topics = Object.values(state.topics);
    const nodes = topics.map(t => ({
      id: t.id,
      title: t.title,
      status: t.status,
      complexity: t.complexity,
    }));
    const edges = [];
    for (const t of topics) {
      for (const dep of t.dependencies) {
        edges.push({ from: dep, to: t.id });
      }
    }
    return { nodes, edges };
  }

  async createPlan({ goal, steps }) {
    const state = await this.readState();
    state.plan = {
      goal,
      steps: steps.map(s => ({ topicId: s.topicId, order: s.order, status: 'pending' })),
      progress: { completed: 0, total: steps.length },
    };
    await this.writeState(state);
    return state.plan;
  }

  async updatePlan(stepUpdates) {
    const state = await this.readState();
    if (!state.plan) {
      throw new Error('No plan exists. Create one first.');
    }
    for (const update of stepUpdates) {
      if (update.action === 'add') {
        state.plan.steps.push({ topicId: update.topicId, order: update.order, status: 'pending' });
      } else {
        const step = state.plan.steps.find(s => s.topicId === update.topicId);
        if (step && update.status) {
          step.status = update.status;
        }
      }
    }
    const completed = state.plan.steps.filter(s => s.status === 'mastered' || s.status === 'skipped').length;
    state.plan.progress = { completed, total: state.plan.steps.length };
    await this.writeState(state);
    return state.plan;
  }

  async getPlan() {
    const state = await this.readState();
    return state.plan;
  }

  async saveSession({ activeTopicId, resumeContext }) {
    const state = await this.readState();
    state.session = {
      activeTopicId,
      resumeContext,
      lastActivity: new Date().toISOString(),
    };
    await this.writeState(state);
    return state.session;
  }

  async restoreSession() {
    const state = await this.readState();
    return state.session;
  }

  async getLearningSummary() {
    const state = await this.readState();
    const topics = Object.values(state.topics);

    const topicsByStatus = {};
    for (const s of TOPIC_STATUSES) topicsByStatus[s] = 0;
    for (const t of topics) topicsByStatus[t.status]++;

    const allMoments = topics
      .flatMap(t => t.moments.map(m => ({ topicId: t.id, ...m })))
      .sort((a, b) => b.ts.localeCompare(a.ts))
      .slice(0, 10);

    return {
      totalTopics: topics.length,
      topicsByStatus,
      plan: state.plan ? { goal: state.plan.goal, progress: state.plan.progress } : null,
      recentMoments: allMoments,
      activeSession: state.session,
    };
  }
}

module.exports = { StateManager, EMPTY_STATE, TOPIC_STATUSES, VALID_TRANSITIONS };
