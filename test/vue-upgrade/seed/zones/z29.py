"""Z29 Applications and API tokens (/apps, /tokens).

One object: the admin API token «Z29 token» that expires on 2030-01-01T00:00:00Z — the only way to
get a real `tokens-rowExpiresAt` on /tokens (the base tokens never expire). Idempotent by name.

Nothing global on purpose: an own application would be an `apps.*` option seen by the «New
template» menu (Z04) and the template forms of other zones, so tests/zones/test_z29.py adds it to
the GET /api/apps answer (page.route) instead. Expired / revoked tokens, 12 tokens, a new token —
also page.route answers.
"""

TOKEN = {"name": "Z29 token", "expires_at": "2030-01-01T00:00:00Z"}


def seed(ctx):
    ctx.ensure("/user/tokens", dict(TOKEN), key="name", kind="api_token")
