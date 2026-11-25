---
form: src/components/EnvironmentForm.vue
skip: varGroupSync-syncInterval
---

# Editar Grupo de Variáveis

Um grupo de variáveis guarda as variáveis extras, as variáveis de ambiente e os segredos que o
Semaphore passa para cada tarefa de um modelo de tarefa que usa o grupo. O Ansible recebe as
variáveis extras como `--extra-vars`, o Terraform e o OpenTofu como `-var`; as variáveis de
ambiente são definidas para o processo da tarefa. Um modelo pode usar vários grupos: quando dois
deles definem o mesmo nome, vence o grupo que vem depois no modelo.

## Nome do Grupo {#varGroup-name}

O nome com que o grupo aparece na página Grupos de Variáveis e no seletor de grupo de variáveis de
um modelo de tarefa. O campo é obrigatório.

## Secret storage {#varGroup-secretStorageId}

Exibido quando o grupo é novo e o servidor oferece suporte a armazenamentos de segredos. Opcional.

O armazenamento de segredos em que os segredos deste grupo são guardados, por exemplo um HashiCorp
Vault. Deixe o campo vazio para guardar os segredos no banco de dados do Semaphore, criptografados.

O armazenamento é escolhido uma única vez, quando o grupo é criado: um grupo existente mantém o seu
armazenamento, e o campo deixa de ser exibido. Um armazenamento marcado como **Read only** não
aceita novos segredos: salvar um segredo nele falha.

## Secret key prefix {#varGroup-secretStorageKeyPrefix}

Exibido quando o grupo é novo e o servidor oferece suporte a armazenamentos de segredos. Disponível
depois que um **Secret storage** é escolhido.

O início do caminho que cada segredo deste grupo recebe no armazenamento. O Semaphore acrescenta a
ele 10 caracteres aleatórios: com o prefixo `semaphore/prod/`, um segredo é armazenado como
`semaphore/prod/` seguido desses caracteres. O prefixo é usado como está, então termine-o com `/`
quando os segredos devem ficar sob ele como em uma pasta.

O prefixo é definido uma única vez, quando o grupo é criado. A aba **Secrets** o mostra como o
padrão do caminho de origem.

## Table, JSON ou YAML {#varGroup-extraVarsMode}

Como você edita as variáveis extras:

- **Table**: uma linha por variável, cada uma com um tipo. Adicione uma linha com o botão **+** ao
  lado do seletor.
- **JSON**: todas as variáveis como um único objeto JSON.
- **YAML**: todas as variáveis como um único mapeamento YAML.

A caixa de diálogo abre em **Table**. Trocar de modo converte o que você digitou; quando o conteúdo
não pode ser lido, a caixa de diálogo permanece no modo em que você estava e mostra o erro. Seja
qual for o modo, as variáveis são salvas como um único objeto de nomes e valores.

## Variáveis extras em JSON {#varGroup-extraVarsJson}

Exibido quando as variáveis extras são editadas como **JSON**.

As variáveis extras como um único objeto JSON. Um nome não pode ser vazio; um valor pode ser uma
string, um número, uma lista ou um objeto:

```json
{
  "app_version": "1.4.2",
  "replicas": 3,
  "regions": ["eu-west-1", "us-east-1"]
}
```

## Variáveis extras em YAML {#varGroup-extraVarsYaml}

Exibido quando as variáveis extras são editadas como **YAML**.

As variáveis extras como um único mapeamento YAML. O texto é lido com os tipos que o JSON conhece:
uma data sem aspas como `2024-01-01` continua sendo uma string, e `.inf` ou `.nan` são recusados.
Uma lista ou um escalar no nível superior também é recusado.

```yaml
app_version: 1.4.2
replicas: 3
regions:
  - eu-west-1
  - us-east-1
```

## Nome da variável {#varGroup-varName}

Exibido quando as variáveis extras são editadas como **Table**, em cada linha.

O nome da variável como o playbook ou o código Terraform a vê, por exemplo `app_version`.

## Tipo da variável {#varGroup-varType}

Exibido quando as variáveis extras são editadas como **Table**, em cada linha.

Como o valor da linha é lido:

