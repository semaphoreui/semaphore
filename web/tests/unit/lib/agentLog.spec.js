import { expect } from 'chai';
import { reactive } from 'vue';
import AgentLog, {
  formatAgentDuration,
  formatAgentTokens,
  inlineSegments,
  inputSummary,
  parseAgentEvent,
  readable,
  readableResult,
  toolTitle,
} from '@/lib/agentLog';

let seq = 0;

function record(output) {
  seq += 1;
  return { id: `r${seq}`, time: `t${seq}`, output: typeof output === 'string' ? output : JSON.stringify(output) };
}

function fold(...outputs) {
  const rows = [];
  const log = new AgentLog(rows);
  outputs.forEach((output) => log.push(record(output)));
  return rows;
}

describe('lib/agentLog.js parseAgentEvent', () => {
  it('reads an event', () => {
    expect(parseAgentEvent('{"@agent":"text","text":"hi"}')).to.deep.equal({ '@agent': 'text', text: 'hi' });
  });

  it('takes nothing else for an event', () => {
    [
      'PLAY [all]',
      '',
      null,
      undefined,
      42,
      '{"type":"assistant"}',
      '{"text":"hi","@agent":"text"}',
      ' {"@agent":"text"}',
      '{"@agent":"text","text":',
      '{"@agent":"shout"}',
    ].forEach((line) => expect(parseAgentEvent(line), String(line)).to.equal(null));
  });
});

describe('lib/agentLog.js toolTitle', () => {
  it('splits the name of an MCP tool', () => {
    expect(toolTitle('mcp__github__list_issues')).to.deep.equal({ server: 'github', tool: 'list_issues' });
    expect(toolTitle('mcp__my_server__get__all')).to.deep.equal({ server: 'my_server', tool: 'get__all' });
  });

  it('keeps any other name', () => {
    expect(toolTitle('Bash')).to.deep.equal({ server: '', tool: 'Bash' });
    expect(toolTitle('mcp__broken')).to.deep.equal({ server: '', tool: 'mcp__broken' });
    expect(toolTitle('mcp____x')).to.deep.equal({ server: '', tool: 'mcp____x' });
    expect(toolTitle(undefined)).to.deep.equal({ server: '', tool: '' });
  });
});

describe('lib/agentLog.js readable', () => {
  it('writes an object a key per line, in its order', () => {
    expect(readable({
      decision: 'keep', relevance: 9, done: true, note: null,
    })).to.equal('decision: keep\nrelevance: 9\ndone: true\nnote: null');
  });

  it('writes a text of several lines as it is', () => {
    expect(readable({ name: 'body-markup', text: '# Writing a body\n\nA body is the deliverable.' }))
      .to.equal('name: body-markup\ntext:\n  # Writing a body\n  \n  A body is the deliverable.');
  });

  it('writes lists with dashes and nests objects', () => {
    expect(readable({
      result: [
        { code: 'SOURCE@1', status: 'pending' },
        { code: 'SOURCE@2', tags: ['a', 'b'], meta: {} },
      ],
      skipped: [],
    })).to.equal([
      'result:',
      '  - code: SOURCE@1',
      '    status: pending',
      '  - code: SOURCE@2',
      '    tags:',
      '      - a',
      '      - b',
      '    meta: {}',
      'skipped: []',
    ].join('\n'));
  });

  it('writes scalars, empty values and nothing', () => {
    expect(readable('two\nlines')).to.equal('two\nlines');
    expect(readable(5)).to.equal('5');
    expect(readable([])).to.equal('[]');
    expect(readable({})).to.equal('{}');
    expect(readable(undefined)).to.equal('');
  });

  it('stops nesting somewhere', () => {
    let deep = 'end';
    for (let i = 0; i < 40; i += 1) {
      deep = { k: deep };
    }
    expect(readable(deep)).to.include('{"k":');
  });
});

