# Plan — SMTP integration tests on Testcontainers

Branch `feat/project-alerts-mail-refactor`. Related: `project-alerts.md` (e-mail
channel with own SMTP server, outbound host policy).

## Goal

Prove that `util/mailer.SendMail` and the alerting e-mail channel deliver mail
through every SMTP server configuration Semaphore supports, against a real SMTP
daemon instead of the in-process fake used by the unit tests:

| Config in `config.json` / alert | Meaning |
|---|---|
| `email_secure=false` | plain SMTP, no auth (relay on port 25) |
| `email_secure=true, email_tls=false` | STARTTLS when offered, then AUTH PLAIN or LOGIN |
| `email_secure=true, email_tls=true` | implicit TLS from the first byte (port 465 style), then AUTH |
| `email_tls_min_version` | minimum TLS version for both modes |

The same four combinations exist per project alert through `smtp_secure` /
`smtp_tls` and a `login_password` access key.

## Non-goals

- TLS 1.0 / 1.1 servers. Go can neither serve nor (by default) dial them, and
  no maintained SMTP image offers them. `parseTlsVersion` stays unit-tested.
- Load, throughput and retry behaviour. One message per test.
- Testing real providers (Gmail, SES). Their quirks are covered by scenario
  shape (AUTH not advertised before STARTTLS, LOGIN only, etc.), not by
  network access.

## Tooling

- **`github.com/testcontainers/testcontainers-go`** for container lifecycle.
  Pin the version in `go.mod`. It pulls the Docker client and a long tail of
  dependencies into `go.sum`, but none of them ship: the package is behind
  the `integration` build tag and imported by tests only, so
  `go-licenses report ./...` never sees it and `THIRD-PARTY-LICENSES.md`
  does not list it. Regenerate the file anyway to prove that.
- **SMTP server image: `axllent/mailpit`** (pin a `v1.x` tag). Reasons:
  - one image covers plain, STARTTLS, implicit TLS, AUTH on/off and
    "AUTH only after TLS" through environment variables;
  - it exposes an HTTP API (`/api/v1/messages`, `/api/v1/message/{id}`) to
    read back what was delivered, including raw headers, so the tests can
    assert on the message and not just on "no error";
  - single static binary, starts in well under a second.

  Mailpit options the tests use:

  | Env | Effect |
  |---|---|
  | `MP_SMTP_AUTH="user:pass"` | enables AUTH PLAIN and LOGIN with those credentials |
  | `MP_SMTP_AUTH_ALLOW_INSECURE=true` | advertise AUTH without TLS (plain-text auth) |
  | `MP_SMTP_TLS_CERT`, `MP_SMTP_TLS_KEY` | enable STARTTLS with this certificate |
  | `MP_SMTP_REQUIRE_STARTTLS=true` | reject MAIL FROM until STARTTLS |
  | `MP_SMTP_REQUIRE_TLS=true` | implicit TLS on the SMTP port (no STARTTLS) |

  Ports: `1025/tcp` SMTP, `8025/tcp` HTTP API.

- **Optional second image for realism: a Postfix container** (`boky/postfix` or
  similar). Only if a scenario cannot be expressed with Mailpit. Not in the
  first iteration.

## Test layout

```
util/mailer/
  mailer.go
  mailer_test.go                     # existing unit tests, in-process fake
  integration/
    mailpit_test.go                  # container + API helpers
    certs_test.go                    # self-signed CA + server cert generation
    sendmail_test.go                 # SendMail scenarios
    channel_email_test.go            # alerting e-mail channel end to end
```

- Package `integration` under `util/mailer/integration/`, all files with
  `//go:build integration`. `go test ./...` keeps passing without Docker; the
  integration package is compiled and run only with `-tags integration`.
- No `TestMain` and no package-level cache (global variables are forbidden
  in this repo). Instead each top-level test starts **one** Mailpit container
  for a server configuration and runs its scenarios as subtests; each subtest
  clears the mailbox with `DELETE /api/v1/messages` before sending.
  `testcontainers.CleanupContainer` terminates the container when the test
  ends. Seven containers per run, about 30 s in total.
- When Docker is not reachable the package skips
  (`testcontainers.SkipIfProviderIsNotHealthy(t)`), it does not fail.

### Helpers

`mailpit_test.go`

