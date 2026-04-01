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
}

module.exports = { StateManager, EMPTY_STATE, TOPIC_STATUSES, VALID_TRANSITIONS };