describe('lib/agentLog.js readableResult', () => {
  it('reads a result that is JSON', () => {
    expect(readableResult(' {"code":"SOURCE@1","status":"kept"}\n')).to.equal('code: SOURCE@1\nstatus: kept');
  });

  it('leaves text, and JSON cut short, as it is', () => {
    expect(readableResult('hello\nworld')).to.equal('hello\nworld');
    expect(readableResult('{"code":"SOURCE@1","body":"cut...')).to.equal('{"code":"SOURCE@1","body":"cut...');
    expect(readableResult('[image]')).to.equal('[image]');
    expect(readableResult(undefined)).to.equal('');
  });
});

describe('lib/agentLog.js inputSummary', () => {
  it('puts the arguments on one line', () => {
    expect(inputSummary({ source_code: 'SOURCE@1', relevance: 9, tags: ['a'] }))
      .to.equal('source_code: SOURCE@1, relevance: 9, tags: ["a"]');
    expect(inputSummary({ command: 'ls\n  -la' })).to.equal('command: ls -la');
    expect(inputSummary('plain')).to.equal('plain');
    expect(inputSummary(undefined)).to.equal('');
  });

  it('cuts a long line', () => {
    expect(inputSummary({ prompt: 'я'.repeat(50) }, 12)).to.equal(`prompt: ${'я'.repeat(4)}…`);
  });
});

describe('lib/agentLog.js inlineSegments', () => {
  it('marks bold and code', () => {
    expect(inlineSegments('Use **React** with `useChat`, not *this*.')).to.deep.equal([
      { text: 'Use ' },
      { text: 'React', bold: true },
      { text: ' with ' },
      { text: 'useChat', code: true },
      { text: ', not *this*.' },
    ]);
  });

  it('makes a heading bold', () => {
    expect(inlineSegments('## The **short** answer')).to.deep.equal([{ text: 'The short answer', bold: true }]);
  });

  it('keeps a plain and an empty line', () => {
    expect(inlineSegments('<b>not html</b>')).to.deep.equal([{ text: '<b>not html</b>' }]);
    expect(inlineSegments('')).to.deep.equal([{ text: '' }]);
    expect(inlineSegments(undefined)).to.deep.equal([{ text: '' }]);
  });
});

describe('lib/agentLog.js formatAgentDuration', () => {
  it('formats', () => {
    expect(formatAgentDuration(900)).to.equal('0.9s');
    expect(formatAgentDuration(19100)).to.equal('19.1s');
    expect(formatAgentDuration(95000)).to.equal('1m 35s');
    expect(formatAgentDuration(3 * 3600 * 1000 + 5 * 60 * 1000)).to.equal('3h 5m');
    expect(formatAgentDuration(undefined)).to.equal('0.0s');
  });
});

describe('lib/agentLog.js formatAgentTokens', () => {
  it('formats', () => {
    expect(formatAgentTokens(191)).to.equal('191');
    expect(formatAgentTokens(91709)).to.equal('91.7k');
    expect(formatAgentTokens(2400000)).to.equal('2.4M');
    expect(formatAgentTokens(undefined)).to.equal('0');
  });
});

