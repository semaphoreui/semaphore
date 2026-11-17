---
form: src/components/SurveyVars.vue
---

# Variável de Pesquisa

Uma variável de pesquisa é uma pergunta que a caixa de diálogo da tarefa faz sempre que executa o
modelo. A resposta chega à tarefa como uma variável: uma variável extra para o Ansible, um `-var`
para o Terraform e o OpenTofu, um argumento para um script, ou uma variável de ambiente. A
variável é guardada com o modelo, quando guarda a caixa de diálogo do modelo. A mesma caixa de
diálogo edita a pesquisa de um workflow, onde o tipo **Secret** não é proposto.

## Nome {#surveyVarDialog-name}

O nome sob o qual a tarefa recebe a resposta: o nome da variável extra, da variável do Terraform,
do argumento `name=value` de um script, ou da variável de ambiente. Os espaços nas extremidades
são removidos. O campo é obrigatório.

Utilize um nome que nenhuma outra variável da pesquisa tenha e que a ferramenta aceite, por
exemplo `app_version`. Uma resposta substitui uma variável com o mesmo nome do grupo de variáveis
do modelo.

## Título {#surveyVarDialog-title}

O rótulo do campo na caixa de diálogo da tarefa, por exemplo `Version to deploy`. É também o
texto do chip da variável na caixa de diálogo do modelo. O campo é obrigatório.

## Descrição {#surveyVarDialog-description}

Opcional. Uma dica apresentada por baixo do campo da caixa de diálogo da tarefa enquanto este tem
o foco: o que introduzir, ou um exemplo.

## Tipo {#surveyVarDialog-type}

Que tipo de resposta a caixa de diálogo da tarefa pede:

- **String**: uma linha de texto. É a predefinição.
- **Integer**: apenas algarismos; a caixa de diálogo da tarefa recusa tudo o resto, incluindo o
  sinal de menos. A resposta continua a chegar à tarefa como texto: converta-a no playbook, por
  exemplo com `| int`.
- **Secret**: texto que a caixa de diálogo da tarefa oculta. A resposta não é guardada com a
  tarefa, por isso uma nova execução volta a pedi-la, e o Ansible recebe-a num ficheiro
  encriptado com vault. Um segredo não tem valor predefinido.
- **Enum**: um valor escolhido de uma lista que define abaixo.
- **Text**: várias linhas de texto.
- **Select**: vários valores escolhidos de uma lista que define abaixo. A tarefa recebe-os como
  uma lista; num argumento, num `-var` ou numa variável de ambiente a lista é escrita em JSON,
  por exemplo `["a","b"]`.

Alterar o tipo limpa o valor predefinido, e um tipo sem lista limpa os valores.

## Pass variable as {#surveyVarDialog-target}

Como a resposta chega à tarefa:

- **Extra variable**: a forma habitual da ferramenta. O Ansible recebe-a como variável extra, o
  Terraform e o OpenTofu como `-var name=value`, um script Bash, PowerShell ou Python como um
  argumento `name=value` a seguir ao script.
- **Environment variable**: uma variável de ambiente do processo, `name=value`, para qualquer
  tipo de modelo. A variável deixa então de ser passada da outra forma.

## Chave {#surveyVarDialog-valueName}

Apresentado quando o tipo é **Enum** ou **Select**.

O texto que a caixa de diálogo da tarefa mostra para este valor na lista. Cada valor precisa de
uma chave, e as chaves têm de ser diferentes: caso contrário, a caixa de diálogo recusa guardar
a variável.

## Valor {#surveyVarDialog-valueValue}

Apresentado quando o tipo é **Enum** ou **Select**.

O que a tarefa recebe quando este valor é escolhido. Pode ser diferente da sua chave: a chave
`Production` pode passar `prod`.

## Add Value {#surveyVarDialog-addValue}

Apresentado quando o tipo é **Enum** ou **Select**.

Acrescenta uma linha vazia à lista de valores. Um **Enum** ou um **Select** precisa de pelo menos
um valor. O ícone do caixote do lixo de uma linha remove-a.

## Default value {#surveyVarDialog-defaultValue}

Apresentado para todos os tipos exceto **Secret**.

A resposta que a caixa de diálogo da tarefa preenche quando abre; pode alterá-la aí. Para um
**Enum** escolha um dos valores, para um **Select** qualquer número deles. Deixe o campo vazio
para começar com uma resposta vazia.

O valor predefinido é preenchido apenas pela caixa de diálogo da tarefa. Uma tarefa iniciada de
outra forma, por exemplo através da API, não recebe valor predefinido.

## Obrigatório {#surveyVarDialog-required}

A caixa de diálogo da tarefa não inicia uma tarefa enquanto a variável não tiver resposta. Tal
como o valor predefinido, a verificação é feita apenas pela caixa de diálogo da tarefa, não pelo
servidor.
