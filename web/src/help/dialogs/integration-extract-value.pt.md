---
form: src/components/IntegrationExtractValueForm.vue
---

# Edit Extract Value

Um valor extraído retira uma parte do pedido recebido, como o ramo de um evento de push, e
transmite-a à tarefa que a integração inicia. Todos os valores da integração são retirados de
cada pedido que inicia uma tarefa.

## Name {#integrationValueForm-name}

O nome com que o valor aparece na página da integração. Não altera o que é retirado nem para onde
vai. O campo é obrigatório.

## Source of the Value {#integrationValueForm-valueSource}

A parte do pedido de onde o valor é retirado:

- **Header**: o valor de um cabeçalho do pedido.
- **Body**: o corpo do pedido.

## Data Type of Body {#integrationValueForm-bodyDataType}

Apresentado quando **Source of the Value** é **Body**.

Como o corpo é lido:

- **JSON**: o corpo é interpretado como JSON, e **Key** é o caminho para um valor dentro dele. O
  valor mantém o seu tipo JSON: um número, uma lista ou um objeto chegam à tarefa como tal.
- **String**: todo o corpo é tomado como texto, e **Key** não é utilizado.

## Key {#integrationValueForm-key}

O que retirar do pedido. O campo é obrigatório.

- Para **Header**: o nome do cabeçalho, por exemplo `X-GitHub-Event`. As maiúsculas e minúsculas
  não importam. Um cabeçalho que falta no pedido dá um valor vazio.
- Para **Body** em **JSON**: o caminho para o valor, com as suas partes unidas por pontos; um
  elemento de um array escreve-se `.[0]`, por exemplo `repository.full_name` ou
  `commits.[0].id`. Um caminho que falta no corpo não dá nenhum valor, e a tarefa mantém o que
  está definido na integração.
- Para **Body** em **String**: não é utilizado, mas o formulário precisa de um valor; escreva
  qualquer coisa.

## Variable Usage {#integrationValueForm-variableType}

Para onde vai o valor numa tarefa de um modelo:

- **Variables**: uma variável extra da tarefa, da mesma forma que é transmitida a resposta a uma
  variável de pesquisa. Substitui uma resposta ou uma variável do grupo de variáveis com o mesmo
  nome.
- **Task Params**: um parâmetro da tarefa, da mesma forma que são guardadas as opções da caixa de
  diálogo da tarefa. O nome `git_branch` define o ramo da tarefa quando o modelo permite
  substituir o ramo; um `refs/heads/` inicial é removido.

Uma integração que executa um workflow ignora a utilização: cada valor torna-se a resposta à
variável de pesquisa do workflow indicada em **Variable**.

## Variable {#integrationValueForm-variable}

O nome com que o valor é transmitido, por exemplo `branch` para uma variável extra que o playbook
lê como `{{ branch }}`. O campo é obrigatório.