describe('lib/agentLog.js AgentLog', () => {
  it('keeps a record that is not an event as its own row', () => {
    const rows = [];
    const log = new AgentLog(rows);
    const plain = record('PLAY [all]');
    const notOurs = record({ type: 'assistant' });
    log.push(plain);
    log.push(notOurs);
    expect(rows).to.deep.equal([plain, notOurs]);
    expect(rows[0]).to.equal(plain);
  });

  it('makes rows of the init, text and retry events', () => {
    const rows = fold(
      { '@agent': 'init', model: 'claude-opus-5-5', mcp: [{ name: 'github', status: 'connected' }, null, 'x'] },
      { '@agent': 'text', text: 'Working.' },
      { '@agent': 'text' },
      {
        '@agent': 'retry', attempt: 2, max_retries: 10, delay_ms: 1500, text: 'overloaded',
      },
    );

    expect(rows.map((row) => row.agent)).to.deep.equal([
      {
        kind: 'init', model: 'claude-opus-5-5', servers: [{ name: 'github', status: 'connected' }], sub: '',
      },
      { kind: 'text', text: 'Working.', sub: '' },
      { kind: 'text', text: '', sub: '' },
      {
        kind: 'retry', attempt: 2, maxRetries: 10, delayMs: 1500, text: 'overloaded', sub: '',
      },
    ]);
    expect(rows[1]).to.include.keys('id', 'time');
  });

  it('puts the result into the row of its call, wherever it comes', () => {
    const rows = fold(
      {
        '@agent': 'tool', id: 't1', name: 'mcp__research__source_get', input: { source_code: 'SOURCE@1' },
      },
      {
        '@agent': 'tool', id: 't2', name: 'Bash', input: { command: 'cat /nope' }, truncated: 3,
      },
      { '@agent': 'text', text: 'Reading.' },
      {
        '@agent': 'result', id: 't2', text: 'No such file', error: true,
      },
      {
        '@agent': 'result', id: 't1', text: '{"code":"SOURCE@1"}', truncated: 120,
      },
    );

    expect(rows).to.have.lengthOf(3);
    expect(rows[0].agent).to.deep.equal({
      kind: 'tool',
      toolId: 't1',
      server: 'research',
      tool: 'source_get',
      subagent: '',
      input: { source_code: 'SOURCE@1' },
      summary: 'source_code: SOURCE@1',
      truncated: 0,
      state: 'done',
      result: '{"code":"SOURCE@1"}',
      resultTruncated: 120,
      open: false,
      sub: '',
    });
    expect(rows[1].agent).to.include({
      tool: 'Bash', state: 'failed', result: 'No such file', truncated: 3, open: true,
    });
    expect(rows[2].agent.kind).to.equal('text');
  });

  it('replaces the arguments of a call that comes again: a plan is one row', () => {
    const rows = fold(
      {
        '@agent': 'tool', id: 'p1', name: 'todo_list', input: { progress: '0/2', steps: ['[ ] a', '[ ] b'] },
      },
      { '@agent': 'thinking', text: 'Reading the test.' },
      {
        '@agent': 'tool',
        id: 'p1',
        name: 'todo_list',
        input: { progress: '1/2', steps: ['[x] a', '[ ] b'] },
        truncated: 4,
      },
      { '@agent': 'result', id: 'p1' },
      { '@agent': 'tool', name: 'no id' },
      { '@agent': 'tool', name: 'no id' },
    );

    expect(rows.map((row) => row.agent.kind)).to.deep.equal(['tool', 'thinking', 'tool', 'tool']);
    expect(rows[0].agent).to.deep.include({
      tool: 'todo_list',
      input: { progress: '1/2', steps: ['[x] a', '[ ] b'] },
      summary: 'progress: 1/2, steps: ["[x] a","[ ] b"]',
      truncated: 4,
      state: 'done',
    });
    expect(rows[1].agent).to.deep.equal({ kind: 'thinking', text: 'Reading the test.', sub: '' });
  });

  it('leaves a call without a result running, and shows a result without a call', () => {
    const rows = fold(
      { '@agent': 'tool', id: 't1', name: 'Bash' },
      {
        '@agent': 'result', id: 'gone', text: 'late', truncated: 2,
      },
    );

    expect(rows[0].agent).to.include({ state: 'running', result: '', summary: '' });
    expect(rows[1].agent).to.deep.equal({
      kind: 'result', text: 'late', failed: false, truncated: 2, sub: '',
    });
  });

  it('names the rows of a subagent after the call that started it', () => {
    const rows = fold(
      {
        '@agent': 'tool', id: 'a1', name: 'Agent', input: { description: 'Research area: AI UI', prompt: 'Long prompt' },
      },
      {
        '@agent': 'tool', id: 'a2', name: 'Task', input: { subagent_type: 'Explore' },
      },
      {
        '@agent': 'tool', id: 't1', name: 'Read', input: { file_path: 'a.txt' }, agent: 'a1',
      },
      { '@agent': 'text', text: 'Found it.', agent: 'a1' },
      { '@agent': 'text', text: 'Also here.', agent: 'a2' },
      { '@agent': 'text', text: 'Whose?', agent: 'unknown' },
      {
        '@agent': 'result', id: 't1', text: 'data', agent: 'a1',
      },
      {
        '@agent': 'tool', id: 'm1', name: 'mcp__srv__Agent', input: { description: 'not a subagent' },
      },
    );

    expect(rows[0].agent).to.include({ tool: 'Agent', subagent: 'Research area: AI UI', summary: 'Research area: AI UI' });
    expect(rows[1].agent).to.include({ subagent: 'Explore' });
    expect(rows[2].agent).to.include({ sub: 'Research area: AI UI', state: 'done', result: 'data' });
    expect(rows[3].agent).to.include({ sub: 'Research area: AI UI', text: 'Found it.' });
    expect(rows[4].agent.sub).to.equal('Explore');
    expect(rows[5].agent.sub).to.equal('subagent');
    expect(rows[6].agent).to.include({ server: 'srv', subagent: '', summary: 'description: not a subagent' });
  });

  it('keeps one row for the ends of the turns, the last, with the totals', () => {
    const rows = fold(
      { '@agent': 'text', text: 'First turn.' },
      {
        '@agent': 'done', turns: 8, duration_ms: 19100, cost_usd: 1.5,
      },
      { '@agent': 'text', text: 'Second turn.' },
      {
        '@agent': 'done', turns: 1, duration_ms: 4700, cost_usd: 2.1784, error: true, errors: ['boom', 5],
      },
      {
        '@agent': 'done',
        turns: 2,
        duration_ms: 15500,
        cost_usd: 2.1784,
        input_tokens: 100,
        output_tokens: 7,
      },
    );

    expect(rows.map((row) => row.agent.kind)).to.deep.equal(['text', 'text', 'done']);
    expect(rows[2].agent).to.deep.equal({
      kind: 'done',
      failed: false,
      subtype: '',
      errors: ['boom', '5'],
      turns: 11,
      durationMs: 39300,
      costUsd: 2.1784,
      inputTokens: 100,
      outputTokens: 7,
      sub: '',
    });
  });

  it('marks a turn that ended with an error', () => {
    const rows = fold({
      '@agent': 'done', subtype: 'error_max_turns', turns: 5, duration_ms: 2000,
    });
    expect(rows[0].agent).to.include({ failed: true, subtype: 'error_max_turns' });
  });

  it('takes values of a wrong type for nothing', () => {
    const rows = fold({
      '@agent': 'tool', id: { a: 1 }, name: ['x'], truncated: 'many', agent: 7,
    }, { '@agent': 'text', text: { html: '<b>' } });

    expect(rows[0].agent).to.include({
      toolId: '', tool: '', truncated: 0, sub: 'subagent',
    });
    expect(rows[1].agent.text).to.equal('');
  });

  it('changes the rows of a reactive array through it', () => {
    const rows = reactive([]);
    const log = new AgentLog(rows);
    log.push(record({ '@agent': 'tool', id: 't1', name: 'Bash' }));
    const row = rows[0];
    log.push(record({ '@agent': 'result', id: 't1', text: 'ok' }));
    log.push(record({ '@agent': 'done', turns: 1 }));
    log.push(record({ '@agent': 'done', turns: 2 }));

    expect(row.agent.state).to.equal('done');
    expect(rows).to.have.lengthOf(2);
    expect(rows[1].agent.turns).to.equal(3);
  });
});
