---
form: src/components/ChangePasswordForm.vue
---

# Alterar senha

Define uma nova palavra-passe para iniciar sessão no Semaphore com um nome de utilizador e uma
palavra-passe. A nova palavra-passe produz efeito quando prime **Save**. Um utilizador externo,
que inicia sessão através de LDAP ou OpenID Connect, não tem aqui nenhuma palavra-passe para
alterar.

## Current password {#changePassword-currentPassword}

Apresentado quando altera a sua própria palavra-passe.

A sua palavra-passe tal como está agora. O campo é obrigatório: o Semaphore verifica-a antes de
definir a nova, para que uma sessão ou um token de API roubados não possam ser transformados na
sua palavra-passe. O servidor recusa a alteração quando a palavra-passe não corresponde.

Um administrador ou um gestor de utilizadores que redefine a palavra-passe de outro utilizador
não vê este campo.

## Senha {#changePassword-password}

A nova palavra-passe. O campo é obrigatório. Clique no ícone do olho para mostrar o que escreveu.

O Semaphore guarda apenas um hash da palavra-passe.
