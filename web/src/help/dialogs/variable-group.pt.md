---
form: src/components/EnvironmentForm.vue
skip: varGroupSync-syncInterval
---

# Editar Grupo de Variáveis

Um grupo de variáveis guarda as variáveis extra, as variáveis de ambiente e os segredos que o
Semaphore transmite a todas as tarefas de um modelo de tarefa que utilize o grupo. O Ansible
recebe as variáveis extra como `--extra-vars`, o Terraform e o OpenTofu como `-var`; as variáveis
de ambiente são definidas para o processo da tarefa. Um modelo pode utilizar vários grupos:
quando dois deles definem o mesmo nome, prevalece o grupo que surge mais tarde no modelo.

## Nome do Grupo {#varGroup-name}

O nome com que o grupo aparece na página Grupos de Variáveis e no seletor de grupos de variáveis
de um modelo de tarefa. O campo é obrigatório.

## Secret storage {#varGroup-secretStorageId}

Apresentado quando o grupo é novo e o servidor suporta armazenamentos de segredos. Opcional.

O armazenamento de segredos em que são guardados os segredos deste grupo, por exemplo um
HashiCorp Vault. Deixe o campo vazio para guardar os segredos na base de dados do Semaphore,
cifrados.

O armazenamento é escolhido uma única vez, quando o grupo é criado: um grupo existente mantém o
seu armazenamento, e o campo deixa de ser apresentado. Um armazenamento marcado como
**Read only** não aceita novos segredos: guardar nele um segredo falha.

## Secret key prefix {#varGroup-secretStorageKeyPrefix}

Apresentado quando o grupo é novo e o servidor suporta armazenamentos de segredos. Disponível
depois de escolhido um **Secret storage**.

O início do caminho que cada segredo deste grupo recebe no armazenamento. O Semaphore
acrescenta-lhe 10 caracteres aleatórios: com o prefixo `semaphore/prod/`, um segredo é guardado
como `semaphore/prod/` seguido desses caracteres. O prefixo é utilizado tal como está, por isso
termine-o com `/` quando os segredos devam ficar dentro dele como numa pasta.

O prefixo é definido uma única vez, quando o grupo é criado. O separador **Secrets** mostra-o como
padrão do caminho de origem.

## Table, JSON ou YAML {#varGroup-extraVarsMode}

Como edita as variáveis extra:

- **Table**: uma linha por variável, cada uma com um tipo. Adicione uma linha com o botão **+**
  junto ao comutador.
- **JSON**: todas as variáveis como um único objeto JSON.
- **YAML**: todas as variáveis como um único mapeamento YAML.

A caixa de diálogo abre em **Table**. Mudar de modo converte o que escreveu; quando não é
possível lê-lo, a caixa de diálogo permanece no modo em que estava e mostra o erro. Qualquer que
seja o modo, as variáveis são guardadas como um único objeto de nomes e valores.

## Variáveis extra em JSON {#varGroup-extraVarsJson}

Apresentado quando as variáveis extra são editadas como **JSON**.

As variáveis extra como um único objeto JSON. Um nome não pode estar vazio; um valor pode ser uma
cadeia de caracteres, um número, uma lista ou um objeto:

```json
{
  "app_version": "1.4.2",
  "replicas": 3,
  "regions": ["eu-west-1", "us-east-1"]
}
```

## Variáveis extra em YAML {#varGroup-extraVarsYaml}

Apresentado quando as variáveis extra são editadas como **YAML**.

As variáveis extra como um único mapeamento YAML. O texto é lido com os tipos que o JSON conhece:
uma data sem aspas como `2024-01-01` continua a ser uma cadeia de caracteres, e `.inf` ou `.nan`
são recusados. Uma lista ou um escalar no nível superior também é recusado.

```yaml
app_version: 1.4.2
replicas: 3
regions:
  - eu-west-1
  - us-east-1
```

## Nome da variável {#varGroup-varName}

Apresentado quando as variáveis extra são editadas como **Table**, em cada linha.

O nome da variável tal como o playbook ou o código Terraform a vê, por exemplo `app_version`.

## Tipo da variável {#varGroup-varType}

Apresentado quando as variáveis extra são editadas como **Table**, em cada linha.

Como é lido o valor da linha:

- **String**: o texto tal como está.
- **Number**: um número, por exemplo `42`.
- **List**: um array JSON, por exemplo `["a", "b"]`.
- **Dict**: um objeto JSON, por exemplo `{"key": "value"}`.

Um valor que não corresponde ao seu tipo interrompe a gravação, indicando o nome da variável.

## Valor da variável {#varGroup-varValue}

Apresentado quando as variáveis extra são editadas como **Table**, em cada linha.

O valor da variável, no formato do seu tipo. Para uma **List** ou um **Dict**, o botão no fim do
campo abre um editor maior.

O Semaphore transmite os valores ao Ansible como dados: uma expressão `{{ … }}` dentro de um valor
não é avaliada, o playbook recebe o texto tal como está.

## Nome da variável de ambiente {#varGroup-envName}

No separador **Variables**, em **Variáveis de Ambiente**, em cada linha. Adicione uma linha com o
botão **+** da secção.

O nome de uma variável definida no ambiente do processo que executa a tarefa, por exemplo
`AWS_REGION`.

## Valor da variável de ambiente {#varGroup-envValue}

No separador **Variables**, em **Variáveis de Ambiente**, em cada linha.

O valor da variável, como texto. É guardado em texto simples e mostrado a todos os que abrem o
grupo: coloque as palavras-passe e os tokens no separador **Secrets**.

## Sync keys enabled {#varGroup-syncEnabled}

Apresentado no separador **Secrets** de um grupo que tem um armazenamento de segredos.

Ative-o para importar segredos do armazenamento para este grupo, a partir dos caminhos definidos
em **Sync paths**. A importação é executada quando prime o botão de sincronização do grupo na
página Grupos de Variáveis, e automaticamente no intervalo definido em **Sync paths**.
Desativado, nada é importado.

## Sync paths {#varGroup-syncPaths}

Apresentado quando **Sync keys enabled** está ativado.

Abre a caixa de diálogo com os caminhos do armazenamento a partir dos quais importar e o
intervalo da importação automática. O número no botão é o número de caminhos. O botão de
sincronização do grupo na página Grupos de Variáveis permanece desativado até que o grupo tenha
pelo menos um caminho.

A caixa de diálogo tem apenas **Fechar**: o que alterar nela é guardado com o grupo, através de
**Salvar** desta caixa de diálogo.

## Nome da variável secreta {#varGroup-secretVarName}

No separador **Secrets**, em **Variáveis Extras**, em cada linha. Adicione uma linha com o botão
**+** da secção.

O nome de uma variável extra cujo valor é um segredo. A tarefa recebe-a tal como recebe as
variáveis extra do separador **Variables**; um segredo prevalece sobre uma variável extra com o
mesmo nome. O Ansible recebe os segredos num ficheiro cifrado com o Ansible Vault, o Terraform e
o OpenTofu como `-var`. Como a caixa de diálogo avisa, um segredo transmitido desta forma pode
ainda assim aparecer em texto simples no registo do Ansible.

## Valor da variável secreta {#varGroup-secretVarValue}

No separador **Secrets**, em **Variáveis Extras**, em cada linha.

O valor do segredo. O Semaphore cifra-o, ou escreve-o no armazenamento de segredos do grupo, e
nunca mais o mostra: quando abrir o grupo mais tarde, o campo está vazio. Deixe-o vazio para
manter o valor guardado, escreva um novo para o substituir. Eliminar a linha elimina o segredo
quando guardar o grupo.

## Nome da variável de ambiente secreta {#varGroup-secretEnvName}

No separador **Secrets**, em **Variáveis de Ambiente**, em cada linha. Adicione uma linha com o
botão **+** da secção.

O nome de uma variável de ambiente cujo valor é um segredo, por exemplo `DB_PASSWORD`. É definida
no ambiente do processo que executa a tarefa, tal como as variáveis de ambiente do separador
**Variables**.

## Valor da variável de ambiente secreta {#varGroup-secretEnvValue}

No separador **Secrets**, em **Variáveis de Ambiente**, em cada linha.

O valor do segredo. Tal como para uma variável extra secreta, nunca volta a ser mostrado: deixe o
campo vazio para manter o valor guardado, escreva um novo para o substituir.
