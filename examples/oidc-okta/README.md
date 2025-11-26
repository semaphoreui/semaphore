# Okta OIDC Configuration for Semaphore

This example demonstrates how to configure Semaphore to use Okta as an OpenID Connect (OIDC) authentication provider.

## Problem Addressed

Some OIDC providers, including Okta, require the use of `client_secret_post` authentication method instead of the default `client_secret_basic` method. The `token_endpoint_auth_method` configuration option allows you to specify which authentication method to use.

## Configuration

Copy the `config.json` file and update the following fields:

- `YOUR_COOKIE_HASH_HERE`: Generate a secure random string (base64 encoded, 32 bytes)
- `YOUR_COOKIE_ENCRYPTION_HERE`: Generate another secure random string (base64 encoded, 32 bytes)
- `your-domain.okta.com`: Your Okta domain
- `YOUR_CLIENT_ID`: Your Okta application's Client ID
- `YOUR_CLIENT_SECRET`: Your Okta application's Client Secret
- `your-semaphore-instance.com`: Your Semaphore instance's hostname

### Token Endpoint Authentication Method

The `token_endpoint_auth_method` field supports the following values:

- `"client_secret_post"` - Sends client credentials in the request body (required by Okta)
- `"client_secret_basic"` - Sends client credentials in the Authorization header (default OAuth2 behavior)
- Empty or omitted - Auto-detect based on provider's configuration

## Okta Configuration

In your Okta application settings:

1. Set the **Sign-in redirect URIs** to: `https://your-semaphore-instance.com/api/auth/oidc/okta/redirect`
2. Set the **Client authentication** method to: **Client secret (Confidential)**
3. Ensure **Token endpoint authentication method** is set to: **Client Secret sent in HTTP POST body** (`client_secret_post`)

## Testing

After configuring, you can test the OIDC authentication by:

1. Starting Semaphore: `./semaphore server --config config.json`
2. Navigate to the login page at `https://your-semaphore-instance.com`
3. You should see an "Okta" login button
4. Click it to authenticate via Okta

## Troubleshooting

If you encounter `invalid_client` errors:

- Verify the `client_id` and `client_secret` are correct
- Ensure the `token_endpoint_auth_method` matches your Okta application's configuration
- Check that the redirect URI in Semaphore matches what's configured in Okta
- Review the Semaphore logs for detailed error messages

## Environment Variables

Alternatively, you can configure OIDC providers using environment variables:

```bash
export SEMAPHORE_OIDC_PROVIDERS='{"okta":{"display_name":"Okta","provider_url":"https://your-domain.okta.com","client_id":"YOUR_CLIENT_ID","client_secret":"YOUR_CLIENT_SECRET","redirect_url":"https://your-semaphore-instance.com/api/auth/oidc/okta/redirect","scopes":["openid","profile","email"],"token_endpoint_auth_method":"client_secret_post","username_claim":"preferred_username","name_claim":"name","email_claim":"email","order":0}}'
```
