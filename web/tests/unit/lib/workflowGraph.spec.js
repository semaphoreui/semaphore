import { expect } from 'chai';
import {
  wouldCreateCycle, nextNodeId, formatDuration, escapeHtml,
} from '@/lib/workflowGraph';

const edge = (from, to) => ({ source_node_id: from, destination_node_id: to });

describe('lib/workflowGraph', () => {
  describe('wouldCreateCycle', () => {
    const chain = [edge(1, 2), edge(2, 3)];

    const tests = [
      {
        name: 'forward edge in a chain', edges: chain, source: 3, dest: 4, expected: false,
      },
      {
        name: 'back edge to the root', edges: chain, source: 3, dest: 1, expected: true,
      },
      {
        name: 'back edge to the parent', edges: chain, source: 2, dest: 1, expected: true,
      },
      {
        name: 'self loop', edges: chain, source: 2, dest: 2, expected: true,
      },
      {
        name: 'parallel branch', edges: chain, source: 1, dest: 3, expected: false,
      },
      {
        name: 'empty graph', edges: [], source: 1, dest: 2, expected: false,
      },
      {
        name: 'missing edges', edges: undefined, source: 1, dest: 2, expected: false,
      },
      {
        name: 'diamond join',
        edges: [edge(1, 2), edge(1, 3), edge(2, 4), edge(3, 4)],
        source: 4,
        dest: 1,
        expected: true,
      },
      {
        name: 'existing cycle elsewhere',
        edges: [edge(5, 6), edge(6, 5)],
        source: 1,
        dest: 2,
        expected: false,
      },
    ];

    tests.forEach(({
      name, edges, source, dest, expected,
    }) => {
      it(`${name} -> ${expected}`, () => {
        expect(wouldCreateCycle(edges, source, dest)).to.equal(expected);
      });
    });
  });

  describe('nextNodeId', () => {
    const tests = [
      { ids: [], expected: 1 },
      { ids: undefined, expected: 1 },
      { ids: [1, 2, 3], expected: 4 },
      { ids: [7, 2], expected: 8 },
      { ids: [0], expected: 1 },
      { ids: ['5', null, undefined], expected: 6 },
    ];
    tests.forEach(({ ids, expected }) => {
      it(`${JSON.stringify(ids)} -> ${expected}`, () => {
        expect(nextNodeId(ids)).to.equal(expected);
      });
    });
  });

  describe('formatDuration', () => {
    const tests = [
      { ms: 0, expected: '0s' },
      { ms: 1, expected: '1s' },
      { ms: 999, expected: '1s' },
      { ms: 1000, expected: '1s' },
      { ms: 59000, expected: '59s' },
      { ms: 60000, expected: '1m 0s' },
      { ms: 61500, expected: '1m 2s' },
      { ms: 3600000, expected: '60m 0s' },
    ];
    tests.forEach(({ ms, expected }) => {
      it(`${ms}ms -> ${expected}`, () => {
        expect(formatDuration(ms)).to.equal(expected);
      });
    });
  });

  describe('escapeHtml', () => {
    const tests = [
      { value: '<script>alert(1)</script>', expected: '&lt;script&gt;alert(1)&lt;/script&gt;' },
      { value: 'a & b', expected: 'a &amp; b' },
      { value: '&lt;', expected: '&amp;lt;' },
      { value: 'plain', expected: 'plain' },
      { value: null, expected: '' },
      { value: undefined, expected: '' },
      { value: 42, expected: '42' },
    ];
    tests.forEach(({ value, expected }) => {
      it(`${JSON.stringify(value)} -> ${expected}`, () => {
        expect(escapeHtml(value)).to.equal(expected);
      });
    });
  });
});
