#!/usr/bin/env python3
"""Checks the ${{ … }} expressions of an agent prompt on the agent stand (stand.sh start, seed.sh).

Gives the seeded Variable Group an extra variable, an environment variable and two secrets, puts
the four expressions into the prompt of the Claude and the Codex templates, runs a task of each
and reads its log: the fake CLI prints the prompt it got, so the values must be there and the
secrets masked. Then checks what is refused: an expression naming a variable that does not exist
(the task fails), a malformed one (the template is not saved). A prompt given by a task must
reach the agent as it is written. Puts the prompts of the templates back.

  python3 AGENTS/tools/agent-stand/check-prompt.py
"""
import http.cookiejar
import json
import os
import sys
import time
import urllib.error
import urllib.request

BASE = os.environ.get('BASE', 'http://127.0.0.1:3102')
ROOT = os.path.dirname(os.path.abspath(__file__))
DATA = os.environ.get('AGENT_STAND_DIR', os.path.join(ROOT, '..', '..', 'tmp', 'agent-stand'))

ids = dict(kv.split('=') for kv in open(os.path.join(DATA, 'ids.txt')).read().split())
P = ids['project']

opener = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar()))


def call(method, path, body=None):
    """Returns (status, decoded body or text)."""
    data = None if body is None else json.dumps(body).encode()
    req = urllib.request.Request(BASE + path, data=data, method=method,
                                 headers={'Content-Type': 'application/json'})
    try:
        with opener.open(req) as res:
            status, raw = res.status, res.read().decode()
    except urllib.error.HTTPError as e:
        status, raw = e.code, e.read().decode()
    try:
        return status, json.loads(raw)
    except ValueError:
        return status, raw


def must(method, path, body=None):
    status, res = call(method, path, body)
    assert status < 300, f'{method} {path}: {status} {res}'
    return res


def run(template_id, params=None):
    """Runs a task, waits for its end; returns (status, raw log)."""
    task = must('POST', f'/api/project/{P}/tasks', {'template_id': int(template_id), 'params': params or {}})
    for _ in range(240):
        status = must('GET', f'/api/project/{P}/tasks/{task["id"]}')['status']
        if status in ('success', 'error', 'stopped'):
            break
        time.sleep(0.5)
    time.sleep(1)
    return status, must('GET', f'/api/project/{P}/tasks/{task["id"]}/raw_output')


def set_prompt(template_id, prompt, allow_override=None):
    tpl = must('GET', f'/api/project/{P}/templates/{template_id}')
    old = dict(tpl.get('task_params') or {})
    tpl['task_params'] = dict(old, prompt=prompt)
    if allow_override is not None:
        tpl['task_params']['allow_override_prompt'] = allow_override
    return old, call('PUT', f'/api/project/{P}/templates/{template_id}', tpl)


must('POST', '/api/auth/login', {'auth': 'admin', 'password': 'admin123'})

# The Variable Group of the templates: one value of each kind.
env = must('GET', f'/api/project/{P}/environment/{ids["env"]}')
extra = json.loads(env.get('json') or '{}')
extra.update({'prompt_repo': 'semaphoreui/semaphore', 'prompt_count': 3})
envs = json.loads(env.get('env') or '{}')
envs.update({'PROMPT_STAGE': 'staging'})
secrets = []
for kind, name, value in (('var', 'prompt_password', 'extra-secret-7f3a'), ('env', 'PROMPT_API_KEY', 'env-secret-91bc')):
    known = [s for s in env.get('secrets') or [] if s['type'] == kind and s['name'] == name]
    secrets.append({'type': kind, 'name': name, 'secret': value,
                    **({'id': known[0]['id'], 'operation': 'update'} if known else {'operation': 'create'})})
env.update({'json': json.dumps(extra), 'env': json.dumps(envs), 'secrets': secrets})
must('PUT', f'/api/project/{P}/environment/{ids["env"]}', env)

PROMPT = ('Review ${{ vars.extra.prompt_repo }} (${{vars.extra.prompt_count}} issues) on ${{ vars.env.PROMPT_STAGE }}. '
          'Password ${{ secrets.extra.prompt_password }}, key ${{ secrets.env.PROMPT_API_KEY }}. '
          'Keep ${{ github.sha }} and $${{ vars.env.PROMPT_STAGE }}.')
EXPECTED = ('prompt: Review semaphoreui/semaphore (3 issues) on staging. Password ********, key ********. '
            'Keep ${{ github.sha }} and ${{ vars.env.PROMPT_STAGE }}.')

restore = {}
try:
    for app in ('claude', 'codex'):
        tid = ids[f'{app}_template']
        restore[tid], (status, res) = set_prompt(tid, PROMPT)
        assert status < 300, f'save the {app} template: {status} {res}'

        status, log = run(tid)
        assert status == 'success', f'{app}: {status}\n{log}'
        assert EXPECTED in log, f'{app}: the prompt in the log\n{log}'
        assert 'extra-secret-7f3a' not in log and 'env-secret-91bc' not in log, f'{app}: a secret in the log\n{log}'
        print(f'{app}: the values are in the prompt, the secrets are masked')

    tid = ids['claude_template']

    _, (status, res) = set_prompt(tid, 'Use ${{ vars.extra.no_such_variable }}.')
    assert status < 300, f'{status} {res}'
    status, log = run(tid)
    assert status == 'error', f'a missing variable: {status}\n{log}'
    assert 'no extra variable "no_such_variable"' in log, log
    print('a missing variable: the task fails and names it')

    _, (status, res) = set_prompt(tid, 'Use ${{ secrets.PROMPT_API_KEY }}.')
    assert status == 400 and 'invalid prompt expression' in json.dumps(res), f'{status} {res}'
    print('a malformed expression: the template is not saved')

    _, (status, res) = set_prompt(tid, PROMPT, allow_override=True)
    assert status < 300, f'{status} {res}'
    own = 'Stage ${{ vars.env.PROMPT_STAGE }}, key ${{ secrets.env.PROMPT_API_KEY }}, ${{ vars.bad }}.'
    status, log = run(tid, {'prompt': own})
    assert status == 'success' and f'prompt: {own}' in log, f'{status}\n{log}'
    assert 'staging' not in log and 'env-secret-91bc' not in log, log
    print('a prompt given by a task: reaches the agent as it is written')
finally:
    for tid, params in restore.items():
        tpl = must('GET', f'/api/project/{P}/templates/{tid}')
        tpl['task_params'] = params
        must('PUT', f'/api/project/{P}/templates/{tid}', tpl)

print('ok')
sys.exit(0)
