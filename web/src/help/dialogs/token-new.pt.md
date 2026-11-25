---
form: src/views/Tokens.vue
---

# New Token

Um token de API permite que um script ou outro serviço chame a API do Semaphore em seu nome, com
os mesmos direitos da sua conta. Envie-o no cabeçalho `Authorization` de um pedido: `Bearer`
seguido do token. O token é mostrado uma única vez, logo depois de ser criado: copie-o nesse
momento.

## Token Name {#token-name}

Opcional. Um nome que distingue o token na lista de tokens, por exemplo o nome do script que o
utiliza. Não faz parte do token e não altera o que o token pode fazer.

## Expires {#token-expires}

Quando o token deixa de funcionar. Escolha um período contado a partir de agora, **Never** para
um token que funciona até ser eliminado, ou **Custom** para escolher o momento por si.

Um token que nunca expira funciona até o eliminar da lista. Prefira um período para um token
entregue a um script ou a outro serviço.

## Expiration date {#token-expiresAt}

Apresentado quando **Expires** é **Custom**.

A data e a hora em que o token deixa de funcionar, no fuso horário do seu navegador. O momento
tem de estar no futuro: o servidor recusa um token que já estaria expirado.
