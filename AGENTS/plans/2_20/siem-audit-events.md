# Plan — SIEM-Ready Audit Events

- **Branch:** `develop`
- **Research:** MCP research «Поддержка SIEM в Semaphore UI» (RESEARCH@a39c0ef615f486972eaff4),
  16.07.2026: анализ кодовой базы, GitHub issues (#158, discussion #2194, #3410),
  конкурентов (AWX External Log Aggregator, Rundeck Audit Stream Plugin) и
  требований OWASP Logging Vocabulary / SIEM ingestion (Splunk HEC, syslog RFC 5424, CEF).

## 1. Problem

Semaphore имеет событийный лог (таблица `event`, страница Activity, API
`/events`), но он не пригоден как audit trail для SIEM/compliance:

1. **Логин/логаут и неудачные попытки входа не логируются вообще** — ни в
   таблицу `event`, ни в файловый лог. В `api/login.go` только одна
   info-строка logrus при успешной LDAP-аутентификации. Это требование №1
   OWASP и любого SOC.
2. **IP-адрес и user-agent не сохраняются** ни в `db.Event`, ни в
   `pro_interfaces.EventLogRecord` — хотя при создании сессии они уже
   извлекаются из запроса (`api/login.go:172-173`: `X-Real-IP`,
   `user-agent`).
3. **Поле `action` (create/update/delete) не попадает в БД** — только в
   logrus-поля и файловый лог (`api/helpers/event_log.go:48-62`). В БД
   событие различимо только по тексту `Description`.
4. **Глобальный CRUD пользователей и API-токенов не логируется** —
   `api/users.go` (AddUser/UpdateUser/UpdateUserPassword/DeleteUser) и
   `api/user.go:129-171` (create/deleteAPIToken) не вызывают
   `helpers.EventLog`. Логируются только членство/роли внутри проекта.
5. **`integration_id` теряется**: есть в модели, но пропущен в INSERT
   (`db/sql/event.go:36`).
6. **Нет push-канала аудита**: syslog (`cli/cmd/syslog.go`) пересылает весь
   logrus-лог вперемешку с отладкой; generic outbound HTTP-вебхука
   (Splunk HEC и т.п.) нет; алерты (`services/tasks/alert.go`) — только о
   статусе задач.

## 2. Current State (для контекста исполнителя)

- `db/Event.go:10-24` — модель `Event`; `db/sql/event.go:32-51` — INSERT.
- `api/helpers/event_log.go:29` — `EventLog(r *http.Request, action EventLogType, item EventLogItem)`:
  единая точка записи; пишет в БД и в `pro_interfaces.LogWriteService.WriteEventLog`.
- `pro_interfaces/log_write_svc.go:5-17` — интерфейс `LogWriteService` и
  `EventLogRecord`. OSS-реализация — заглушка (`pro/services/server/log_write_svc.go`),
  рабочая (JSON/raw файл + lumberjack) — в `pro_impl/services/server/log_write_svc.go`.
- `util/config.go:290-311` — `EventLogType`/`TaskLogType` (env
  `SEMAPHORE_EVENT_LOG_*`); `util/config.go:320-326` — `SyslogConfig`.
- Миграции: `db/sql/migrations/v2.20.0.sql`, регистрация в
  `db/Migration.go:GetMigrations`.

## 3. Design

Принцип (стандарт де-факто по AWX/Rundeck + OWASP): **единый словарь
аудит-событий → одна точка эмиссии → несколько эмиттеров**.

### 3.1 Словарь событий

Действия расширяются с create/update/delete до:

```go
// api/helpers/event_log.go
const (
    EventLogCreate EventLogType = "create"
    EventLogUpdate EventLogType = "update"
    EventLogDelete EventLogType = "delete"

    EventLogLoginSuccess EventLogType = "login_success"
    EventLogLoginFail    EventLogType = "login_fail"
    EventLogLogout       EventLogType = "logout"
)
```

Новый тип объекта: `EventSession EventObjectType = "session"` (в
`db/Event.go` рядом с остальными константами).

### 3.2 Обогащение Event

`db.Event` получает `Action`, `IP`, `UserAgent`; всё заполняется в
`helpers.EventLog` из `*http.Request` — вызывающий код не меняется.

### 3.3 OSS / Pro split (продуктовое решение)

Следуем модели Rundeck и текущему split'у репозитория:

- **OSS**: обогащённые события в БД (action/IP/UA), auth-события,
  user/token CRUD. Это закрывает доверие/basics.
- **Pro** (`pro_impl/`): файловый JSON-лог уже есть; добавляется
  **generic HTTP-вебхук аудита** (совместимый со Splunk HEC) —
  асинхронный, с ретраями, отказ доставки не влияет на запрос.
- Syslog RFC 5424 уже есть в OSS для всего лога; отдельный audit-канал в
  syslog не делаем в этой версии (события и так попадают в logrus-поток
  через `event.ToFields()` — этого достаточно; выделенный канал/CEF —
  кандидат на следующую версию, если будет спрос).

## 4. Tasks

### Task 1 — Миграция и модель: action, ip, user_agent в event

**Files:**
- Create: `db/sql/migrations/v2.20.1.sql`
- Modify: `db/Migration.go` (добавить `{Version: "2.20.1"}` в список),
  `db/Event.go`, `db/sql/event.go` (bolt-хранилища в репо нет — только SQL)

**Миграция** (`v2.20.1.sql`):

```sql
alter table event add `action` varchar(20) null;
alter table event add `ip` varchar(45) null;
alter table event add `user_agent` varchar(255) null;
```

**Модель** (`db/Event.go`, добавить в `Event`):

```go
Action    *string `db:"action" json:"action"`
IP        *string `db:"ip" json:"ip"`
UserAgent *string `db:"user_agent" json:"user_agent"`
```

`db/sql/event.go:CreateEvent` — расширить INSERT новыми колонками и
заодно починить потерю `integration_id`:

```go
_, err = d.exec(
    "insert into event(user_id, project_id, integration_id, object_id, object_type, description, created, action, ip, user_agent) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
    evt.UserID, evt.ProjectID, evt.IntegrationID, evt.ObjectID, evt.ObjectType,
    evt.Description, created, evt.Action, evt.IP, evt.UserAgent)
```

**Test:** `db/sql/event_test.go` — CreateEvent сохраняет и возвращает
action/ip/user_agent и integration_id (по образцу существующих sql-тестов;
testify, `require.NoError`).

### Task 2 — helpers.EventLog: заполнение новых полей из запроса

**Files:**
- Modify: `api/helpers/event_log.go`, `pro_interfaces/log_write_svc.go`,
  `pro/services/server/log_write_svc.go` (заглушка — сигнатуры не меняются),
  `pro_impl/services/server/log_write_svc.go` (запись новых полей в JSON/raw)

В `EventLog` перед записью:

```go
func extractClientIP(r *http.Request) string {
    if ip := r.Header.Get("X-Real-IP"); ip != "" {
        return ip
    }
    host, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        return r.RemoteAddr
    }
    return host
}
```

```go
actionStr := string(action)
ip := extractClientIP(r)
ua := r.Header.Get("user-agent")
event.Action = &actionStr
event.IP = &ip
event.UserAgent = &ua
```

`EventLogRecord` (`pro_interfaces/log_write_svc.go`) дополнить:

```go
IP        string `json:"ip,omitempty"`
UserAgent string `json:"user_agent,omitempty"`
ObjectType *string `json:"object_type,omitempty"`
ObjectID   *int    `json:"object_id,omitempty"`
```

(сейчас в файловый лог не попадают даже object_type/object_id — добавить.)

**Test:** `api/helpers/event_log_test.go` — httptest-запрос с `X-Real-IP`
и `User-Agent`; mock store; проверить, что `CreateEvent` получил
заполненные Action/IP/UserAgent.

### Task 3 — Auth-события: login/logout/fail

**Files:**
- Modify: `api/login.go`, `db/Event.go` (константа `EventSession`),
  `api/helpers/event_log.go` (константы login_success/login_fail/logout)

Точки вставки:

1. `login(w, r)` (`api/login.go:292`) — после успешной аутентификации
   (перед/после `createSession`):
   ```go
   helpers.EventLog(r, helpers.EventLogLoginSuccess, helpers.EventLogItem{
       UserID:      user.ID,
       ObjectType:  db.EventSession,
       ObjectID:    user.ID,
       Description: fmt.Sprintf("User %s logged in", user.Username),
   })
   ```
   На каждой ветке отказа (неверный пароль, user not found, LDAP fail) —
   `EventLogLoginFail` с login из запроса в Description. **Не логировать
   пароль.** Для "user not found" UserID остаётся 0 (не пишется).
2. `logout` (`api/login.go:455`) — `EventLogLogout`.
3. OIDC-callback (обработчик после `oidcLogin`/redirect, там где вызывается
   `createSession` для OIDC-пользователя) — те же login_success/login_fail.
4. Неудачная TOTP-проверка (обработчик verify в `api/auth.go` /
   `api/login.go`, найти по `SessionVerificationTotp`) — `EventLogLoginFail`
   с Description "MFA verification failed".

Замечание: события уровня инстанса (ProjectID == 0) не видны не-админам в
`/events` — это корректно (`api/events.go:21-23`).

**Test:** `api/login_test.go` — существующие тестовые хелперы логина;
проверить, что после неверного пароля в store появилось событие с
action=login_fail и IP.

### Task 4 — События глобального CRUD пользователей и API-токенов

**Files:**
- Modify: `api/users.go` (AddUser, UpdateUser, UpdateUserPassword,
  DeleteUser, DeleteUserIdentity), `api/user.go` (createAPIToken,
  deleteAPIToken)
- Modify: `db/Event.go` — новый тип `EventAPIToken EventObjectType = "api_token"`

Образец (AddUser, после успешного создания):

```go
helpers.EventLog(r, helpers.EventLogCreate, helpers.EventLogItem{
    UserID:      editor.ID, // кто совершил действие
    ObjectType:  db.EventUser,
    ObjectID:    newUser.ID,
    Description: fmt.Sprintf("User %s created", newUser.Username),
})
```

- UpdateUserPassword → Description "Password changed for user %s"
  (сам пароль — никогда).
- createAPIToken/deleteAPIToken → ObjectType `api_token`, Description без
  значения токена, только ID.

**Test:** дополнить `api/users_test.go` (или создать) — AddUser пишет
событие create с ObjectType user.

### Task 5 (Pro) — Audit webhook emitter (generic HTTP / Splunk HEC)

**Files (в `pro_impl/`, отдельный репо!):**
- Create: `pro_impl/services/server/audit_webhook_svc.go` + `_test.go`
- Modify: `pro_impl/services/server/log_write_svc.go` — после записи в
  файл отдать запись в webhook-эмиттер
- Modify (OSS): `util/config.go` — конфиг:

```go
// util/config.go, в ConfigLog:
type AuditWebhookConfig struct {
    Enabled  bool              `json:"enabled" env:"SEMAPHORE_AUDIT_WEBHOOK_ENABLED"`
    URL      string            `json:"url" env:"SEMAPHORE_AUDIT_WEBHOOK_URL"`
    // Заголовки авторизации, например:
    //   Authorization: "Splunk <hec-token>"  → Splunk HEC
    //   Authorization: "Bearer <token>"      → generic
    Headers  map[string]string `json:"headers" env:"SEMAPHORE_AUDIT_WEBHOOK_HEADERS"`
    // splunk_hec | json (default json)
    Format   string            `json:"format" env:"SEMAPHORE_AUDIT_WEBHOOK_FORMAT"`
}
```

Требования к эмиттеру (из research, стандарт AWX/Rundeck):

- Асинхронно: буферизованный канал (ёмкость ~1000) + одна горутина-отправитель;
  `EventLog` никогда не блокируется и не возвращает ошибку доставки.
- Ретраи: 3 попытки с backoff (1s/5s/30s); после — запись теряется с
  logrus-warn. `// ponytail: in-memory буфер, дисковая очередь — если попросят`.
- Формат `splunk_hec`: конверт `{"time": <unix>, "event": {...}, "sourcetype": "semaphore:audit"}`
  POST на `<url>/services/collector/event`.
- Формат `json`: POST записи as-is.
- Санитизация: CR/LF в Description заменяются пробелом.
- Graceful shutdown: дослать буфер при остановке сервера (с таймаутом 5s).

**Test:** httptest.Server как приёмник; проверить конверт HEC, ретрай при
500, отсутствие блокировки при недоступном приёмнике.

### Task 6 — Документация и схема конфига

**Files:**
- Modify: `config.schema.yaml` — новые поля `log.audit_webhook.*`
  (использовать skill `semaphore-config-schema`)
- Modify: `api-docs.yml` / swagger-описание Event (новые поля action/ip/user_agent)
- Modify: docs `admin-guide` — страница про SIEM-интеграцию: таблица
  «канал → формат → как подключить Splunk/Elastic/Wazuh»
- Frontend: `web/src/views/project/Activity.vue` (или где рендерится
  Activity) — показать action и IP в списке событий (опционально, можно
  отдельной задачей)

## 5. Out of Scope (осознанно, кандидаты на 2.21+)

- CEF/LEEF-форматтер и выделенный audit-канал в syslog — только при спросе
  (QRadar-клиенты).
- Before/after (diff) в событиях изменения объектов.
- Дисковая очередь для вебхука (сейчас — in-memory буфер).
- Retention/архивация таблицы `event`.
- События доступа к секретам на чтение (secret access trail) — в roadmap
  Enterprise (#3410-смежное), отдельный план.

## 6. Порядок и зависимости

Task 1 → Task 2 → (Task 3, Task 4 — независимы, параллельно) → Task 5
(зависит от Task 2 по полям EventLogRecord) → Task 6.

Каждая задача — отдельный коммит(ы) `feat(audit): ...`, тесты по
правилам `.claude/CLAUDE.md` (testify, таблично-управляемые где уместно).
