---
form: src/views/Tokens.vue
---

# New Token

Um token de API permite que um script ou outro serviço chame a API do Semaphore em seu nome, com
os mesmos direitos da sua conta. Envie-o no cabeçalho `Authorization` de uma requisição: `Bearer`
e o token. O token é exibido uma única vez, logo depois de ser criado: copie-o nesse momento.

## Token Name {#token-name}

Opcional. Um nome que diferencia o token na lista de tokens, por exemplo o nome do script que o
usa. Ele não faz parte do token e não altera o que o token pode fazer.

## Expires {#token-expires}

Quando o token deixa de funcionar. Escolha um período contado a partir de agora, **Never** para um
token que funciona até ser excluído, ou **Custom** para escolher o momento você mesmo.

Um token que nunca expira funciona até você excluí-lo da lista. Prefira um período para um token
entregue a um script ou a outro serviço.

## Expiration date {#token-expiresAt}

Exibido quando **Expires** é **Custom**.

A data e a hora em que o token deixa de funcionar, no fuso horário do seu navegador. O momento
precisa estar no futuro: o servidor recusa um token que já estaria expirado.
