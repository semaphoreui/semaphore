# Third-Party Licenses

Semaphore UI is built on the work of many open-source projects. This document identifies every third-party component distributed with Semaphore UI, in compliance with the attribution requirements of the respective licenses and with §3.6 of our Master Service Agreement (identification of open-source components by name, version, and license type).

_Generated on **2026-05-05 08:19 UTC** by `scripts/collect_licenses.sh`._

To regenerate this file, run:

```bash
scripts/collect_licenses.sh
scripts/check_policy.py .licenses-cache/
scripts/generate_md.py .licenses-cache/ > THIRD-PARTY-LICENSES.md
```

## Summary
This document lists **116** third-party components distributed with Semaphore UI, grouped by ecosystem.

| Ecosystem | Components |
|-----------|------------|
| Go (backend) | 63 |
| npm (frontend) | 53 |

### License distribution

| License | Count |
|---------|-------|
| MIT | 72 |
| BSD-3-Clause | 27 |
| Apache-2.0 | 12 |
| BSD-2-Clause | 3 |
| MPL-2.0 | 1 |
| ISC | 1 |

## Go Backend Dependencies
Modules statically linked into the Semaphore UI server binary. Sourced from `go.mod` (production dependencies only).

| Component | Version(s) | License | Source |
|-----------|------------|---------|--------|
| `dario.cat/mergo` | v1.0.1 | BSD-3-Clause | [link](https://github.com/imdario/mergo/blob/v1.0.1/LICENSE) |
| `filippo.io/edwards25519` | v1.1.1 | BSD-3-Clause | [link](https://github.com/FiloSottile/edwards25519/blob/v1.1.1/LICENSE) |
| `github.com/Azure/go-ntlmssp` | v0.1.1 | MIT | [link](https://github.com/Azure/go-ntlmssp/blob/v0.1.1/LICENSE) |
| `github.com/boombuler/barcode` | v1.0.1-0.20190219062509-6c824513bacc | MIT | [link](https://github.com/boombuler/barcode/blob/6c824513bacc/LICENSE) |
| `github.com/cloudflare/circl` | v1.6.3 | BSD-3-Clause | [link](https://github.com/cloudflare/circl/blob/v1.6.3/LICENSE) |
| `github.com/coreos/go-oidc/v3` | v3.17.0 | Apache-2.0 | [link](https://github.com/coreos/go-oidc/blob/v3.17.0/LICENSE) |
| `github.com/creack/pty` | v1.1.24 | MIT | [link](https://github.com/creack/pty/blob/v1.1.24/LICENSE) |
| `github.com/cyphar/filepath-securejoin` | v0.4.1 | BSD-3-Clause | [link](https://github.com/cyphar/filepath-securejoin/blob/v0.4.1/LICENSE) |
| `github.com/dustin/go-humanize` | v1.0.1 | MIT | [link](https://github.com/dustin/go-humanize/blob/v1.0.1/LICENSE) |
| `github.com/emirpasic/gods` | v1.18.1 | BSD-2-Clause | [link](https://github.com/emirpasic/gods/blob/v1.18.1/LICENSE) |
| `github.com/felixge/httpsnoop` | v1.0.4 | MIT | [link](https://github.com/felixge/httpsnoop/blob/v1.0.4/LICENSE.txt) |
| `github.com/go-asn1-ber/asn1-ber` | v1.5.8-0.20250403174932-29230038a667 | MIT | [link](https://github.com/go-asn1-ber/asn1-ber/blob/29230038a667/LICENSE) |
| `github.com/go-git/gcfg` | v1.5.1-0.20230307220236-3a3c6141e376 | BSD-3-Clause | [link](https://github.com/go-git/gcfg/blob/3a3c6141e376/LICENSE) |
| `github.com/go-git/go-billy/v5` | v5.8.0 | Apache-2.0 | [link](https://github.com/go-git/go-billy/blob/v5.8.0/LICENSE) |
| `github.com/go-git/go-git/v5` | v5.18.0 | Apache-2.0 | [link](https://github.com/go-git/go-git/blob/v5.18.0/LICENSE) |
| `github.com/go-gorp/gorp/v3` | v3.1.0 | MIT | [link](https://github.com/go-gorp/gorp/blob/v3.1.0/LICENSE) |
| `github.com/go-jose/go-jose/v4` | v4.1.4 | Apache-2.0 | [link](https://github.com/go-jose/go-jose/blob/v4.1.4/LICENSE) |
| `github.com/go-ldap/ldap/v3` | v3.4.12 | MIT | [link](https://github.com/go-ldap/ldap/blob/v3.4.12/v3/LICENSE) |
| `github.com/go-sql-driver/mysql` | v1.9.3 | MPL-2.0 | [link](https://github.com/go-sql-driver/mysql/blob/v1.9.3/LICENSE) |
| `github.com/golang/groupcache` | v0.0.0-20241129210726-2c02b8208cf8 | Apache-2.0 | [link](https://github.com/golang/groupcache/blob/2c02b8208cf8/LICENSE) |
| `github.com/google/go-github` | v17.0.0+incompatible | BSD-3-Clause | [link](https://github.com/google/go-github/blob/v17.0.0/LICENSE) |
| `github.com/google/go-querystring` | v1.1.0 | BSD-3-Clause | [link](https://github.com/google/go-querystring/blob/v1.1.0/LICENSE) |
| `github.com/google/uuid` | v1.6.0 | BSD-3-Clause | [link](https://github.com/google/uuid/blob/v1.6.0/LICENSE) |
| `github.com/gorilla/handlers` | v1.5.2 | BSD-3-Clause | [link](https://github.com/gorilla/handlers/blob/v1.5.2/LICENSE) |
| `github.com/gorilla/mux` | v1.8.1 | BSD-3-Clause | [link](https://github.com/gorilla/mux/blob/v1.8.1/LICENSE) |
| `github.com/gorilla/securecookie` | v1.1.2 | BSD-3-Clause | [link](https://github.com/gorilla/securecookie/blob/v1.1.2/LICENSE) |
| `github.com/gorilla/websocket` | v1.5.3 | BSD-2-Clause | [link](https://github.com/gorilla/websocket/blob/v1.5.3/LICENSE) |
| `github.com/jbenet/go-context` | v0.0.0-20150711004518-d14ea06fba99 | MIT | [link](https://github.com/jbenet/go-context/blob/d14ea06fba99/LICENSE) |
| `github.com/kevinburke/ssh_config` | v1.2.0 | MIT | [link](https://github.com/kevinburke/ssh_config/blob/v1.2.0/LICENSE) |
| `github.com/lann/builder` | v0.0.0-20180802200727-47ae307949d0 | MIT | [link](https://github.com/lann/builder/blob/47ae307949d0/LICENSE) |
| `github.com/lann/ps` | v0.0.0-20150810152359-62de8c46ede0 | MIT | [link](https://github.com/lann/ps/blob/62de8c46ede0/LICENSE) |
| `github.com/lib/pq` | v1.11.2 | MIT | [link](https://github.com/lib/pq/blob/v1.11.2/LICENSE) |
| `github.com/Masterminds/squirrel` | v1.5.4 | MIT | [link](https://github.com/Masterminds/squirrel/blob/v1.5.4/LICENSE) |
| `github.com/mattn/go-isatty` | v0.0.20 | MIT | [link](https://github.com/mattn/go-isatty/blob/v0.0.20/LICENSE) |
| `github.com/mdp/qrterminal/v3` | v3.2.1 | MIT | [link](https://github.com/mdp/qrterminal/blob/v3.2.1/LICENSE) |
| `github.com/ncruces/go-strftime` | v0.1.9 | MIT | [link](https://github.com/ncruces/go-strftime/blob/v0.1.9/LICENSE) |
| `github.com/pjbgf/sha1cd` | v0.3.2 | Apache-2.0 | [link](https://github.com/pjbgf/sha1cd/blob/v0.3.2/LICENSE) |
| `github.com/pquerna/otp` | v1.5.0 | Apache-2.0 | [link](https://github.com/pquerna/otp/blob/v1.5.0/LICENSE) |
| `github.com/ProtonMail/go-crypto` | v1.1.6 | BSD-3-Clause | [link](https://github.com/ProtonMail/go-crypto/blob/v1.1.6/LICENSE) |
| `github.com/remyoudompheng/bigfft` | v0.0.0-20230129092748-24d4a6f8daec | BSD-3-Clause | [link](https://github.com/remyoudompheng/bigfft/blob/24d4a6f8daec/LICENSE) |
| `github.com/robfig/cron/v3` | v3.0.1 | MIT | [link](https://github.com/robfig/cron/blob/v3.0.1/LICENSE) |
| `github.com/sergi/go-diff` | v1.3.2-0.20230802210424-5b0b94c5c0d3 | MIT | [link](https://github.com/sergi/go-diff/blob/5b0b94c5c0d3/LICENSE) |
| `github.com/sirupsen/logrus` | v1.9.4 | MIT | [link](https://github.com/sirupsen/logrus/blob/v1.9.4/LICENSE) |
| `github.com/skeema/knownhosts` | v1.3.1 | Apache-2.0 | [link](https://github.com/skeema/knownhosts/blob/v1.3.1/LICENSE) |
| `github.com/snikch/goodman` | v0.0.0-20171125024755-10e37e294daa | MIT | [link](https://github.com/snikch/goodman/blob/10e37e294daa/LICENSE) |
| `github.com/spf13/cobra` | v1.10.2 | Apache-2.0 | [link](https://github.com/spf13/cobra/blob/v1.10.2/LICENSE.txt) |
| `github.com/spf13/pflag` | v1.0.9 | BSD-3-Clause | [link](https://github.com/spf13/pflag/blob/v1.0.9/LICENSE) |
| `github.com/thedevsaddam/gojsonq/v2` | v2.5.2 | MIT | [link](https://github.com/thedevsaddam/gojsonq/blob/v2.5.2/LICENSE.md) |
| `github.com/xanzy/ssh-agent` | v0.3.3 | Apache-2.0 | [link](https://github.com/xanzy/ssh-agent/blob/v0.3.3/LICENSE) |
| `go.etcd.io/bbolt` | v1.4.1 | MIT | [link](https://github.com/etcd-io/bbolt/blob/v1.4.1/LICENSE) |
| `golang.org/x/crypto` | v0.48.0 | BSD-3-Clause | [link](https://cs.opensource.google/go/x/crypto/+/v0.48.0:LICENSE) |
| `golang.org/x/exp` | v0.0.0-20250620022241-b7579e27df2b | BSD-3-Clause | [link](https://cs.opensource.google/go/x/exp/+/b7579e27:LICENSE) |
| `golang.org/x/net` | v0.49.0 | BSD-3-Clause | [link](https://cs.opensource.google/go/x/net/+/v0.49.0:LICENSE) |
| `golang.org/x/oauth2` | v0.35.0 | BSD-3-Clause | [link](https://cs.opensource.google/go/x/oauth2/+/v0.35.0:LICENSE) |
| `golang.org/x/sys` | v0.41.0 | BSD-3-Clause | [link](https://cs.opensource.google/go/x/sys/+/v0.41.0:LICENSE) |
| `golang.org/x/term` | v0.40.0 | BSD-3-Clause | [link](https://cs.opensource.google/go/x/term/+/v0.40.0:LICENSE) |
| `gopkg.in/natefinch/lumberjack.v2` | v2.2.1 | MIT | [link](https://github.com/natefinch/lumberjack/blob/v2.2.1/LICENSE) |
| `gopkg.in/warnings.v0` | v0.1.2 | BSD-2-Clause | [link](https://github.com/go-warnings/warnings/blob/v0.1.2/LICENSE) |
| `modernc.org/libc` | v1.66.10 | BSD-3-Clause | [link](https://gitlab.com/cznic/libc/blob/v1.66.10/LICENSE-GO) |
| `modernc.org/mathutil` | v1.7.1 | BSD-3-Clause | [link](https://gitlab.com/cznic/mathutil/-/blob/master/LICENSE) |
| `modernc.org/memory` | v1.11.0 | BSD-3-Clause | [link](https://gitlab.com/cznic/memory/blob/v1.11.0/LICENSE-GO) |
| `modernc.org/sqlite` | v1.40.1 | BSD-3-Clause | [link](https://gitlab.com/cznic/sqlite/blob/v1.40.1/LICENSE) |
| `rsc.io/qr` | v0.2.0 | BSD-3-Clause | [link](https://github.com/rsc/qr/blob/v0.2.0/LICENSE) |

## Frontend Dependencies (npm)
Packages bundled into the web UI assets, which are embedded in the server binary at build time. Sourced from the frontend `package.json` (production dependencies only; dev dependencies are not distributed).

| Component | Version(s) | License | Source |
|-----------|------------|---------|--------|
| `@babel/helper-string-parser` | 7.25.9 | MIT | [link](https://github.com/babel/babel) |
| `@babel/helper-validator-identifier` | 7.25.9 | MIT | [link](https://github.com/babel/babel) |
| `@babel/parser` | 7.27.0 | MIT | [link](https://github.com/babel/babel) |
| `@babel/types` | 7.27.0 | MIT | [link](https://github.com/babel/babel) |
| `@mdi/font` | 7.4.47 | Apache-2.0 | [link](https://github.com/Templarian/MaterialDesign-Webfont) |
| `@vue/compiler-sfc` | 2.7.16 | MIT | — |
| `ansi_up` | 6.0.6 | MIT | [link](https://github.com/drudru/ansi_up) |
| `asynckit` | 0.4.0 | MIT | [link](https://github.com/alexindigo/asynckit) |
| `axios` | 1.12.0 | MIT | [link](https://github.com/axios/axios) |
| `call-bind-apply-helpers` | 1.0.2 | MIT | [link](https://github.com/ljharb/call-bind-apply-helpers) |
| `chart.js` | 3.9.1 | MIT | [link](https://github.com/chartjs/Chart.js) |
| `codemirror` | 5.65.6 | MIT | [link](https://github.com/codemirror/CodeMirror) |
| `combined-stream` | 1.0.8 | MIT | [link](https://github.com/felixge/node-combined-stream) |
| `core-js` | 3.41.0 | MIT | [link](https://github.com/zloirock/core-js) |
| `cron-parser` | 5.3.0 | MIT | [link](https://github.com/harrisiirak/cron-parser) |
| `csstype` | 3.1.3 | MIT | [link](https://github.com/frenic/csstype) |
| `dayjs` | 1.11.13 | MIT | [link](https://github.com/iamkun/dayjs) |
| `delayed-stream` | 1.0.0 | MIT | [link](https://github.com/felixge/node-delayed-stream) |
| `diff-match-patch` | 1.0.5 | Apache-2.0 | [link](https://github.com/JackuB/diff-match-patch) |
| `dunder-proto` | 1.0.1 | MIT | [link](https://github.com/es-shims/dunder-proto) |
| `es-define-property` | 1.0.1 | MIT | [link](https://github.com/ljharb/es-define-property) |
| `es-errors` | 1.3.0 | MIT | [link](https://github.com/ljharb/es-errors) |
| `es-object-atoms` | 1.1.1 | MIT | [link](https://github.com/ljharb/es-object-atoms) |
| `es-set-tostringtag` | 2.1.0 | MIT | [link](https://github.com/es-shims/es-set-tostringtag) |
| `follow-redirects` | 1.15.6 | MIT | [link](https://github.com/follow-redirects/follow-redirects) |
| `form-data` | 4.0.4 | MIT | [link](https://github.com/form-data/form-data) |
| `function-bind` | 1.1.2 | MIT | [link](https://github.com/Raynos/function-bind) |
| `get-intrinsic` | 1.3.0 | MIT | [link](https://github.com/ljharb/get-intrinsic) |
| `get-proto` | 1.0.1 | MIT | [link](https://github.com/ljharb/get-proto) |
| `gopd` | 1.2.0 | MIT | [link](https://github.com/ljharb/gopd) |
| `has-symbols` | 1.1.0 | MIT | [link](https://github.com/inspect-js/has-symbols) |
| `has-tostringtag` | 1.0.2 | MIT | [link](https://github.com/inspect-js/has-tostringtag) |
| `hasown` | 2.0.2 | MIT | [link](https://github.com/inspect-js/hasOwn) |
| `luxon` | 3.7.1 | MIT | [link](https://github.com/moment/luxon) |
| `math-intrinsics` | 1.1.0 | MIT | [link](https://github.com/es-shims/math-intrinsics) |
| `mime-db` | 1.52.0 | MIT | [link](https://github.com/jshttp/mime-db) |
| `mime-types` | 2.1.35 | MIT | [link](https://github.com/jshttp/mime-types) |
| `nanoid` | 3.3.7 | MIT | [link](https://github.com/ai/nanoid) |
| `picocolors` | 1.1.1 | ISC | [link](https://github.com/alexeyraspopov/picocolors) |
| `postcss` | 8.4.49 | MIT | [link](https://github.com/postcss/postcss) |
| `prettier` | 2.8.8 | MIT | [link](https://github.com/prettier/prettier) |
| `proxy-from-env` | 1.1.0 | MIT | [link](https://github.com/Rob--W/proxy-from-env) |
| `sortablejs` | 1.10.2 | MIT | [link](https://github.com/SortableJS/Sortable) |
| `source-map` | 0.6.1 | BSD-3-Clause | [link](https://github.com/mozilla/source-map) |
| `source-map-js` | 1.2.1 | BSD-3-Clause | [link](https://github.com/7rulnik/source-map-js) |
| `vue` | 2.7.16 | MIT | [link](https://github.com/vuejs/vue) |
| `vue-chartjs` | 4.1.2 | MIT | [link](https://github.com/apertureless/vue-chartjs) |
| `vue-codemirror` | 4.0.6 | MIT | [link](https://github.com/surmon-china/vue-codemirror) |
| `vue-i18n` | 8.28.2 | MIT | [link](https://github.com/kazupon/vue-i18n) |
| `vue-router` | 3.6.5 | MIT | [link](https://github.com/vuejs/vue-router) |
| `vue-virtual-scroll-list` | 2.3.5 | MIT | [link](https://github.com/tangbc/vue-virtual-scroll-list) |
| `vuedraggable` | 2.24.3 | MIT | [link](https://github.com/SortableJS/Vue.Draggable) |
| `vuetify` | 2.7.2 | MIT | [link](https://github.com/vuetifyjs/vuetify) |

---

## License texts

Full license texts for each component are available at the source URLs listed above. For permissively-licensed packages (MIT, BSD, ISC, Apache-2.0), the original LICENSE and NOTICE files are preserved in their respective package directories within the Semaphore UI distribution.

If you believe a component is missing from this list or incorrectly attributed, please open an issue at https://github.com/semaphoreui/semaphore/issues.

<!-- end of generated file -->
