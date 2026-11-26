#!/usr/bin/env python3
"""The made-up run fake-agent.sh reports with FAKE_AGENT_SCENARIO=leaks: an agent repeating the
secrets it was given, the way a real one may, so check-masking.cjs can look for them in the log.

It reads the prompt on stdin, which check-masking.cjs fills with `multi: <a multiline secret>`
and `quote: <a secret with a quote and a backslash>`, and reports, in the events of the CLI it
was started as (Claude Code stream-json, or codex exec --json):

- the prompt as the agent's text, streamed in pieces (Claude) or as one message (Codex);
- a command printing the environment secret named by FAKE_AGENT_LEAK_ENV;
- an MCP tool result too large to keep as JSON, holding the quote secret JSON-escaped.

Calls nothing.
"""
import json
import os
import sys
import time

PAUSE = float(os.environ.get('FAKE_AGENT_PAUSE', '0.15'))


def ev(obj):
    print(json.dumps(obj), flush=True)
    time.sleep(PAUSE)


def after(prefix, text):
    """The rest of the prompt after the line starting with prefix, up to the next `name:` line."""
    lines = text.split('\n')
    for i, line in enumerate(lines):
        if line.startswith(prefix):
            value = [line[len(prefix):]]
            for rest in lines[i + 1:]:
                if rest.startswith('quote: '):
                    break
                value.append(rest)
            return '\n'.join(value).rstrip('\n')
    return ''


prompt = sys.stdin.read()
env_name = os.environ.get('FAKE_AGENT_LEAK_ENV', '')
env_value = os.environ.get(env_name, '') if env_name else ''
quote = after('quote: ', prompt)

# Rows of numbers only: cut to 200 characters a string, this is still over the 16 KB a result
# keeps as JSON. The secret goes first, so it is in what is left of the text.
big = json.dumps({'note': 'the key is ' + quote, 'rows': [{'n': i} for i in range(4000)]})

claude = '--output-format=stream-json' in sys.argv

if claude:
    ev({'type': 'system', 'subtype': 'init', 'model': 'claude-sonnet-5-5', 'mcp_servers': []})
    ev({'type': 'stream_event', 'event': {'type': 'message_start', 'message': {'id': 'msg_1'}}, 'parent_tool_use_id': None})
    text = 'You asked me:\n' + prompt.rstrip('\n') + '\nDone reading.\n'
    # In pieces of 7 characters, so a secret line is split between deltas too.
    for i in range(0, len(text), 7):
        ev({'type': 'stream_event', 'event': {'type': 'content_block_delta', 'index': 0,
            'delta': {'type': 'text_delta', 'text': text[i:i + 7]}}, 'parent_tool_use_id': None})
    ev({'type': 'stream_event', 'event': {'type': 'content_block_stop', 'index': 0}, 'parent_tool_use_id': None})
    ev({'type': 'assistant', 'message': {'id': 'msg_1', 'content': [{'type': 'text', 'text': text}]}, 'parent_tool_use_id': None})
    # A subagent's text is logged per whole message.
    ev({'type': 'assistant', 'message': {'id': 'msg_s', 'content': [{'type': 'text', 'text': 'Subagent saw:\n' + prompt}]},
        'parent_tool_use_id': 'toolu_sub'})
    ev({'type': 'assistant', 'message': {'id': 'msg_1', 'content': [{'type': 'tool_use', 'id': 'toolu_env', 'name': 'Bash',
        'input': {'command': 'printenv ' + env_name}}]}, 'parent_tool_use_id': None})
    ev({'type': 'user', 'message': {'role': 'user', 'content': [{'type': 'tool_result', 'tool_use_id': 'toolu_env',
        'content': env_value + '\n'}]}, 'parent_tool_use_id': None})
    ev({'type': 'assistant', 'message': {'id': 'msg_1', 'content': [{'type': 'tool_use', 'id': 'toolu_big', 'name': 'mcp__db__dump',
        'input': {'table': 'rows'}}]}, 'parent_tool_use_id': None})
    ev({'type': 'user', 'message': {'role': 'user', 'content': [{'type': 'tool_result', 'tool_use_id': 'toolu_big',
        'content': big}]}, 'parent_tool_use_id': None})
    ev({'type': 'result', 'subtype': 'success', 'is_error': False, 'result': 'done', 'num_turns': 1, 'duration_ms': 900})
else:
    ev({'type': 'thread.started', 'thread_id': 'leaks'})
    ev({'type': 'turn.started'})
    ev({'type': 'item.completed', 'item': {'id': 'item_0', 'type': 'agent_message', 'text': 'You asked me:\n' + prompt}})
    ev({'type': 'item.completed', 'item': {'id': 'item_1', 'type': 'command_execution', 'command': 'printenv ' + env_name,
        'aggregated_output': env_value + '\n', 'exit_code': 0, 'status': 'completed'}})
    ev({'type': 'item.completed', 'item': {'id': 'item_2', 'type': 'mcp_tool_call', 'server': 'db', 'tool': 'dump',
        'arguments': {'table': 'rows'}, 'result': {'content': [{'type': 'text', 'text': big}], 'structured_content': None},
        'error': None, 'status': 'completed'}})
    ev({'type': 'turn.completed', 'usage': {'input_tokens': 10, 'cached_input_tokens': 0, 'output_tokens': 5}})
