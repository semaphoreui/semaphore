---
form: src/components/ChangePasswordForm.vue
---

# Change password

Sets a new password for signing in to Semaphore with a username and a password. The new password
takes effect when you press **Save**. An external user, who signs in through LDAP or OpenID
Connect, has no password here to change.

## Current password {#changePassword-currentPassword}

Shown when you change your own password.

Your password as it is now. The field is required: Semaphore checks it before it sets the new
one, so that a stolen session or API token can not be turned into your password. The server
refuses the change when the password does not match.

An administrator or a user manager who resets the password of another user does not see this
field.

## Password {#changePassword-password}

The new password. The field is required. Click the eye icon to show what you typed.

Semaphore stores only a hash of the password.