- **String**: o texto como ele é.
- **Number**: um número, por exemplo `42`.
- **List**: um array JSON, por exemplo `["a", "b"]`.
- **Dict**: um objeto JSON, por exemplo `{"key": "value"}`.

Um valor que não corresponde ao seu tipo interrompe o salvamento, com o nome da variável.

## Valor da variável {#varGroup-varValue}

Exibido quando as variáveis extras são editadas como **Table**, em cada linha.

O valor da variável, no formato do seu tipo. Para um **List** ou um **Dict**, o botão no final do
campo abre um editor maior.

O Semaphore passa os valores para o Ansible como dados: uma expressão `{{ … }}` dentro de um valor
não é avaliada, o playbook recebe o texto como ele é.

## Nome da variável de ambiente {#varGroup-envName}

Na aba **Variables**, em **Variáveis de ambiente**, em cada linha. Adicione uma linha com o botão
**+** da seção.

O nome de uma variável definida no ambiente do processo que executa a tarefa, por exemplo
`AWS_REGION`.

## Valor da variável de ambiente {#varGroup-envValue}

Na aba **Variables**, em **Variáveis de ambiente**, em cada linha.

O valor da variável, como texto. Ele é armazenado como texto simples e exibido a todos que abrem o
grupo: coloque senhas e tokens na aba **Secrets**.

## Sync keys enabled {#varGroup-syncEnabled}

Exibido na aba **Secrets** de um grupo que tem um armazenamento de segredos.

Ative para importar segredos do armazenamento para este grupo, a partir dos caminhos definidos em
**Sync paths**. A importação é executada quando você clica no botão de sincronização do grupo na
página Grupos de Variáveis, e por conta própria no intervalo definido em **Sync paths**.
Desativado, nada é importado.

## Sync paths {#varGroup-syncPaths}

Exibido quando **Sync keys enabled** está ativado.

Abre a caixa de diálogo com os caminhos do armazenamento a partir dos quais importar e o intervalo
da importação automática. O número no botão é o número de caminhos. O botão de sincronização do
grupo na página Grupos de Variáveis permanece desativado até que o grupo tenha pelo menos um
caminho.

A caixa de diálogo tem apenas **Fechar**: o que você altera nela é salvo com o grupo, pelo
**Salvar** desta caixa de diálogo.

## Nome da variável secreta {#varGroup-secretVarName}

Na aba **Secrets**, em **Variáveis extras**, em cada linha. Adicione uma linha com o botão **+** da
seção.

O nome de uma variável extra cujo valor é um segredo. A tarefa a recebe do mesmo jeito que recebe
as variáveis extras da aba **Variables**; um segredo vence uma variável extra com o mesmo nome. O
Ansible recebe os segredos em um arquivo criptografado com o Ansible Vault, o Terraform e o
OpenTofu como `-var`. Como a caixa de diálogo avisa, um segredo passado dessa forma ainda pode
aparecer em texto claro no log do Ansible.

## Valor da variável secreta {#varGroup-secretVarValue}

Na aba **Secrets**, em **Variáveis extras**, em cada linha.

O valor do segredo. O Semaphore o criptografa, ou o grava no armazenamento de segredos do grupo, e
nunca mais o mostra: quando você abre o grupo depois, o campo fica vazio. Deixe-o vazio para
manter o valor armazenado, digite um novo para substituí-lo. Excluir a linha exclui o segredo
quando você salva o grupo.

## Nome da variável de ambiente secreta {#varGroup-secretEnvName}

Na aba **Secrets**, em **Variáveis de ambiente**, em cada linha. Adicione uma linha com o botão
**+** da seção.

O nome de uma variável de ambiente cujo valor é um segredo, por exemplo `DB_PASSWORD`. Ela é
definida no ambiente do processo que executa a tarefa, como as variáveis de ambiente da aba
**Variables**.

## Valor da variável de ambiente secreta {#varGroup-secretEnvValue}

Na aba **Secrets**, em **Variáveis de ambiente**, em cada linha.

O valor do segredo. Assim como para uma variável extra secreta, ele nunca mais é exibido: deixe o
campo vazio para manter o valor armazenado, digite um novo para substituí-lo.