```go
type mailpitConfig struct {
    auth              string     // "user:pass" or ""
    allowInsecureAuth bool
    tls               *certFiles // copy cert into the container, enable STARTTLS
    requireStartTLS   bool
    implicitTLS       bool       // MP_SMTP_REQUIRE_TLS
    allowedRecipients string     // MP_SMTP_ALLOWED_RECIPIENTS regexp
}

type mailpit struct {
    t              *testing.T
    container      testcontainers.Container
    host, smtpPort string // mapped on the Docker host
    apiURL         string
    ip             string // container address, reachable on Linux only
}

func startMailpit(t *testing.T, cfg mailpitConfig) *mailpit
func (m *mailpit) dial(ctx, network, _ string) (net.Conn, error) // to the mapped port
func (m *mailpit) options(host string) mailer.Options            // valid baseline message
func (m *mailpit) clear()
func (m *mailpit) messages() []mailpitMessage
func (m *mailpit) raw(id string) string                          // headers+body
func (m *mailpit) waitForMessages(n int) []mailpitMessage
func (m *mailpit) waitForOne() mailpitMessage
func (m *mailpit) assertNoMessages()
```

`certs_test.go` generates, per top-level test and into `t.TempDir()`:

- a CA key pair;
- a server certificate signed by it with SANs `localhost`, `127.0.0.1`, `::1`
  and `smtp.test` (see "Host name vs address" below);
- an unrelated self-signed certificate for the "untrusted CA" scenario.

Files are copied into the container at `/certs` with
`testcontainers.ContainerFile` (bind mounts are unreliable under Docker Desktop).

### Host name vs address

The mailer verifies the certificate against `Options.Host` and
`plainOrLoginAuth` allows plain-text AUTH only when the host is
`localhost` / `127.0.0.1` / `::1`. Testcontainers exposes the container on
`localhost:<random port>`. To exercise the "remote host" branches without
DNS, the tests set `Options.Host = "smtp.test"` and route the connection
with `Options.Dial` to the mapped address:

```go
dial := func(ctx context.Context, network, _ string) (net.Conn, error) {
    return (&net.Dialer{}).DialContext(ctx, network, net.JoinHostPort(m.smtpHost, m.smtpPort))
}
```

The server certificate carries `smtp.test` in its SANs, so verification
passes for the remote-host scenarios and fails only where the test wants it
to.

### Required production change

`mailer.Options` has no way to trust a private CA: `tls.Config.RootCAs` is
nil, so a self-signed test certificate can never verify. Add

```go
// RootCAs overrides the system pool used to verify the server certificate.
// nil keeps the system pool.
RootCAs *x509.CertPool
```

to `Options`, plumb it into `tlsConfig`. `Send` (legacy wrapper) and the
alerting channel leave it nil. A user-facing `email_tls_ca_file` config option
is a separate feature and out of scope here; note it as a follow-up.

## Scenario matrix

Each scenario is one `t.Run` in `sendmail_test.go` unless stated otherwise.
"Expect delivered" means: no error, exactly one message in Mailpit with the
expected From, To, Subject and body.

### A. Transport and authentication

