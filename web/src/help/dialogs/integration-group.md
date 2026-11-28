---
form: src/components/IntegrationGroupForm.vue, src/components/IntegrationAuthForm.vue
---

# Edit Integration

An integration is a set of runs that start from one endpoint. When a request to a URL of that
endpoint passes the authorization, every run of the integration whose matchers accept the request
starts.

## Name {#integrationGroupForm-name}

The name the integration is listed under on the Integrations page. The field is required.

## Endpoint {#integrationGroupForm-endpointId}

The endpoint whose URLs start the runs of this integration. Several integrations can use one
endpoint. The field is required.

## Override endpoint authorization {#integrationGroupForm-authOverride}

Leave it off to check requests the way the endpoint does. Turn it on to check the requests of this
integration with a method and a secret of its own, set in the fields that appear below; the settings
of the endpoint then do not apply to this integration.

## Auth method {#integrationAuthForm-authMethod}

How Semaphore checks that a request comes from the service you expect. A request that fails the
check starts nothing, and the webhook answers it the same as any other request.

- **None**: no check. Anyone who knows the URL can start the task.
- **GitHub Webhooks**: the `X-Hub-Signature-256` header must hold the HMAC-SHA256 signature of
  the body made with the secret, as GitHub sends it.
- **Bitbucket Webhooks**: the same check on the `X-Hub-Signature` header, as Bitbucket sends it.
- **Token**: the header in **Auth header** must hold the secret itself.
- **HMAC** and **HMAC (SHA-512)**: the header in **Auth header** must hold the HMAC-SHA256 or
  HMAC-SHA512 of the body made with the secret, in lowercase hexadecimal, without a prefix
  such as `sha256=`.
- **BasicAuth**: the request must carry the login and the password of the key in
  **Auth secret** in its `Authorization` header.

## Auth header {#integrationAuthForm-authHeader}

Shown when **Auth method** is **Token**, **HMAC** or **HMAC (SHA-512)**.

The name of the request header that carries the token or the signature, for example
`X-Webhook-Token`. The case of the name does not matter. A request without that header is
refused.

## Auth secret {#integrationAuthForm-authSecretId}

Shown when **Auth method** is not **None**.

The key from the Key Store the request is checked against. For **BasicAuth** only a key of the
type **Login with password** fits. For the other methods a key of the type **String** fits too:
the check uses its value, or the password of a **Login with password** key.

The field is needed by every method but **None**: without a key, or with an empty secret, every
request is refused.
