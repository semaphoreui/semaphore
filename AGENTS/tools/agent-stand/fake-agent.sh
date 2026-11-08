#!/bin/sh
# Stands in for claude / codex on the agent stand: shows what the agent would get, calls nothing.
echo "fake agent args: $*"
echo "prompt: $(cat)"
echo "mcp token variables: $(env | grep -c '^SEMAPHORE_MCP_TOKEN_')"
echo "first token: ${SEMAPHORE_MCP_TOKEN_1:-<none>}"
echo "anthropic key set: $([ -n "${ANTHROPIC_API_KEY:-}" ] && echo yes || echo no)"
