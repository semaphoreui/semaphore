---
form: src/components/SurveyVars.vue
---

# Variável de Pesquisa

Uma variável de pesquisa é uma pergunta que a caixa de diálogo da tarefa faz toda vez que você
executa o modelo. A resposta chega à tarefa como uma variável: uma variável extra para o Ansible,
um `-var` para o Terraform e o OpenTofu, um argumento para um script ou uma variável de ambiente. A
variável é salva com o modelo, quando você salva a caixa de diálogo do modelo. A mesma caixa de
diálogo edita a pesquisa de um workflow, em que o tipo **Secret** não é oferecido.

## Nome {#surveyVarDialog-name}

O nome com que a tarefa recebe a resposta: o nome da variável extra, da variável do Terraform, do
argumento `name=value` de um script ou da variável de ambiente. Os espaços nas extremidades são
removidos. O campo é obrigatório.

Use um nome que nenhuma outra variável da pesquisa tenha e que a ferramenta aceite, por exemplo
`app_version`. Uma resposta substitui uma variável com o mesmo nome do grupo de variáveis do
modelo.

## Título {#surveyVarDialog-title}

O rótulo do campo na caixa de diálogo da tarefa, por exemplo `Version to deploy`. É também o texto
do chip da variável na caixa de diálogo do modelo. O campo é obrigatório.

## Descrição {#surveyVarDialog-description}

Opcional. Uma dica exibida abaixo do campo da caixa de diálogo da tarefa enquanto ele está em
foco: o que informar, ou um exemplo.

## Tipo {#surveyVarDialog-type}

Que tipo de resposta a caixa de diálogo da tarefa pede:

- **String**: uma linha de texto. Este é o padrão.
- **Integer**: apenas dígitos; a caixa de diálogo da tarefa recusa qualquer outra coisa, inclusive
  um sinal de menos. A resposta ainda assim chega à tarefa como texto: converta-a no playbook, por
  exemplo com `| int`.
- **Secret**: texto que a caixa de diálogo da tarefa mascara. A resposta não é salva com a tarefa,
  então uma nova execução a pede novamente, e o Ansible a recebe em um arquivo criptografado com o
  vault. Um segredo não tem valor padrão.
- **Enum**: um valor escolhido de uma lista que você define abaixo.
- **Text**: várias linhas de texto.
- **Select**: vários valores escolhidos de uma lista que você define abaixo. A tarefa os recebe
  como uma lista; em um argumento, em um `-var` ou em uma variável de ambiente, a lista é escrita
  como JSON, por exemplo `["a","b"]`.

Alterar o tipo limpa o valor padrão, e um tipo sem lista limpa os valores.

## Pass variable as {#surveyVarDialog-target}

Como a resposta chega à tarefa:

- **Extra variable**: a forma usual da ferramenta. O Ansible a recebe como variável extra, o
  Terraform e o OpenTofu como `-var name=value`, um script Bash, PowerShell ou Python como um
  argumento `name=value` depois do script.
- **Environment variable**: uma variável de ambiente do processo, `name=value`, para todo tipo de
  modelo. A variável então não é passada da outra forma.

## Chave {#surveyVarDialog-valueName}

Exibido quando o tipo é **Enum** ou **Select**.

O texto que a caixa de diálogo da tarefa mostra para este valor na lista. Cada valor precisa de uma
chave, e as chaves precisam ser diferentes: caso contrário, a caixa de diálogo se recusa a salvar a
variável.

## Valor {#surveyVarDialog-valueValue}

Exibido quando o tipo é **Enum** ou **Select**.

O que a tarefa recebe quando este valor é escolhido. Ele pode ser diferente da sua chave: a chave
`Production` pode passar `prod`.

## Add Value {#surveyVarDialog-addValue}

Exibido quando o tipo é **Enum** ou **Select**.

Adiciona uma linha vazia à lista de valores. Um **Enum** ou um **Select** precisa de pelo menos um
valor. O ícone de lixeira de uma linha a remove.

## Default value {#surveyVarDialog-defaultValue}

Exibido para todos os tipos, exceto **Secret**.

A resposta que a caixa de diálogo da tarefa preenche ao abrir; você pode alterá-la lá. Para um
**Enum**, escolha um dos valores; para um **Select**, qualquer quantidade deles. Deixe o campo
vazio para começar com uma resposta vazia.

O padrão é preenchido apenas pela caixa de diálogo da tarefa. Uma tarefa iniciada de outra forma,
pela API por exemplo, não recebe padrão.

## Obrigatório {#surveyVarDialog-required}

A caixa de diálogo da tarefa não inicia uma tarefa até que a variável tenha uma resposta. Assim
como o padrão, a verificação é feita apenas pela caixa de diálogo da tarefa, não pelo servidor.
