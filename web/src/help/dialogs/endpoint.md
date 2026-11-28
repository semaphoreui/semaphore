---
form: src/components/EndpointForm.vue, src/components/IntegrationAuthForm.vue
---

# Edit Endpoint

An endpoint is an entry point for webhooks: the URLs other services call, and the way their requests
are checked. Integrations attach to an endpoint and start their runs on the requests it accepts. The
URLs of an endpoint, its aliases, are added on the Endpoints tab.

## Name {#endpointForm-name}

The name the endpoint is listed under on the Endpoints tab and in the form of an integration. The
field is required.

## Description {#endpointForm-description}

Optional. A note on what the endpoint is for, shown next to its name.

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
