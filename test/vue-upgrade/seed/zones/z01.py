"""Z01 Sign in and invitation (Auth.vue, AcceptInvite.vue): no data of its own.

Every state of the catalogue (subtask TASK@9ABF1875F4, note «Z01: стартовый интерфейс») comes from the
base set (admin, seed_runner, Seed Demo id 1) or from interception in the tests: LDAP / OIDC / TOTP
login options (GET /api/auth/login), the second factor (GET /api/user 401 TOTP_REQUIRED /
EMAIL_OTP_REQUIRED), every sign-in, verify, recovery, logout and invitation answer. The stand needs no
LDAP, OIDC, TOTP or invite, so the zone adds nothing to the fixture; the module exists so
`./vu.sh seed zone z01` works.
"""


def seed(ctx):
    ctx.log("=", "zone z01", "no data: base set + interception")
