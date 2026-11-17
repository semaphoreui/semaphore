---
form: src/components/IntegrationMatcherForm.vue
---

# Edit Matcher

Um correspondente é uma condição sobre a requisição recebida. Os correspondentes decidem qual
integração uma requisição ao alias do projeto inicia, de modo que uma única URL pode atender a
várias integrações: uma integração é executada apenas quando todos os seus correspondentes são
satisfeitos. Eles são usados apenas quando **Usar alias do projeto** está ativado na integração;
com ele ativado, a integração só é alcançada pelo alias do projeto.

Uma integração encontrada pelo alias do projeto sem nenhum correspondente nunca é executada:
adicione pelo menos um.

## Name {#integrationMatcherForm-name}

O nome com que o correspondente aparece na página da integração. Ele não altera o que o
correspondente verifica. O campo é obrigatório.

## Match on {#integrationMatcherForm-matchType}

A parte da requisição que o correspondente lê:

- **Header**: o valor de um cabeçalho da requisição.
- **Body**: o corpo da requisição.

## Body Data Format {#integrationMatcherForm-bodyDataType}

Exibido quando **Match on** é **Body**.

Como o corpo é lido:

- **JSON**: o corpo é interpretado como JSON, e **Key** é o caminho para um valor dentro dele.
- **String**: o corpo inteiro é comparado como texto, e **Key** não é usado.

## Key {#integrationMatcherForm-key}

O que ler da requisição. O campo é obrigatório.

- Para **Header**: o nome do cabeçalho, por exemplo `X-GitHub-Event`. Maiúsculas e minúsculas não
  importam.
- Para **Body** em **JSON**: o caminho para o valor, com as partes unidas por pontos; um item de
  um array é escrito como `.[0]`, por exemplo `repository.full_name` ou `commits.[0].id`.
- Para **Body** em **String**: não é usado, mas o formulário exige um valor; digite qualquer coisa.

## Comparison Method {#integrationMatcherForm-method}

Como o valor lido da requisição é comparado com **Value**:

- **==**: os dois são iguais.
- **!=**: os dois são diferentes. Um cabeçalho ou uma chave ausente da requisição conta como
  diferente.
- **Contains**: o valor lido da requisição contém **Value**.

## Value {#integrationMatcherForm-value}

O texto com que comparar, por exemplo `push` ou `refs/heads/main`. A comparação é exata e
diferencia maiúsculas de minúsculas. Um número ou um booleano do corpo JSON é comparado na sua
forma de texto: `42`, `true`. O campo é obrigatório.
