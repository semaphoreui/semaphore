# The agent log: events in the task log, a transcript in the client

Status: implemented 2026-11-12 on `feat/agent_app` (TASK@C202422088, TASK@9BF9DF3D50), for Claude
Code. Codex still logs plain text.

## What it is

A Claude Code run writes its progress to the task log as **events, one JSON object per log line**,
in a schema Semaphore owns (`pkg/agentlog`). The web client folds the events into the rows of a
transcript; everything that shows the log as text prints them as plain lines.

```
claude -p --output-format=stream-json        the CLI's own events (JSONL, large, CLI-specific)
  → db_lib/AgentOutput.go claudeStreamOutput  decodes, masks, line-buffers the text, cuts
  → task log line  {"@agent":"tool","id":…}   pkg/agentlog.Event, one per line
      → web: lib/agentLog.js AgentLog         rows; TaskLogAgentRecord.vue renders a row
      → raw_output, `semaphore client task logs`   agentlog.Lines() → plain text
```

A line is an event when it starts with `{"@agent":"` and decodes to a known kind. Any other line
is a plain log line, as before — Semaphore's own messages, the output of hooks, a CLI error.

## Events

| Kind | Fields | Meaning |
| --- | --- | --- |
| `init` | `model`, `mcp[{name,status}]` | Start of a run. The CLI repeats it every turn; logged only when it changed. |
| `text` | `text` | One line of the agent's text (the server buffers the streamed pieces to whole lines). |
| `tool` | `id`, `name`, `input`, `truncated` | A tool call; `input` is the JSON of its arguments, keys in their order. |
| `result` | `id`, `text`, `error`, `truncated` | The result of the call `id`. JSON results stay valid JSON. |
| `retry` | `attempt`, `max_retries`, `delay_ms`, `text` | An API request is retried. |
| `done` | `error`, `subtype`, `errors`, `turns`, `duration_ms`, `cost_usd` | End of a turn. |

Every event may carry `agent`: the id of the tool call that started the subagent it belongs to.

## Decisions and their reasons

- **Events, not rendered text, in the log.** The first version (TASK@C202422088) rendered text on
  the server (`> tool {json}`, indented results). The owner's real run showed why a client can not
  work from that: parallel calls and background subagents interleave, so a result does not follow
  its call and only an id pairs them; a text line of the agent starting with `> ` is not a call;
  JSON cut at N characters is not JSON.
- **Our schema, not the CLI's.** The CLI's stream is large (tool lists, signatures, a delta per
  token), version-specific, and would put the knowledge of one CLI into the client. A second CLI
  maps to the same events on the server.
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
  last `done`, with turns and durations summed (the cost the CLI reports is already a total).

## Open

- Codex: `codex exec --json` could be mapped to the same events (`item.completed` with
  `agent_message` / `command_execution`); until then its log is plain text.
- `GET /tasks/{id}/output` returns the event lines as they are stored; an API consumer that wants
  text uses `raw_output`. Not yet in `api-docs.yml` or the docs site (no agent page exists there).
- Subagent text is logged per whole message, not streamed, so two subagents do not mix lines.
- Check the log view after changing it: `AGENTS/tools/agent-stand/shoot-log.cjs`.
