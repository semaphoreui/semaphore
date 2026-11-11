import './setup';
import { expect } from 'chai';
import { mount } from '@vue/test-utils';
import TaskLogAgentRecord from '@/components/TaskLogAgentRecord.vue';
import TaskLogViewRecord from '@/components/TaskLogViewRecord.vue';
import AgentLog from '@/lib/agentLog';

// The rows AgentLog makes of the events.
function rowsOf(...events) {
  const rows = [];
  const log = new AgentLog(rows);
  events.forEach((event, i) => log.push({
    id: `r${i}`, time: '2026-11-12T10:00:00Z', output: JSON.stringify(event),
  }));
  return rows;
}

function mountAgent(agent, props = {}) {
  return mount(TaskLogAgentRecord, {
    props: { agent, ...props },
    global: {
      mocks: { $t: (key, params) => (params ? `${key} ${JSON.stringify(params)}` : key) },
    },
  });
}

const find = (wrapper, id) => wrapper.find(`[data-testid="${id}"]`);

describe('TaskLogAgentRecord.vue', () => {
  it('shows a line of the agent with bold and code, as text', () => {
    const [row] = rowsOf({
      '@agent': 'text', text: 'Use **React** and `useChat` <img src=x onerror=alert(1)>',
    });
    const wrapper = mountAgent(row.agent);
    const text = find(wrapper, 'taskLog-agentText');

    expect(text.text()).to.equal('Use React and useChat <img src=x onerror=alert(1)>');
    expect(text.find('b').text()).to.equal('React');
    expect(text.find('code').text()).to.equal('useChat');
    expect(text.find('img').exists(), 'the text of the agent is not HTML').to.equal(false);
  });

  it('shows a tool call closed: its server, name and arguments on one line', () => {
    const [row] = rowsOf({
      '@agent': 'tool',
      id: 't1',
      name: 'mcp__research__source_review',
      input: { decision: 'keep', relevance: 9 },
    });
    const wrapper = mountAgent(row.agent);
    const head = find(wrapper, 'taskLog-agentTool');

    expect(head.text().replace(/\s+/g, ' ')).to.equal('research source_review decision: keep, relevance: 9');
    expect(head.attributes('aria-expanded')).to.equal('false');
    expect(find(wrapper, 'taskLog-agentToolState').attributes('data-state')).to.equal('running');
    expect(find(wrapper, 'taskLog-agentToolDetails').exists()).to.equal(false);
  });

  it('asks to open a tool call on a click', async () => {
    const [row] = rowsOf({ '@agent': 'tool', id: 't1', name: 'Bash' });
    const wrapper = mountAgent(row.agent);

    await find(wrapper, 'taskLog-agentTool').trigger('click');
    expect(wrapper.emitted('toggle')).to.have.lengthOf(1);
  });

  it('shows the arguments and the result of an open call in a readable form', () => {
    const [row] = rowsOf(
      {
        '@agent': 'tool',
        id: 't1',
        name: 'mcp__research__skill_get',
        input: { skill_name: 'body-markup', sections: ['markdown'] },
        truncated: 12,
      },
      {
        '@agent': 'result',
        id: 't1',
        text: '{"name":"body-markup","text":"# Writing a body\\n\\n<script>alert(1)</script>"}',
        truncated: 340,
      },
    );
    row.agent.open = true;
    const wrapper = mountAgent(row.agent);

    expect(find(wrapper, 'taskLog-agentTool').attributes('aria-expanded')).to.equal('true');
    expect(find(wrapper, 'taskLog-agentToolState').attributes('data-state')).to.equal('done');
    expect(find(wrapper, 'taskLog-agentToolInput').text())
      .to.equal('skill_name: body-markup\nsections:\n  - markdown');

    const result = find(wrapper, 'taskLog-agentToolResult');
    expect(result.text()).to.equal('name: body-markup\ntext:\n  # Writing a body\n  \n  <script>alert(1)</script>');
    expect(result.find('script').exists(), 'the result of a tool is not HTML').to.equal(false);

    const details = find(wrapper, 'taskLog-agentToolDetails').text();
    expect(details).to.include('agentLogTruncated {"n":12}');
    expect(details).to.include('agentLogTruncated {"n":340}');
  });

  it('opens a failed call by itself and marks its result', () => {
    const [row] = rowsOf(
      {
        '@agent': 'tool', id: 't1', name: 'Bash', input: { command: 'cat /nope' },
      },
      {
        '@agent': 'result', id: 't1', text: 'No such file', error: true,
      },
    );
    const wrapper = mountAgent(row.agent);

    expect(find(wrapper, 'taskLog-agentToolState').attributes('data-state')).to.equal('failed');
    const result = find(wrapper, 'taskLog-agentToolResult');
    expect(result.text()).to.equal('No such file');
    expect(result.classes()).to.include('task-log-agent__value--failed');
  });

  it('tells a call without a result apart in a running and in an ended task', () => {
    const [row] = rowsOf({ '@agent': 'tool', id: 't1', name: 'Bash' });
    row.agent.open = true;

    const running = mountAgent(row.agent, { running: true });
    expect(find(running, 'taskLog-agentToolState').attributes('data-state')).to.equal('running');
    expect(find(running, 'taskLog-agentToolNoResult').text()).to.equal('agentLogTool_running');
    expect(find(running, 'taskLog-agentToolInput').exists(), 'a call without arguments').to.equal(false);

    const ended = mountAgent(row.agent, { running: false });
    expect(find(ended, 'taskLog-agentToolState').attributes('data-state')).to.equal('interrupted');
    expect(find(ended, 'taskLog-agentToolNoResult').text()).to.equal('agentLogTool_interrupted');
  });

  it('names a subagent in the call that starts it and on its rows', () => {
    const rows = rowsOf(
      {
        '@agent': 'tool', id: 'a1', name: 'Agent', input: { description: 'Research area: AI UI', prompt: 'p' },
      },
      { '@agent': 'text', text: 'Found it.', agent: 'a1' },
    );

    const call = mountAgent(rows[0].agent);
    expect(find(call, 'taskLog-agentTool').text().replace(/\s+/g, ' '))
      .to.equal('agentLogSubagent Research area: AI UI');
    expect(find(call, 'taskLog-agentSub').exists()).to.equal(false);

    const line = mountAgent(rows[1].agent);
    expect(find(line, 'taskLog-agentSub').text()).to.equal('Research area: AI UI');
    expect(line.classes()).to.include('task-log-agent--sub');
  });

  it('shows the model and the MCP servers, a server that is down with its status', () => {
    const [row] = rowsOf({
      '@agent': 'init',
      model: 'claude-opus-5-5',
      mcp: [{ name: 'github', status: 'connected' }, { name: 'docs', status: 'failed' }],
    });
    const wrapper = mountAgent(row.agent);

    expect(find(wrapper, 'taskLog-agentInit').text()).to.include('agentLogModel claude-opus-5-5');
    const servers = wrapper.findAll('[data-testid="taskLog-agentServer"]');
    expect(servers.map((s) => s.text())).to.deep.equal(['github', 'docs: failed']);
    expect(servers[0].classes()).to.not.include('task-log-agent__server--down');
    expect(servers[1].classes()).to.include('task-log-agent__server--down');
  });

  it('shows a retry and the end of the run', () => {
    const rows = rowsOf(
      {
        '@agent': 'retry', attempt: 2, max_retries: 10, text: 'overloaded',
      },
      {
        '@agent': 'done', turns: 11, duration_ms: 95000, cost_usd: 2.1784,
      },
    );

    expect(find(mountAgent(rows[0].agent), 'taskLog-agentRetry').text())
      .to.equal('agentLogRetry {"attempt":2,"max":10}: overloaded');

    const done = find(mountAgent(rows[1].agent), 'taskLog-agentDone');
    expect(done.text().replace(/\s+/g, ' '))
      .to.equal('agentLogFinished · agentLogTurns {"n":11} · 1m 35s · $2.18');
    expect(done.classes()).to.not.include('task-log-agent__done--failed');
  });

  it('shows a failed end with its errors', () => {
    const [row] = rowsOf({
      '@agent': 'done', error: true, subtype: 'error_max_turns', errors: ['Reached the turn limit'], turns: 5,
    });
    const done = find(mountAgent(row.agent), 'taskLog-agentDone');

    expect(done.text().replace(/\s+/g, ' '))
      .to.equal('Reached the turn limit agentLogFailed (error_max_turns) · agentLogTurns {"n":5} · 0.0s');
    expect(done.classes()).to.include('task-log-agent__done--failed');
  });

  it('shows the result of a call the log does not hold', () => {
    const [row] = rowsOf({ '@agent': 'result', id: 'gone', text: '{"ok":true}' });
    expect(find(mountAgent(row.agent), 'taskLog-agentResult').text()).to.equal('ok: true');
  });
});

