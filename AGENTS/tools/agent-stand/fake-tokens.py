#!/usr/bin/env python3
"""What fake-agent.sh reports with FAKE_AGENT_SCENARIO=tokens, started as Claude Code: for each MCP
server of --mcp-config, its URL and a short hash of the token the agent would send there, so
check-runner-tokens.cjs can tell which token went with which URL (the token itself is masked).
Calls nothing."""
import hashlib
import json
import os
import re
import sys

sys.stdin.read()
config = next((a.split('=', 1)[1] for a in sys.argv[1:] if a.startswith('--mcp-config=')), '')
servers = json.load(open(config))['mcpServers'] if config else {}
for name, server in sorted(servers.items()):
    header = (server.get('headers') or {}).get('Authorization', '')
    match = re.search(r'\$\{(\w+)\}', header)
    token = os.environ.get(match.group(1), '') if match else ''
    digest = hashlib.sha256(token.encode()).hexdigest()[:12] if token else 'none'
    print(f'mcp server {name} url={server["url"]} token-sha={digest}', flush=True)
