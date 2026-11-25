---
form: src/components/IntegrationMatcherForm.vue
---

# Edit Matcher

Um correspondente é uma condição sobre o pedido recebido. Os correspondentes decidem que
integração um pedido ao alias do projeto inicia, para que um único URL possa servir várias
integrações: uma integração só é executada quando todos os seus correspondentes se verificam. São
utilizados apenas quando **Usar alias do projeto** está ativado na integração; com essa opção
ativada, a integração só é alcançada através do alias do projeto.

Uma integração encontrada pelo alias do projeto sem correspondentes nunca é executada: adicione
pelo menos um.

## Name {#integrationMatcherForm-name}

O nome com que o correspondente aparece na página da integração. Não altera o que o
correspondente verifica. O campo é obrigatório.

## Match on {#integrationMatcherForm-matchType}

A parte do pedido que o correspondente lê:

- **Header**: o valor de um cabeçalho do pedido.
- **Body**: o corpo do pedido.

## Body Data Format {#integrationMatcherForm-bodyDataType}

Apresentado quando **Match on** é **Body**.

Como o corpo é lido:

- **JSON**: o corpo é interpretado como JSON, e **Key** é o caminho para um valor dentro dele.
- **String**: todo o corpo é comparado como texto, e **Key** não é utilizado.

## Key {#integrationMatcherForm-key}

O que ler do pedido. O campo é obrigatório.

- Para **Header**: o nome do cabeçalho, por exemplo `X-GitHub-Event`. As maiúsculas e minúsculas
  não importam.
- Para **Body** em **JSON**: o caminho para o valor, com as suas partes unidas por pontos; um
  elemento de um array escreve-se `.[0]`, por exemplo `repository.full_name` ou
  `commits.[0].id`.
- Para **Body** em **String**: não é utilizado, mas o formulário precisa de um valor; escreva
  qualquer coisa.

## Comparison Method {#integrationMatcherForm-method}

Como o valor lido do pedido é comparado com **Value**:

- **==**: os dois são iguais.
- **!=**: os dois são diferentes. Um cabeçalho ou uma chave que falta no pedido conta como
  diferente.
- **Contains**: o valor lido do pedido contém **Value**.

## Value {#integrationMatcherForm-value}

O texto com que comparar, por exemplo `push` ou `refs/heads/main`. A comparação é exata e
distingue maiúsculas de minúsculas. Um número ou um booleano do corpo JSON é comparado na sua
forma de texto: `42`, `true`. O campo é obrigatório.
