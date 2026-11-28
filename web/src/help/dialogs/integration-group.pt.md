---
form: src/components/IntegrationGroupForm.vue, src/components/IntegrationAuthForm.vue
---

# Editar integração

Uma integração é um conjunto de execuções que se iniciam a partir de um endpoint. Quando um pedido a
um URL desse endpoint passa a autorização, inicia-se cada execução da integração cujos matchers
aceitam o pedido.

## Nome {#integrationGroupForm-name}

O nome com que a integração aparece na página «Integrações». O campo é obrigatório.

## Endpoint {#integrationGroupForm-endpointId}

O endpoint cujos URL iniciam as execuções desta integração. Várias integrações podem usar o mesmo
endpoint. O campo é obrigatório.

## Substituir autorização {#integrationGroupForm-authOverride}

Deixe desativado para verificar os pedidos como o endpoint o faz. Ative para verificar os pedidos
desta integração com um método e um segredo próprios, definidos nos campos que aparecem abaixo; as
definições do endpoint deixam então de se aplicar a esta integração.

## Método de autenticação {#integrationAuthForm-authMethod}

Como o Semaphore verifica que um pedido provém do serviço esperado. Um pedido que falha a
verificação não inicia nada, e o webhook responde-lhe da mesma forma que a qualquer outro pedido.

- **Nenhum**: sem verificação. Qualquer pessoa que conheça o URL pode iniciar a tarefa.
- **GitHub Webhooks**: o cabeçalho `X-Hub-Signature-256` tem de conter a assinatura HMAC-SHA256
  do corpo feita com o segredo, tal como o GitHub a envia.
- **Bitbucket Webhooks**: a mesma verificação no cabeçalho `X-Hub-Signature`, tal como o
  Bitbucket o envia.
- **Token**: o cabeçalho indicado em **Cabeçalho de autenticação** tem de conter o próprio segredo.
- **HMAC** e **HMAC (SHA-512)**: o cabeçalho indicado em **Cabeçalho de autenticação** tem de conter o
  HMAC-SHA256 ou o HMAC-SHA512 do corpo feito com o segredo, em hexadecimal minúsculo, sem um
  prefixo como `sha256=`.
- **BasicAuth**: o pedido tem de levar, no seu cabeçalho `Authorization`, o nome de utilizador e
  a palavra-passe da chave indicada em **Segredo de autenticação**.

## Cabeçalho de autenticação {#integrationAuthForm-authHeader}

Apresentado quando **Método de autenticação** é **Token**, **HMAC** ou **HMAC (SHA-512)**.

O nome do cabeçalho do pedido que leva o token ou a assinatura, por exemplo `X-Webhook-Token`.
As maiúsculas e minúsculas do nome não importam. Um pedido sem esse cabeçalho é recusado.

## Segredo de autenticação {#integrationAuthForm-authSecretId}

Apresentado quando **Método de autenticação** não é **Nenhum**.

A chave do Armazenamento de Chaves com que o pedido é verificado. Para **BasicAuth** só serve uma
chave do tipo **Login com senha**. Para os outros métodos serve também uma chave do tipo
**Texto**: a verificação utiliza o seu valor, ou a palavra-passe de uma chave **Login com senha**.

O campo é necessário para todos os métodos exceto **Nenhum**: sem uma chave, ou com um segredo
vazio, todos os pedidos são recusados.
