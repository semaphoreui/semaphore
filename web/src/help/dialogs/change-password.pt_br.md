---
form: src/components/ChangePasswordForm.vue
---

# Alterar senha

Define uma nova senha para entrar no Semaphore com nome de usuário e senha. A nova senha entra em
vigor quando você clica em **Save**. Um usuário externo, que entra por LDAP ou OpenID Connect,
não tem aqui uma senha para alterar.

## Current password {#changePassword-currentPassword}

Exibido quando você altera a sua própria senha.

A sua senha atual. O campo é obrigatório: o Semaphore a verifica antes de definir a nova, para
que uma sessão ou um token de API roubado não possa ser transformado na sua senha. O servidor
recusa a alteração quando a senha não confere.

Um administrador ou um gerente de usuários que redefine a senha de outro usuário não vê este
campo.

## Senha {#changePassword-password}

A nova senha. O campo é obrigatório. Clique no ícone de olho para mostrar o que você digitou.

O Semaphore armazena apenas um hash da senha.
