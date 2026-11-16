import { expect } from 'chai';
import {
  workflowRunFromQuery,
  workflowRunLocation,
  queryWithoutWorkflowRun,
} from '@/lib/workflowRunQuery';

describe('lib/workflowRunQuery.js', () => {
  describe('workflowRunFromQuery', () => {
    [
      ['a workflow and its run', { w: '4/10' }, { workflowId: 4, runId: 10 }],
      ['a run next to a task', { w: '4/10', t: '4844' }, { workflowId: 4, runId: 10 }],
      ['a repeated parameter', { w: ['4/10', '5/11'] }, { workflowId: 4, runId: 10 }],
      ['no parameter', { t: '4844' }, null],
      ['no query', undefined, null],
      ['an empty value', { w: '' }, null],
      ['a run without a workflow', { w: '10' }, null],
      ['a number', { w: 10 }, null],
      ['a workflow without a run', { w: '4/' }, null],
      ['a run without a workflow id', { w: '/10' }, null],
      ['a zero workflow', { w: '0/10' }, null],
      ['a zero run', { w: '4/0' }, null],
      ['a negative number', { w: '4/-3' }, null],
      ['text after the numbers', { w: '4/10abc' }, null],
      ['a third part', { w: '4/10/2' }, null],
      ['a fraction', { w: '4/1.5' }, null],
    ].forEach(([name, query, expected]) => {
      it(`reads ${name}`, () => {
        expect(workflowRunFromQuery(query)).to.deep.equal(expected);
      });
    });
  });

  describe('workflowRunLocation', () => {
    it('stays on the page and keeps the rest of the query', () => {
      const route = { path: '/project/1/workflows/4/runs', query: { t: '4844' } };

      expect(workflowRunLocation(route, 4, 10)).to.deep.equal({
        path: '/project/1/workflows/4/runs',
        query: { t: '4844', w: '4/10' },
      });
    });

    it('replaces the run the panel is open on', () => {
      const route = { path: '/project/1/workflows', query: { w: '4/10' } };

      expect(workflowRunLocation(route, 5, 11).query).to.deep.equal({ w: '5/11' });
    });

    it('is read back as the same run', () => {
      const { query } = workflowRunLocation({ path: '/', query: {} }, 4, 10);

      expect(workflowRunFromQuery(query)).to.deep.equal({ workflowId: 4, runId: 10 });
    });
  });

  describe('queryWithoutWorkflowRun', () => {
    it('drops the run and nothing else', () => {
      expect(queryWithoutWorkflowRun({ w: '4/10', t: '4844' })).to.deep.equal({
        w: undefined,
        t: '4844',
      });
    });
  });
});
