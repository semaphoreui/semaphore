---
form: src/components/IntegrationGroupForm.vue, src/components/IntegrationAuthForm.vue
---

# Editar integração

Uma integração é um conjunto de execuções iniciadas a partir de um endpoint. Quando uma requisição
para uma URL desse endpoint passa pela autorização, é iniciada cada execução da integração cujos
matchers aceitam a requisição.

## Nome {#integrationGroupForm-name}

O nome com que a integração aparece na página “Integrações”. O campo é obrigatório.

## Endpoint {#integrationGroupForm-endpointId}

O endpoint cujas URLs iniciam as execuções desta integração. Várias integrações podem usar o mesmo
endpoint. O campo é obrigatório.

## Sobrescrever autorização {#integrationGroupForm-authOverride}

Deixe desativado para verificar as requisições da mesma forma que o endpoint. Ative para verificar
as requisições desta integração com um método e um segredo próprios, definidos nos campos que
aparecem abaixo; as configurações do endpoint deixam então de valer para esta integração.

## Método de autenticação {#integrationAuthForm-authMethod}

Como o Semaphore verifica que uma requisição vem do serviço que você espera. Uma requisição que
falha na verificação não inicia nada, e o webhook a responde da mesma forma que qualquer outra
requisição.

- **Nenhum**: nenhuma verificação. Qualquer pessoa que conheça a URL pode iniciar a tarefa.
- **GitHub Webhooks**: o cabeçalho `X-Hub-Signature-256` precisa conter a assinatura
  HMAC-SHA256 do corpo feita com o segredo, como o GitHub a envia.
- **Bitbucket Webhooks**: a mesma verificação no cabeçalho `X-Hub-Signature`, como o Bitbucket a
  envia.
- **Token**: o cabeçalho indicado em **Cabeçalho de autenticação** precisa conter o próprio segredo.
- **HMAC** e **HMAC (SHA-512)**: o cabeçalho indicado em **Cabeçalho de autenticação** precisa conter o
  HMAC-SHA256 ou o HMAC-SHA512 do corpo feito com o segredo, em hexadecimal minúsculo, sem um
  prefixo como `sha256=`.
- **BasicAuth**: a requisição precisa levar o login e a senha da chave de
  **Segredo de autenticação** no seu cabeçalho `Authorization`.

## Cabeçalho de autenticação {#integrationAuthForm-authHeader}

Exibido quando **Método de autenticação** é **Token**, **HMAC** ou **HMAC (SHA-512)**.

O nome do cabeçalho da requisição que leva o token ou a assinatura, por exemplo
`X-Webhook-Token`. Maiúsculas e minúsculas no nome não importam. Uma requisição sem esse
cabeçalho é recusada.

## Segredo de autenticação {#integrationAuthForm-authSecretId}

Exibido quando **Método de autenticação** não é **Nenhum**.

A chave do Armazenamento de Chaves com a qual a requisição é verificada. Para **BasicAuth**, serve
apenas uma chave do tipo **Login com senha**. Para os outros métodos, uma chave do tipo **Texto**
também serve: a verificação usa o valor dela, ou a senha de uma chave **Login com senha**.

O campo é necessário para todos os métodos, exceto **Nenhum**: sem uma chave, ou com um segredo
vazio, toda requisição é recusada.
