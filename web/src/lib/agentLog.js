// The agent log: the progress of an AI agent as the task log holds it — one event per log line,
// as JSON (pkg/agentlog on the server). This module makes the rows of the log view of the
// records: a tool call with its arguments and its result, the agent's text, the subagents.
//
// Nothing here makes HTML. The values of an event come from the agent and its tools; the log
// view shows them as text only.

const EVENT_PREFIX = '{"@agent":"';
const EVENT_KINDS = new Set(['init', 'text', 'thinking', 'tool', 'result', 'retry', 'done']);

// The tools which start a subagent; their rows name it.
const SUBAGENT_TOOLS = new Set(['Agent', 'Task']);

const SUMMARY_MAX = 160;
const READABLE_MAX_DEPTH = 12;

function str(value) {
  if (typeof value === 'string') {
    return value;
  }
  return typeof value === 'number' ? String(value) : '';
}

function num(value) {
  return typeof value === 'number' && Number.isFinite(value) ? value : 0;
}

// parseAgentEvent returns the event of a log line, null when the line is not one.
export function parseAgentEvent(line) {
  if (typeof line !== 'string' || !line.startsWith(EVENT_PREFIX)) {
    return null;
  }

  let event;
  try {
    event = JSON.parse(line);
  } catch {
    return null;
  }

  if (event == null || typeof event !== 'object' || !EVENT_KINDS.has(event['@agent'])) {
    return null;
  }
  return event;
}

// toolTitle splits the name of a tool into the MCP server it belongs to and its own name:
// mcp__github__list_issues is the tool list_issues of the server github.
export function toolTitle(name) {
  const full = str(name);
  if (full.startsWith('mcp__')) {
    const rest = full.slice('mcp__'.length);
    const i = rest.indexOf('__');
    if (i > 0 && i < rest.length - 2) {
      return { server: rest.slice(0, i), tool: rest.slice(i + 2) };
    }
  }
  return { server: '', tool: full };
}

function isScalar(value) {
  return value === null || typeof value !== 'object';
}

function scalarText(value) {
  if (value === null) {
    return 'null';
  }
  return typeof value === 'string' ? value : String(value);
}

function readableLines(value, indent, depth) {
  if (depth > READABLE_MAX_DEPTH) {
    return [`${indent}${JSON.stringify(value)}`];
  }

  if (isScalar(value)) {
    return scalarText(value).split('\n').map((line) => `${indent}${line}`);
  }

  const nested = `${indent}  `;

  if (Array.isArray(value)) {
    if (value.length === 0) {
      return [`${indent}[]`];
    }
    return value.flatMap((item) => {
      const lines = readableLines(item, nested, depth + 1);
      lines[0] = `${indent}- ${lines[0].slice(nested.length)}`;
      return lines;
    });
  }

  const keys = Object.keys(value);
  if (keys.length === 0) {
    return [`${indent}{}`];
  }
  return keys.flatMap((key) => {
    const item = value[key];
    if (isScalar(item)) {
      const text = scalarText(item);
      if (!text.includes('\n')) {
        return [`${indent}${key}: ${text}`];
      }
      return [`${indent}${key}:`, ...text.split('\n').map((line) => `${nested}${line}`)];
    }
    if ((Array.isArray(item) ? item : Object.keys(item)).length === 0) {
      return [`${indent}${key}: ${Array.isArray(item) ? '[]' : '{}'}`];
    }
    return [`${indent}${key}:`, ...readableLines(item, nested, depth + 1)];
  });
}

// readable writes a JSON value the way a person reads it: a key and its value per line, lists
// with dashes, a text of several lines as it is, not as one string with escapes.
export function readable(value) {
  return value === undefined ? '' : readableLines(value, '', 0).join('\n');
}

// readableResult returns the result of a tool call for reading: a result that is JSON as
// readable() writes it, any other as it is.
export function readableResult(text) {
  const trimmed = str(text).trim();
  if (trimmed.startsWith('{') || trimmed.startsWith('[')) {
    try {
      return readable(JSON.parse(trimmed));
    } catch {
      // Not JSON, or JSON cut short: it is text.
    }
  }
  return str(text);
}

