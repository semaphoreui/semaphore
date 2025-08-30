# Example configuration to disable email login with OIDC

This example shows how to configure Semaphore to disable email login when OIDC providers are configured.

## Configuration

Add the following to your `config.json`:

```json
{
  "password_login_disable": true,
  "email_login_disable_with_oidc": true,
  "oidc_providers": {
    "authentik": {
      "display_name": "Sign in with Authentik",
      "provider_url": "https://auth.your-domain.com/application/o/semaphore/",
      "client_id": "your-client-id",
      "client_secret": "your-client-secret",
      "redirect_url": "https://semaphore.your-domain.com/api/auth/oidc/authentik/redirect/",
      "scopes": ["openid", "profile", "email"],
      "username_claim": "preferred_username",
      "name_claim": "preferred_username"
    }
  }
}
```

## Environment Variables

Alternatively, you can use environment variables:

```bash
SEMAPHORE_PASSWORD_LOGIN_DISABLED=true
SEMAPHORE_EMAIL_LOGIN_DISABLE_WITH_OIDC=true
```

## Behavior

- When `email_login_disable_with_oidc` is `true` and OIDC providers are configured, the email login form will be hidden
- When only one OIDC provider is configured and both password and email login are disabled, users will be automatically redirected to the OIDC provider
- When multiple OIDC providers are configured, users will see buttons for each provider without the email login option

## Use Cases

This feature is useful when:
- You want to enforce authentication through your organization's SSO/OIDC provider
- You want to prevent unauthorized users from attempting email-based login
- You want to simplify the login experience by removing unnecessary options