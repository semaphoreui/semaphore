import { expect } from 'chai';
import {
  isWorkflowEditorPath,
  navMiniForPath,
  readSideCollapsed,
  writeSideCollapsed,
} from '@/lib/workflowEditorPrefs';

// The unit environment has no Web Storage; a Map-backed stand-in is enough.
function fakeStorage() {
  const map = new Map();
  return {
    getItem: (k) => (map.has(k) ? map.get(k) : null),
    setItem: (k, v) => map.set(k, String(v)),
    removeItem: (k) => map.delete(k),
    clear: () => map.clear(),
  };
}

describe('workflowEditorPrefs', () => {
  let savedStorage;

  beforeEach(() => {
    savedStorage = global.localStorage;
    global.localStorage = fakeStorage();
  });

  afterEach(() => {
    global.localStorage = savedStorage;
  });

  it('round-trips the collapsed flag through localStorage', () => {
    expect(readSideCollapsed()).to.equal(false);
    writeSideCollapsed(true);
    expect(readSideCollapsed()).to.equal(true);
    writeSideCollapsed(false);
    expect(readSideCollapsed()).to.equal(false);
    expect(localStorage.getItem('workflowEditor__sideCollapsed')).to.equal(null);
  });

  describe('isWorkflowEditorPath', () => {
    [
      ['/project/1/workflows/new', true],
      ['/project/12/workflows/34/edit', true],
      ['/project/12/workflows/34/edit/', true],
      ['/project/1/workflows', false],
      ['/project/1/workflows/34', false],
      ['/project/1/workflows/34/runs', false],
      ['/project/1/workflows/34/runs/5', false],
      ['/project/1/templates', false],
      ['', false],
      [undefined, false],
    ].forEach(([path, expected]) => {
      it(`${JSON.stringify(path)} -> ${expected}`, () => {
        expect(isWorkflowEditorPath(path)).to.equal(expected);
      });
    });
  });

  it('asks for a mini navigation only on the editor with the flag set', () => {
    expect(navMiniForPath('/project/1/workflows/2/edit')).to.equal(false);
    writeSideCollapsed(true);
    expect(navMiniForPath('/project/1/workflows/2/edit')).to.equal(true);
    expect(navMiniForPath('/project/1/workflows')).to.equal(false);
  });
});
