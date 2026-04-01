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
}

module.exports = { StateManager, EMPTY_STATE, TOPIC_STATUSES, VALID_TRANSITIONS };
