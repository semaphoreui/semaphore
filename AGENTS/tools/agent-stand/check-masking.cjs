// Checks in the browser that the task log of an agent run masks every secret the agent repeats:
// an environment secret of the Variable Group the agent prints with a command, a multiline secret
// put into the prompt (the agent repeats it as text: streamed by Claude, one message by Codex), a
// secret with a quote and a backslash inside a JSON tool result too large to stay JSON.
// The fake CLI reports such a run with FAKE_AGENT_SCENARIO=leaks (fake-leaks.py). Opens every
// call of the log, then reads the log view, the raw log and the stored events. Puts the prompts
// back and takes the scenario out of the Variable Group; the secrets stay.
// NODE_PATH=~/.npm/_npx/9833c18b2d85bc59/node_modules node AGENTS/tools/agent-stand/check-masking.cjs
const path = require('path');
const fs = require('fs');
const assert = require('assert');
const { chromium } = require('playwright');

const BASE = process.env.BASE || 'http://127.0.0.1:3102';
const DATA = process.env.AGENT_STAND_DIR || path.join(__dirname, '..', '..', 'tmp', 'agent-stand');
const ids = Object.fromEntries(fs.readFileSync(path.join(DATA, 'ids.txt'), 'utf8').trim().split(' ').map((kv) => kv.split('=')));
const P = ids.project;

const SECRETS = {
  env: { type: 'env', name: 'STAND_LEAK_TOKEN', value: 'env-leak-5e1d', parts: ['env-leak-5e1d'] },
  multiline: {
    type: 'var',
    name: 'leak_multi',
    value: '-----BEGIN TEST KEY-----\nmulti-line-one-4c2a\nmulti-line-two-9b7e\n-----END TEST KEY-----',
    parts: ['multi-line-one-4c2a', 'multi-line-two-9b7e'],
  },
  quote: { type: 'var', name: 'leak_quote', value: 'quo"te\\back-3d8f', parts: ['back-3d8f'] },
};
const PROMPT = 'multi: ${{ secrets.extra.leak_multi }}\nquote: ${{ secrets.extra.leak_quote }}';

(async () => {
  const browser = await chromium.launch({ channel: 'chrome' });
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 }, baseURL: BASE });
  assert.ok((await ctx.request.post('/api/auth/login', { data: { auth: 'admin', password: 'admin123' } })).ok());
  const api = async (method, url, data) => {
    const res = await ctx.request.fetch(url, { method, data });
    assert.ok(res.ok(), `${method} ${url}: ${res.status()} ${await res.text()}`);
    return res.status() === 204 ? null : res.json().catch(() => null);
  };
  const page = await ctx.newPage();
  const at = (testId) => page.locator(`[data-testid="${testId}"]`);

  const envUrl = `/api/project/${P}/environment/${ids.env}`;
  const setScenario = async (on) => {
    const env = await api('GET', envUrl);
    const vars = JSON.parse(env.env || '{}');
    if (on) {
      Object.assign(vars, { FAKE_AGENT_SCENARIO: 'leaks', FAKE_AGENT_LEAK_ENV: SECRETS.env.name, FAKE_AGENT_PAUSE: '0.02' });
    } else {
      ['FAKE_AGENT_SCENARIO', 'FAKE_AGENT_LEAK_ENV', 'FAKE_AGENT_PAUSE'].forEach((k) => delete vars[k]);
    }
    const secrets = on ? Object.values(SECRETS).map((s) => {
      const known = (env.secrets || []).find((k) => k.type === s.type && k.name === s.name);
      return { type: s.type, name: s.name, secret: s.value, ...(known ? { id: known.id, operation: 'update' } : { operation: 'create' }) };
    }) : [];
    await api('PUT', envUrl, { ...env, env: JSON.stringify(vars), secrets });
  };

  const restore = {};
  const leaks = [];
  try {
    await setScenario(true);
    for (const app of ['claude', 'codex']) {
      const tid = ids[`${app}_template`];
      const tpl = await api('GET', `/api/project/${P}/templates/${tid}`);
      restore[tid] = tpl.task_params;
      await api('PUT', `/api/project/${P}/templates/${tid}`, { ...tpl, task_params: { ...tpl.task_params, prompt: PROMPT } });

      const task = await api('POST', `/api/project/${P}/tasks`, { template_id: Number(tid) });
      await page.goto(`/project/${P}/templates/${tid}?t=${task.id}`);
      for (let i = 0; i < 120; i += 1) {
        // eslint-disable-next-line no-await-in-loop
        const t = await api('GET', `/api/project/${P}/tasks/${task.id}`);
        if (['success', 'error', 'stopped'].includes(t.status)) {
          assert.strictEqual(t.status, 'success', `${app} task ${task.id}: ${t.status}`);
          break;
        }
        // eslint-disable-next-line no-await-in-loop
        await page.waitForTimeout(500);
      }
      await at('taskLog-agentDone').waitFor({ timeout: 20000 });
      await page.waitForTimeout(1500);
      const tools = at('taskLog-agentTool');
      for (let i = 0; i < await tools.count(); i += 1) {
        // eslint-disable-next-line no-await-in-loop
        await tools.nth(i).click();
      }
      await page.waitForTimeout(300);

      const where = {
        'log view: text': (await at('taskLog-agentText').allInnerTexts()).join('\n'),
        'log view: tool calls': (await at('taskLog-agentToolDetails').allInnerTexts()).join('\n'),
        'log view: whole': await at('taskLog-output').last().innerText(),
        'raw log': await (await ctx.request.get(`/api/project/${P}/tasks/${task.id}/raw_output`)).text(),
        'stored events': await (await ctx.request.get(`/api/project/${P}/tasks/${task.id}/output`)).text(),
      };
      fs.mkdirSync(path.join(DATA, 'shots-masking'), { recursive: true });
      await at('taskLog-page').last().screenshot({ path: path.join(DATA, 'shots-masking', `${app}.png`) });

      for (const [kind, secret] of Object.entries(SECRETS)) {
        for (const part of secret.parts) {
          const found = Object.entries(where).filter(([, text]) => text.includes(part)).map(([w]) => w);
          if (found.length) leaks.push(`${app} task ${task.id}: ${kind} secret (${part}) in ${found.join(', ')}`);
        }
      }
      console.log(`${app}: task ${task.id} checked`);
    }
  } finally {
    for (const [tid, params] of Object.entries(restore)) {
      // eslint-disable-next-line no-await-in-loop
      const tpl = await api('GET', `/api/project/${P}/templates/${tid}`);
      // eslint-disable-next-line no-await-in-loop
      await api('PUT', `/api/project/${P}/templates/${tid}`, { ...tpl, task_params: params });
    }
    await setScenario(false);
    await browser.close();
  }

  if (leaks.length) {
    console.log(`LEAKED:\n${leaks.join('\n')}`);
    process.exit(1);
  }
  console.log('ok: no secret in the log');
})().catch((e) => { console.error(e); process.exit(1); });
