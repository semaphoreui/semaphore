# Expressions in the agent prompt: `${{ … }}`

Status: implemented 2026-11-12 on `feat/agent_app` (TASK@A3824ECC91), core and `pro_impl`.

## What it is

The prompt kept in an agent template can name the variables and secrets of the template's
Variable Groups, in the spelling of GitHub Actions. They are replaced when the task starts.
A prompt a task brings itself (`allow_override_prompt`) goes to the agent as it is written.

| Expression | Value |
| --- | --- |
| `${{ vars.extra.NAME }}` | an extra variable (`Environment.JSON`, with the survey variables of the task) |
| `${{ vars.env.NAME }}` | an environment variable (`Environment.ENV`) |
| `${{ secrets.extra.NAME }}` | a secret of type `var`, or a secret survey variable |
| `${{ secrets.env.NAME }}` | a secret of type `env` |

```
pkg/agentprompt            Validate / Expand: the grammar, no knowledge of where values come from
db.Template.AgentPrompt        the prompt as written, expressions checked — validates a new task
db.Template.ExpandAgentPrompt  the prompt with values + the secrets put into it — at the run
db_lib.NewAgentPromptValues    the values, from the Environment, the survey secrets, the survey vars
  ← services/tasks/local_executor.go (Prepare) → LocalAppRunningArgs.AgentPromptValues → AgentApp.Run
  ← pro_impl/services/tasks/containerexec/agent.go (agentPlan) → /secrets/agent/prompt
```

## Rules

- **A missing variable fails the task**, with a message naming the expression. GitHub puts an
  empty string there; a prompt with a hole in it is an instruction nobody wrote.
- **A malformed `vars.…` / `secrets.…` expression is refused when the template is saved**:
  `${{ secrets.TOKEN }}` has no `extra`/`env`.
- **An expression of another context stays as it is** — `${{ github.sha }}`, `${{ matrix.os }}`:
  an agent is asked to write workflows. `$${{ … }}` gives a literal `${{ … }}` for the `vars` and
  `secrets` ones. An unclosed `${{` stays too.
- **A value is not looked through again**: a variable holding `${{ secrets.env.X }}` gives that
  text, not the secret.
- **A name** is everything after `extra.` / `env.` without spaces, so `a.b` is one name.
- **Values as text**: a string as it is, a number as it is written (`UseNumber`, no `1e+07`),
  a list or an object as JSON, `null` as an empty string.
- **Survey variables have no context of their own** (the owner's decision, 2026-11-12: no
  `inputs.…`): they are extra variables, so `vars.extra.NAME`, a secret one `secrets.extra.NAME`.
  One delivered as an environment variable (`target: env`) is under `env` instead. A secret
  survey variable wins over a Variable Group secret of its name, as for a script.
- **A survey variable the task left unanswered is an empty string**, not a missing variable —
  unless the Variable Groups have a value of that name, which then stays.

## Security

- **Only the prompt kept in the template is expanded** (the owner's decision, 2026-11-12). A
  prompt given by a task (`allow_override_prompt`) is neither expanded nor checked — its
  `${{ … }}` reach the agent as text (`db.Template.ExpandAgentPrompt`, `fromTask`). Whoever may
  only run a task would otherwise read, through the agent, the `var` secrets the agent does not
  get at all. The task form says so under its Prompt field, because the field starts with the
  template's prompt: editing it there turns the expressions into text.
- **A substituted secret is masked in the task log** like an MCP token: its value is added to
  `AgentLaunch.Secrets` before `NewOutput()` (local) and `AppPlan.setSecrets` (Pro), so both the
  events and the plain lines carry `********`. Only the secrets the prompt uses are masked.
- The prompt still goes to the CLI on stdin (local) or as a file under `/secrets/agent` (Pro),
  never in argv or the script.
- ⚠️ 2026-11-12: a secret in the prompt is sent to the model provider, and the agent can do with
  it what the prompt allows. That is what the expression is for; the form does not warn about it.
- An error message names the expression, never a value.

## Open

- The Pro executors read the secret survey variables from `Task.Secret`; nothing else in
  `containerexec` uses them, so whether the Docker/K8s path receives them was not checked in a
  real container.
- The user docs are `docs/docs/user-guide/apps/ai-agents.md` (and its ten translations): a change
  of a rule here changes that page.
- Check after a change: `AGENTS/tools/agent-stand/check-prompt.py`.
