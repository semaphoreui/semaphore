# ai-agents — zone index

AI agent templates (Claude Code, Codex): how an agent run reports its progress to the task log,
how the web client shows it, and what the prompt can name in `${{ … }}`.

| When you need it | Read |
| --- | --- |
| Adding a field or a kind to the agent's events in the task log, or changing their limits and masking | `agent-log.md` |
| Changing how the log view shows an agent run (tool calls, subagents, the summary) | `agent-log.md` |
| Adding an agent CLI, or an item of Codex / an event of Claude Code the log does not show yet | `agent-log.md` |
| A Codex task whose MCP tool calls are refused or need an approval | `agent-log.md` |
| Something reads the task log as text (raw log, CLI client, a new consumer) and meets JSON lines | `agent-log.md` |
| Adding a context or a source of values to the `${{ … }}` expressions of the prompt, or changing what a missing or malformed one does | `prompt-expressions.md` |
| Deciding what a prompt given by a task may use, or where a secret put into the prompt is masked | `prompt-expressions.md` |