describe('TaskLogViewRecord.vue', () => {
  const global = { mocks: { $t: (key) => key } };

  it('shows a record of any other task as before: ANSI colours as HTML', () => {
    const wrapper = mount(TaskLogViewRecord, {
      props: { source: { id: 'r1', time: '2026-11-12T10:00:00Z', output: '\u001b[32mok\u001b[0m: [web1]' } },
      global,
    });

    const output = find(wrapper, 'taskLog-recordOutput');
    expect(output.text()).to.equal('ok: [web1]');
    expect(output.find('span').exists()).to.equal(true);
    expect(find(wrapper, 'taskLog-agentRecord').exists()).to.equal(false);
  });

  it('shows a row of an agent with the time, and passes the toggle on', async () => {
    const [row] = rowsOf({ '@agent': 'tool', id: 't1', name: 'Bash' });
    const wrapper = mount(TaskLogViewRecord, { props: { source: row, running: false }, global });

    expect(find(wrapper, 'taskLog-recordTime').exists()).to.equal(true);
    expect(find(wrapper, 'taskLog-recordOutput').exists()).to.equal(false);
    expect(find(wrapper, 'taskLog-agentToolState').attributes('data-state')).to.equal('interrupted');

    await find(wrapper, 'taskLog-agentTool').trigger('click');
    expect(wrapper.emitted('toggle')).to.have.lengthOf(1);
  });
});
