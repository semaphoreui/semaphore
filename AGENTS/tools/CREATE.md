# Creating a tool

How to add, document and retire an environment tool. What exists here now is in
[`INDEX.md`](INDEX.md).

## What goes here

- Put here a script only the agent runs to carry out an operation: a query to a stand database,
  an audit, a probe against an external service, a one-off check of a rule over the sources.
- Do not confuse this with the harness's built-in tools (Bash, Read, MCP): this is code in the
  repository, aware of the specific stand's container names, credentials and network model.
- Do not put the developers' own scripts here — building, running or setting up the project; they
  live where the project keeps them, not in this registry.

## Shape of a tool

- Give the tool its own folder: `AGENTS/tools/<tool>/`, with an executable file and a `README.md`
  inside.
- Name the folder after its subject, not a verb: `db-query`, `paddle-probe`.
- Let the tool run from the repository root: a path to it in the command, not a `cd` into it.

## A tool's `README.md`

This is a cheat sheet for invocation, not a description of the implementation. It is read to run
the tool, not to understand how it is written.

- Open with one sentence: what the tool does and why it exists.
- Give the call form — positional arguments and flags, on one line.
- Show two or three examples of real commands with what they return. An example convinces faster
  than a description, and checks out by copy-paste.
- Name the trap, if there is one: an irreversible action, an environment requirement, an access
  restriction.
- Keep it to one screen. A README longer than its script means the implementation leaked into it.
- Do not describe the internal workings: argument parsing, function order, the shape of
  intermediate data. That is read from the script itself, and a copy drifts from it silently while
  reading more confidently than the code.
- Update `README.md` when the call or the purpose changes. A change to the implementation alone
  does not touch it.

## Secrets

- Keep secrets out of the repository: the tool's own `.gitignore` in its folder, `.env.example`
  instead of `.env`.
- Also lock down the tool's output: a saved request dump or a provider's response can carry live
  credentials and customer data.
- Read the application's own credentials from its own `.env` when the tool reaches into the
  project's databases or providers. A second copy of a password drifts silently — a tool like this
  gets no `.env` of its own.

## The registry entry

- Keep one entry per tool in `INDEX.md`: file name, what it does, how it differs from its
  neighbor.
- Update the entry in the same pass that adds, renames or retires the tool.
- Keep the reason to open the tool in the registry; leave call details to its own `README.md`.

## Invariants

- A tool exists because a raw `docker` or SQL command does not know the project's names, grants
  and network model. Improve the tool, don't work around it by hand.
- The implementation lives in the code, the call in `README.md`, the existence in `INDEX.md`.
  Three different questions, three different places.
- A retired tool takes both its folder and its registry line with it — a dead entry costs more
  than a missing one.
