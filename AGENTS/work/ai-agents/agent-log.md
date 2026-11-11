# The agent log: events in the task log, a transcript in the client

Status: implemented 2026-11-12 on `feat/agent_app` for Claude Code (TASK@C202422088,
TASK@9BF9DF3D50) and for Codex (TASK@0D2CFDA994).

## What it is

An agent run writes its progress to the task log as **events, one JSON object per log line**, in a
schema Semaphore owns (`pkg/agentlog`). The web client folds the events into the rows of a
transcript; everything that shows the log as text prints them as plain lines.

```
claude -p --output-format=stream-json        the CLI's own events (JSONL, large, CLI-specific)
  → db_lib/AgentOutput.go claudeStreamOutput  decodes, masks, line-buffers the text, cuts
codex exec --json                            the CLI's own events (JSONL)
  → db_lib/AgentOutputCodex.go codexJSONOutput  decodes, masks, cuts, times the turn
      → task log line  {"@agent":"tool","id":…}   pkg/agentlog.Event, one per line
          → web: lib/agentLog.js AgentLog         rows; TaskLogAgentRecord.vue renders a row
          → raw_output, `semaphore client task logs`   agentlog.Lines() → plain text
```

A line is an event when it starts with `{"@agent":"` and decodes to a known kind. Any other line
is a plain log line, as before — Semaphore's own messages, the output of hooks, a CLI error, the
CLI's own log on stderr (Codex reports a failed MCP connection there, as `ERROR rmcp::…`).

## Events

| Kind | Fields | Meaning |
| --- | --- | --- |
| `init` | `model`, `mcp[{name,status}]` | Start of a run. Claude repeats it every turn; logged only when it changed. Codex has none. |
| `text` | `text` | One line of the agent's text (Claude: the streamed pieces buffered to whole lines; Codex: a whole message at once). |
| `thinking` | `text` | One line of the agent's reasoning, when the CLI shows it (Codex `reasoning` items). |
| `tool` | `id`, `name`, `input`, `truncated` | A tool call; `input` is the JSON of its arguments, keys in their order. Logged again with the same `id` it replaces the arguments (a plan ticked off). |
| `result` | `id`, `text`, `error`, `truncated` | The result of the call `id`. JSON results stay valid JSON. |
| `retry` | `attempt`, `max_retries`, `delay_ms`, `text` | An API request is retried (Claude). |
| `done` | `error`, `subtype`, `errors`, `turns`, `duration_ms`, `cost_usd`, `input_tokens`, `output_tokens` | End of a turn, with what the CLI tells of it. |

Every event may carry `agent`: the id of the tool call that started the subagent it belongs to.

## How Codex items map

| Codex (`item.*`) | Event |
| --- | --- |
| `agent_message` | `text` lines, at `item.completed` |
| `reasoning` | `thinking` lines |
| `command_execution` | `tool` `exec` `{command}` at the first report, `result` at completion (`Exit code N` first when not 0) |
| `mcp_tool_call` | `tool` `mcp__<server>__<tool>` — Claude's spelling, the client splits server and tool by it; `result` from `error.message`, the text of `result.content`, or `structured_content` |
| `file_change` | `tool` `file_change` `{changes:[{path,kind}]}` + `result` |
| `web_search` | `tool` `web_search` `{query}` + empty `result` |
| `todo_list` | `tool` `todo_list` `{progress:"1/3", steps:["[x] …"]}` at every report, `result` at completion |
| `collab_tool_call` | `tool` named by its action (`spawn_agent`…) + `result` |
| `error` item, `error` event | a plain line `Error: …` |
| `turn.completed` / `turn.failed` | `done` with the tokens; the duration is measured by the reader (the CLI does not tell it); a failed turn does not repeat the error already logged |

`thread.started`, `turn.started` and unknown types log nothing.

## Decisions and their reasons

- **Events, not rendered text, in the log.** The first version (TASK@C202422088) rendered text on
  the server (`> tool {json}`, indented results). The owner's real run showed why a client can not
  work from that: parallel calls and background subagents interleave, so a result does not follow
  its call and only an id pairs them; a text line of the agent starting with `> ` is not a call;
  JSON cut at N characters is not JSON.
- **Our schema, not the CLI's.** The CLI's stream is large (tool lists, signatures, a delta per
  token), version-specific, and would put the knowledge of one CLI into the client. Codex was
  added on the server alone, plus three general things in the schema and the client (`thinking`,
  a repeated `tool`, tokens).
- **Masking before encoding.** A secret is replaced in the decoded strings (text, every string
  and key of the arguments, every string of a JSON result), then the event is encoded: a secret
  with a character JSON escapes would not be found in the encoded line, and the client decodes
  the line. The executors' own masking of the finished line stays on top.
- **Cut strings, keep JSON.** Limits (`pkg/agentlog`): 4000 characters per string of the
  arguments, 2000 per string of a JSON result, 8000 for a text result; arguments over 32 KB or a
  result over 16 KB get their strings cut to 200. `truncated` counts what was left out. The JSON
  is rewritten token by token, so key order and number spelling survive.
- **Nothing from the agent is HTML.** The client shows every value as text (`{{ }}`); `**bold**`
  and `` `code` `` are split into segments and rendered as elements, not parsed as markup. No
  markdown library, no `v-html` for agent rows. Plain log lines keep the ANSI → HTML path.
- **Folding happens as records arrive** (`AgentLog.push`), not as a computed over the whole log:
  the log view is a virtual list fed once a second, and a log can hold 10⁵ lines.
- **One `done` row.** An agent with background subagents ends several turns; the client keeps the
  last `done`, with turns, durations and tokens summed (the cost Claude reports is already a
  total). A part the CLI does not tell (turns and cost for Codex) is not shown.
- **Codex MCP tools are approved in the config** (`default_tools_approval_mode = "approve"`).
  ⚠️ 2026-11-12: with `"auto"`, which the config had, `codex exec` (codex-cli 0.167.0) refused
  every MCP tool call — "requires approval, but approval policy is never" — because exec can not
  ask. Found only with a real MCP server; the stand's fake CLI calls nothing.

## Open

- `GET /tasks/{id}/output` returns the event lines as they are stored; an API consumer that wants
  text uses `raw_output`. Not yet in `api-docs.yml` or the docs site (no agent page exists there).
- Subagent text of Claude is logged per whole message, not streamed, so two subagents do not mix
  lines. Codex text is not streamed at all: `--json` gives a message only when it is complete.
- Codex gives no model name and no MCP server status in its events, so its log has no `init` row.
- Check the log view after changing it: `AGENTS/tools/agent-stand/shoot-log.cjs` (both CLIs).
