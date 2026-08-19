# TC-001 — Admin login with valid credentials

| Field        | Value                            |
|--------------|----------------------------------|
| Area         | Authentication                   |
| Priority     | Critical                         |
| Type         | Functional / Positive            |
| Automatable  | Yes (Playwright)                 |

## Objective

Verify that the bootstrap admin user can sign in to the web UI and lands on the
project selection screen with a valid session.

## Preconditions

* Semaphore instance freshly started with environment variables
  `SEMAPHORE_ADMIN=admin`, `SEMAPHORE_ADMIN_PASSWORD=changeme`,
  `SEMAPHORE_ADMIN_NAME=Admin`, `SEMAPHORE_ADMIN_EMAIL=admin@localhost`.
* The instance is reachable at `http://localhost:3000`.
* No browser cookies/local storage from a prior session.

## Test data

* Username: `admin`
* Password: `changeme`

## Steps

1. Open `http://localhost:3000` in a clean browser session.
2. On the login page enter `admin` as the username/email.
3. Enter `changeme` as the password.
4. Click **Sign in**.
5. Observe the destination page.
6. Open browser dev tools → Application → Cookies and inspect the session cookie.

## Expected results

* The user is redirected to `/projects` (or the new project wizard when no
  project exists) within 2 seconds.
* The top navigation shows the admin avatar and the "Settings"/"Users" menu
  entries that are admin-only.
* A `semaphore` session cookie is present, has `HttpOnly` set, and `Secure` when
  served over HTTPS.
* `GET /api/user` returns 200 with `"admin": true` and the configured email.

## Postconditions

Logged-in admin session usable by subsequent test cases.
