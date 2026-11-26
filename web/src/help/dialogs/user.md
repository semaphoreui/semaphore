---
form: src/components/UserForm.vue, src/components/GlobalRoleAssignments.vue
---

# Edit User

A user is an account that signs in to Semaphore. The **Settings** tab holds the profile of the
account and its flags; **Save** writes them. The **Security** tab holds the password, two-factor
authentication and the linked sign-in accounts, and the **Roles** tab the global roles of the
user: both act at once, without **Save**. What a user may do inside a project is set on the
**Team** page of that project, not here.

## Name {#user-name}

The full name of the user, shown in the side bar and in the lists of users and of project
members. The field is required.

For an **External** user, Semaphore takes the name from the identity provider at sign-in, when
the provider sends one, so a change made here may not last past the next sign-in.

## Username {#user-username}

The login of the user. The field is required. The user signs in with the username or with the
email, and the password.

The username of an **External** user can not be changed: the server refuses to save it.

## Email {#user-email}

The email address of the user. The field is required. Semaphore sends the task alerts of the
user to it, and an authenticator app shows it as the name of the account. It is private: only
the user and those who may manage users see it.

For an **External** user, the email comes from the identity provider: the field is read-only, and
Semaphore updates the address at every sign-in.

## Password {#user-password}

Shown when a new user is created.

The password the user signs in with. The field is required, except for an **External** user, who
has no password in Semaphore. To change the password of an existing user, use
**Change password** on the **Security** tab.

## Send alerts {#user-alert}

Sends the user an email when a task fails in a project where the user may see the task. The
project must allow alerts, and the administrator of the server must have turned email alerts
on. A task template can keep its failed tasks from sending alerts.

The flag does not affect the alerts of a project sent to Telegram, Slack and the other
messengers.

## Admin user {#user-admin}

Shown when you are an administrator.

An administrator may do everything on the server: manage the users, the global roles and the
settings of the server, and open every project with full rights.

You can not take the flag from your own account, and the server keeps at least one
administrator: it refuses to take the flag from the last one.

## Pro user {#user-pro}

Shown in the Pro edition. Only an administrator can change the flag.

Gives the user a seat of the Pro subscription. The subscription has a fixed number of Pro users:
when it is reached, the server refuses to save the flag. Without an active subscription the flag
has no effect.

## External {#user-external}

Shown when you are an administrator. It can be set only when the user is created.

An external user signs in through LDAP or OpenID Connect, not with a password of Semaphore: the
password field is disabled, and a sign-in with a password is refused. The name and the email are
taken from the identity provider at every sign-in. Leave the box empty for a user who signs in
with a password.

## Change password {#user-changePassword}

Shown on the **Security** tab of an existing user who is not **External**, when the server allows
signing in with a password.

Opens a dialog that sets a new password. To change your own password, you have to give the
current one; to reset the password of another user, you do not.

## Time-based one-time password {#user-totp}

Shown on the **Security** tab when the administrator has turned two-factor authentication on for
the server.

Turns two-factor authentication on for this user. When you turn the switch on, Semaphore
creates a secret at once and shows it as a QR code: scan it with an authenticator app. From then
on, a sign-in asks for the code of the app after the password.

When the server allows recovery, a recovery code is shown under the QR code. It is shown only
once: copy it and keep it safe. Turning the switch off removes the secret at once.

## LDAP provider {#user-ldapProvider}

Shown on the **Security** tab of your own account when the server has an LDAP provider your
account is not linked to yet.

The LDAP server to link your account to, when there are several.

## LDAP username {#user-ldapUsername}

Shown with **LDAP provider**.

Your username on the LDAP server. Linking lets you sign in to this account through LDAP as well.

## LDAP password {#user-ldapPassword}

Shown with **LDAP provider**.

Your password on the LDAP server. Semaphore uses it once, to check that the LDAP account is
yours, and does not store it.

## Link {#user-ldapLink}

Shown with **LDAP provider**.

Links your account to the LDAP account above. The link is refused when the credentials are
wrong, when the LDAP account is linked to another user, or when your account is linked to that
provider already. A linked account appears under **Linked accounts**; close its chip to unlink
it.

## Global role {#global-role-select}

Shown on the **Roles** tab of an existing user, in the Pro edition, when you may manage global
roles.

A global role to give the user, out of the roles the user does not hold yet. A global role
grants permissions over the whole server: to manage users, global roles or the settings of the
server, or to read the audit log. The global roles themselves are created on the **Global roles**
tab of the Users page.

## Assign {#globalRoleAssign-assign}

Shown with **Global role**.

Gives the user the selected role at once. The roles the user holds are listed below, each with a
button to take it away. **Effective global permissions** shows what the user may do in the end,
and where each permission comes from.
