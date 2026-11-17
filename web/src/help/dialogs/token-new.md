---
form: src/views/Tokens.vue
---

# New Token

An API token lets a script or another service call the Semaphore API as you, with the same
rights as your account. Send it in the `Authorization` header of a request: `Bearer` and the
token. The token is shown once, right after it is created: copy it then.

## Token Name {#token-name}

Optional. A name that tells the token apart in the list of tokens, for example the name of the
script that uses it. It is not a part of the token and does not change what the token may do.

## Expires {#token-expires}

When the token stops working. Choose a period counted from now, **Never** for a token that works
until it is deleted, or **Custom** to pick the moment yourself.

A token that never expires works until you delete it from the list. Prefer a period for a token
given to a script or to another service.

## Expiration date {#token-expiresAt}

Shown when **Expires** is **Custom**.

The date and time the token stops working, in the time zone of your browser. The moment must be
in the future: the server refuses a token that would be expired already.
