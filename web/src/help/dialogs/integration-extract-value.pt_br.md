---
form: src/components/IntegrationExtractValueForm.vue
---

# Edit Extract Value

Um valor extraído pega uma parte da requisição recebida, como o ramo de um evento de push, e a
passa para a tarefa que a integração inicia. Cada valor da integração é extraído de toda
requisição que inicia uma tarefa.

## Name {#integrationValueForm-name}

O nome com que o valor aparece na página da integração. Ele não altera o que é extraído nem para
onde vai. O campo é obrigatório.

## Source of the Value {#integrationValueForm-valueSource}

A parte da requisição de onde o valor é extraído:

- **Header**: o valor de um cabeçalho da requisição.
- **Body**: o corpo da requisição.

## Data Type of Body {#integrationValueForm-bodyDataType}

Exibido quando **Source of the Value** é **Body**.

Como o corpo é lido:

- **JSON**: o corpo é interpretado como JSON, e **Key** é o caminho para um valor dentro dele. O
  valor mantém o seu tipo JSON: um número, uma lista ou um objeto chegam à tarefa como tais.
- **String**: o corpo inteiro é extraído como texto, e **Key** não é usado.

## Key {#integrationValueForm-key}

O que extrair da requisição. O campo é obrigatório.

- Para **Header**: o nome do cabeçalho, por exemplo `X-GitHub-Event`. Maiúsculas e minúsculas não
  importam. Um cabeçalho ausente da requisição resulta em um valor vazio.
- Para **Body** em **JSON**: o caminho para o valor, com as partes unidas por pontos; um item de
  um array é escrito como `.[0]`, por exemplo `repository.full_name` ou `commits.[0].id`. Um
  caminho ausente do corpo não resulta em nenhum valor, e a tarefa mantém o definido na
  integração.
- Para **Body** em **String**: não é usado, mas o formulário exige um valor; digite qualquer coisa.

## Variable Usage {#integrationValueForm-variableType}

Para onde o valor vai em uma tarefa de um modelo:

- **Variables**: uma variável extra da tarefa, do mesmo jeito que a resposta a uma variável de
  pesquisa é passada. Ela substitui uma resposta ou uma variável do grupo de variáveis com o mesmo
  nome.
- **Task Params**: um parâmetro da tarefa, do mesmo jeito que as opções da caixa de diálogo da
  tarefa são armazenadas. O nome `git_branch` define o ramo da tarefa quando o modelo permite
  substituir o ramo; um `refs/heads/` no início é removido.

Uma integração que executa um workflow ignora o uso: cada valor se torna a resposta à variável de
pesquisa do workflow indicada em **Variable**.

## Variable {#integrationValueForm-variable}

O nome com que o valor é passado, por exemplo `branch` para uma variável extra que o playbook lê
como `{{ branch }}`. O campo é obrigatório.
