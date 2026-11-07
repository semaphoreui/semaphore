# agent-cli

Runs the real `codex` or `claude` CLI the way `db_lib/AgentApp.go` runs it, with an isolated home, an MCP config that reads the token from an env var, and the prompt on stdin. A local fake MCP server logs the `Authorization` header it receives. Use it to check that the CLI still accepts the generated config and that the API key works.

```
AGENTS/tools/agent-cli/probe.sh codex|claude [prompt]      # env: PROBE_PORT (default 8765)
```

```
$ AGENTS/tools/agent-cli/probe.sh codex
codex
OK
--- exit: 0
--- MCP requests:
POST /mcp auth=Bearer <token ok>
```

`<token ok>` means the CLI sent the expected token. `(none)` means it never reached the server.
The fake server answers 401, so the CLI logs an `rmcp ... HTTP 401` error. That is expected.

Keys live in `.env` next to the script (git-ignored, mode 600). Template: `.env.example`.
Never commit `.env` and never paste its contents into docs or logs.
Each run spends real API tokens.