function oneLine(value) {
  const text = typeof value === 'string' ? value : JSON.stringify(value);
  return str(text).replace(/\s+/g, ' ').trim();
}

// inputSummary puts the arguments of a tool call on one line.
export function inputSummary(input, max = SUMMARY_MAX) {
  if (input === undefined || input === null) {
    return '';
  }

  const summary = isScalar(input) || Array.isArray(input)
    ? oneLine(input)
    : Object.keys(input).map((key) => `${key}: ${oneLine(input[key])}`).join(', ');

  const chars = Array.from(summary);
  return chars.length > max ? `${chars.slice(0, max).join('')}…` : summary;
}

// inlineSegments splits a line of the agent's text into plain, **bold** and `code` pieces, so
// that the view can mark them without making HTML of the text. A heading is bold as a whole.
export function inlineSegments(text) {
  const line = str(text);

  const heading = /^#{1,6}\s+(.*)$/.exec(line);
  if (heading) {
    return [{ text: heading[1].replace(/\*\*/g, ''), bold: true }];
  }

  const segments = [];
  const pattern = /\*\*([^*\n]+)\*\*|`([^`\n]+)`/g;
  let last = 0;
  let match = pattern.exec(line);
  while (match !== null) {
    if (match.index > last) {
      segments.push({ text: line.slice(last, match.index) });
    }
    segments.push(match[1] !== undefined
      ? { text: match[1], bold: true }
      : { text: match[2], code: true });
    last = match.index + match[0].length;
    match = pattern.exec(line);
  }
  if (last < line.length || segments.length === 0) {
    segments.push({ text: line.slice(last) });
  }
  return segments;
}

// formatAgentTokens: 191 → 191, 91709 → 91.7k, 2400000 → 2.4M.
export function formatAgentTokens(count) {
  const n = num(count);
  if (n < 1000) {
    return String(Math.round(n));
  }
  return n < 1000000 ? `${(n / 1000).toFixed(1)}k` : `${(n / 1000000).toFixed(1)}M`;
}

// formatAgentDuration: 900 → 0.9s, 19100 → 19.1s, 95000 → 1m 35s.
export function formatAgentDuration(ms) {
  const seconds = num(ms) / 1000;
  if (seconds < 60) {
    return `${seconds.toFixed(1)}s`;
  }
  const whole = Math.round(seconds);
  const minutes = Math.floor(whole / 60);
  if (minutes < 60) {
    return `${minutes}m ${whole % 60}s`;
  }
  return `${Math.floor(minutes / 60)}h ${minutes % 60}m`;
}

// AgentLog folds the records of a task log into the rows of the log view, as they arrive. A
// record that is not an event of an agent is its own row. Of the events:
//
// - the result of a tool call goes into the row of the call, wherever it comes: calls run in
//   parallel and subagents work at the same time, so results do not follow their calls;
// - a call that comes again with the same id replaces the arguments of its row: a plan the
//   agent ticks off is one row;
// - a row of a subagent carries the name the agent gave it;
// - the ends of the agent's turns are one row, the last one, with the totals.
//
// `rows` is the array the view renders. Give it the reactive one: the rows a later event
// changes are read back from it.
export default class AgentLog {
  constructor(rows) {
    this.rows = rows;
    this.tools = new Map();
    this.subagents = new Map();
    this.done = null;
  }

  push(record) {
    const event = parseAgentEvent(record.output);
    if (event === null) {
      this.rows.push(record);
      return;
    }

    const sub = str(event.agent);

    switch (event['@agent']) {
      case 'init':
        this.add(record, sub, {
          kind: 'init',
          model: str(event.model),
          servers: (Array.isArray(event.mcp) ? event.mcp : [])
            .filter((server) => server != null && typeof server === 'object')
            .map((server) => ({ name: str(server.name), status: str(server.status) })),
        });
        break;

      case 'text':
      case 'thinking':
        this.add(record, sub, { kind: event['@agent'], text: str(event.text) });
        break;

      case 'tool':
        this.addTool(record, sub, event);
        break;

      case 'result':
        this.addResult(record, sub, event);
        break;

      case 'retry':
        this.add(record, sub, {
          kind: 'retry',
          attempt: num(event.attempt),
          maxRetries: num(event.max_retries),
          delayMs: num(event.delay_ms),
          text: str(event.text),
        });
        break;

      case 'done':
        this.addDone(record, event);
        break;

      default:
        break;
    }
  }

  // add appends a row and returns it as the rows hold it (reactive, when they are).
  add(record, sub, agent) {
    this.rows.push({
      id: record.id,
      time: record.time,
      agent: {
        ...agent,
        // The subagent the row belongs to: the name the agent gave it, when it is known.
        sub: sub === '' ? '' : (this.subagents.get(sub) || 'subagent'),
      },
    });
    return this.rows[this.rows.length - 1];
  }

  addTool(record, sub, event) {
    const { server, tool } = toolTitle(event.name);
    const id = str(event.id);
    const { input } = event;

    let subagent = '';
    if (SUBAGENT_TOOLS.has(tool) && server === '') {
      const named = input != null && typeof input === 'object' ? input : {};
      subagent = str(named.description) || str(named.subagent_type) || 'subagent';
      if (id !== '') {
        this.subagents.set(id, subagent);
      }
    }

    const summary = subagent === '' ? inputSummary(input) : subagent;
    const truncated = num(event.truncated);

    const known = id === '' ? undefined : this.tools.get(id);
    if (known !== undefined) {
      Object.assign(known.agent, { input, summary, truncated });
      return;
    }

    const row = this.add(record, sub, {
      kind: 'tool',
      toolId: id,
      server,
      tool,
      // The name of the subagent the call starts, empty for any other tool.
      subagent,
      input,
      summary,
      truncated,
      // running until the result comes, then done or failed.
      state: 'running',
      result: '',
      resultTruncated: 0,
      open: false,
    });

    if (id !== '') {
      this.tools.set(id, row);
    }
  }

  addResult(record, sub, event) {
    const failed = event.error === true;
    const call = this.tools.get(str(event.id));

    if (call === undefined) {
      // The result of a call the log does not hold.
      this.add(record, sub, {
        kind: 'result',
        text: str(event.text),
        failed,
        truncated: num(event.truncated),
      });
      return;
    }

    call.agent.state = failed ? 'failed' : 'done';
    call.agent.result = str(event.text);
    call.agent.resultTruncated = num(event.truncated);
    // A failure is what one opens a call for.
    if (failed) {
      call.agent.open = true;
    }
  }

  addDone(record, event) {
    const errors = Array.isArray(event.errors) ? event.errors.map(str).filter((e) => e !== '') : [];
    const done = {
      kind: 'done',
      failed: event.error === true || str(event.subtype) !== '',
      subtype: str(event.subtype),
      errors,
      turns: num(event.turns),
      durationMs: num(event.duration_ms),
      costUsd: num(event.cost_usd),
      inputTokens: num(event.input_tokens),
      outputTokens: num(event.output_tokens),
    };

    // An agent with background work ends several turns; the cost the CLI reports is already
    // the total.
    if (this.done !== null) {
      const previous = this.done.agent;
      done.turns += previous.turns;
      done.durationMs += previous.durationMs;
      done.inputTokens += previous.inputTokens;
      done.outputTokens += previous.outputTokens;
      done.costUsd = Math.max(done.costUsd, previous.costUsd);
      done.errors = [...previous.errors, ...done.errors];

      const at = this.rows.lastIndexOf(this.done);
      if (at >= 0) {
        this.rows.splice(at, 1);
      }
    }

    this.done = this.add(record, '', done);
  }
}