| # | Mailpit config | Client options | Expect |
|---|---|---|---|
| A1 | no auth, no TLS | `Secure=false` | delivered |
| A2 | no auth, no TLS | `Secure=true, TLS=false`, user/pass set | delivered, no AUTH sent (server does not advertise it) |
| A3 | auth, allow insecure, no TLS | `Secure=true, TLS=false`, `Host=localhost` | delivered via plain-text AUTH (localhost exception) |
| A4 | auth, allow insecure, no TLS | `Secure=true, TLS=false`, `Host=smtp.test` via Dial | error `unencrypted connection`, nothing delivered |
| A5 | auth, STARTTLS | `Secure=true, TLS=false`, `RootCAs=test CA` | delivered with a username: AUTH is offered after STARTTLS only, so that proves the upgrade (Mailpit's `Received` header says `with SMTP` either way) |
| A6 | auth, STARTTLS, require STARTTLS | `Secure=false` | error from server on MAIL FROM, nothing delivered |
| A7 | auth, implicit TLS | `Secure=true, TLS=true`, `RootCAs=test CA` | delivered |
| A8 | auth, implicit TLS | `Secure=true, TLS=false` | error (STARTTLS client talks plain text to a TLS socket), no hang: fails within the 15 s default timeout |
| A9 | auth, STARTTLS | `Secure=true, TLS=true` | error (TLS handshake against a plain-text greeting), no hang |
| A10 | auth, STARTTLS | wrong password | error containing the 535 reply, nothing delivered |
| A11 | auth, STARTTLS | correct password, `Username` empty | error, nothing delivered |

### B. Certificate verification

| # | Mailpit config | Client options | Expect |
|---|---|---|---|
| B1 | STARTTLS with test CA cert | `RootCAs=nil` (system pool) | `x509` error, nothing delivered |
| B2 | STARTTLS with the unrelated self-signed cert | `RootCAs=test CA` | `x509` error |
| B3 | STARTTLS, cert SANs do not include host | `Host=smtp.test` via Dial, cert without that SAN | `x509` host-name error |
| B4 | implicit TLS | `TLSMinVersion="1.3"` | delivered (Mailpit supports 1.3) |
| B5 | implicit TLS | `TLSMinVersion="9.9"` | `unsupported TLS version` error before any dial (assert Dial never called) |

### C. Message content

| # | Input | Expect (via raw message from the API) |
|---|---|---|
| C1 | `Subject="Task\r\nBcc: x@y.z"` | single `Subject:` header `TaskBcc: x@y.z`, no `Bcc` header |
| C2 | `From="a@b.c\nX-Injected: 1"` | envelope sender and `From:` header contain no injected header |
| C3 | HTML body with UTF-8 | `Content-Type: text/html; charset=UTF-8`, body intact |
| C4 | `Date` header | present, parses with RFC 1123 |

### D. Timeouts and cancellation

| # | Setup | Expect |
|---|---|---|
| D1 | `Dial` returns a connection to a listener that never greets (in-process, no container) | `SendMail` returns within the context deadline |
| D2 | context already cancelled | error, `Dial` not called or returns ctx error; nothing delivered |
| D3 | Mailpit paused with `docker pause` mid-session (optional, Linux only) | `SendMail` returns after the deadline, not hangs |

### E. Alerting e-mail channel end to end (`channel_email_test.go`)

Uses `services/alerting` with a real `util.ConfigType` and `Destination`.

| # | Setup | Expect |
|---|---|---|
| E1 | server-wide SMTP = Mailpit on `localhost`, `Trusted` destination, 3 recipients | 3 messages, one per recipient, same subject |
| E2 | project alert with own SMTP host = Mailpit `containerIP` (untrusted, `guardedDialContext`), Linux only, skipped elsewhere | delivered |
| E3 | project alert with own SMTP host = `localhost:<mapped port>` (untrusted) | `smtp_host is not allowed`, Mailpit sees no connection |
| E4 | project alert with own credentials (`login_password` key) against Mailpit with auth + STARTTLS | delivered with AUTH |
| E5 | one recipient rejected by the server (Mailpit `MP_SMTP_DISABLE_RDNS` irrelevant; use a syntactically invalid RCPT the server refuses) | error is joined per recipient; the other recipients still receive mail |

E2 needs the container's bridge IP, which is routable on Linux (CI) but not
under Docker Desktop on macOS. Guard with
`if runtime.GOOS != "linux" { t.Skip(...) }`.

## CI

- `Taskfile.yml`: add `test:be:integration` running
  `go test -tags integration -count=1 ./util/mailer/integration/...`.
- `.github/workflows/dev.yml`: new job `test-integration` on
  `ubuntu-latest` (Docker is preinstalled), `needs: build-local`, runs the
  task above. Keep it separate from the unit test job so a Docker Hub hiccup
  does not block unit results. Set `TESTCONTAINERS_RYUK_DISABLED=false`
  (default) so containers are reaped if the job is cancelled.
- Cache the Mailpit image between runs is not worth it (tens of MB); pull on
  demand.

## Definition of done

- All scenarios above implemented, green on Linux CI and green or skipped
  (E2) on macOS. D3 (`docker pause`) works on Docker Desktop too.
- `go test ./...` without the tag still needs no Docker.
- `Options.RootCAs` added with a doc comment; `THIRD-PARTY-LICENSES.md`
  regenerated.
- `docs/docs/user-guide/alerts.md` does not change (no user-facing behaviour
  changes). If `email_tls_ca_file` is picked up as a follow-up, it gets its
  own plan and `task docs:gen`.

## Tasks

1. **Bootstrap**: add `testcontainers-go`, regenerate third-party licenses,
   create `util/mailer/integration/` with build tag and Docker skip.
2. **Helpers**: certificate generation, `startMailpit`, API client
   (`clear`, `messages`, `raw`, `waitForMessages`).
3. **Production change**: `Options.RootCAs` + unit test in
   `util/mailer/mailer_test.go` that a custom pool is honoured (in-process
   TLS listener with a self-signed cert).
4. **Scenarios A + B** (transport, auth, certificates).
5. **Scenarios C + D** (content, timeouts).
6. **Scenarios E** (alerting channel), including the Linux-only guard.
7. **CI**: Taskfile target and workflow job; verify a run on a PR.

Tasks 4–6 are independent once 1–3 land and can go to separate agents.
