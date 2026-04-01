'use strict';

const fs = require('fs');
const path = require('path');

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

  async writeState(state) {
    fs.mkdirSync(this._dir, { recursive: true });
    const tmp = this._file + '.tmp';
    fs.writeFileSync(tmp, JSON.stringify(state, null, 2) + '\n');
    fs.renameSync(tmp, this._file);
  }
}

module.exports = { StateManager, EMPTY_STATE };
