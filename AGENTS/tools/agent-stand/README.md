# agent-stand

A throwaway Semaphore from this work copy on `http://127.0.0.1:3102` (admin/admin123, its own
SQLite) with the AI agent UI seeded. It listens on the loopback interface only — `stand.sh` writes
`"interface": "127.0.0.1"` into the config on every start — because the admin password is known
and `--real` puts real API keys behind it. Use it to check the mcp inventory form, the agent template
form and agent task runs in the browser after changing them.

```
AGENTS/tools/agent-stand/stand.sh {build|start [--real]|stop|status|log [n]}
AGENTS/tools/agent-stand/seed.sh
NODE_PATH=~/.npm/_npx/9833c18b2d85bc59/node_modules node AGENTS/tools/agent-stand/verify.cjs
NODE_PATH=~/.npm/_npx/9833c18b2d85bc59/node_modules node AGENTS/tools/agent-stand/shoot.cjs [out_dir] [--dark]
```

```
$ AGENTS/tools/agent-stand/stand.sh build && AGENTS/tools/agent-stand/stand.sh start
stand: up at http://127.0.0.1:3102 (pid 93629, agents: fake)
$ AGENTS/tools/agent-stand/seed.sh
project=1 key=2 repo=1 env=1 inventory=1 claude_template=1 codex_template=2
$ node AGENTS/tools/agent-stand/verify.cjs
1 edit inventory: ok [{"n":"github"},{"n":"docs","t":"delete"},{"n":"search","t":"create"}]
...
5 run with template prompt: ok, task 2
```

- `verify.cjs` saves through the UI and checks the API and the task logs: token operations,
  no token in a GET or a log, the task prompt reaching the agent. It changes the seeded data, so
  shoot first.
- By default `claude` and `codex` run `fake-agent.sh`, which prints what it got and calls
  nothing. `start --real` uses the CLIs from `PATH`: a task then calls the vendor API with the
  key of the Variable Group and the MCP servers of the inventory — replace the seeded fake URLs
  and tokens first.
- Data, logs and screenshots live in `AGENTS/tmp/agent-stand` (`AGENT_STAND_DIR` to move it);
  delete the folder to start over. The schema is this branch's, never point another build at it.
- `build` runs the Vite build, which empties `api/public`; the script restores the tracked
  `api/public/.gitkeep`.
- A console `400` on `/repositories/1/branches` is expected: the seeded repository is a local
  path.
