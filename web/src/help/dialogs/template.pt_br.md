---
form: src/components/TemplateForm.vue, src/components/SurveyVars.vue, src/components/TemplateVaults.vue
skip: surveyVarDialog-name, surveyVarDialog-title, surveyVarDialog-description, surveyVarDialog-type, surveyVarDialog-target, surveyVarDialog-valueName, surveyVarDialog-valueValue, surveyVarDialog-defaultValue, surveyVarDialog-required, templateVaultDialog-name, templateVaultDialog-type, templateVaultDialog-vaultKeyId, templateVaultDialog-script
---

# Editar Modelo 'Ansible Playbook'

Um modelo de tarefa diz o que o Semaphore executa e como: o repositório e o arquivo nele, o
inventário e as variáveis, e quais opções um usuário pode alterar quando uma tarefa é iniciada.
Toda tarefa é iniciada a partir de um modelo. Os campos da caixa de diálogo dependem do app do
modelo (Ansible, Terraform, OpenTofu, Terragrunt, Bash, Python, PowerShell e outros): um campo que
só alguns apps têm diz isso. As alterações entram em vigor para as tarefas iniciadas depois de
**Salvar**.

## Tipo de modelo {#templateForm-typeTabs}

As abas no topo definem que tipo de modelo ele é:

- **Tarefa**: executa o código. A maioria dos modelos é deste tipo.
- **Construir**: produz um artefato. Cada tarefa de construção recebe uma versão, contada a partir
  de **Versão Inicial**.
- **Implantar**: implanta um artefato produzido por uma construção. Ao iniciá-la, você escolhe a
  tarefa de construção cuja versão ela implanta.

Uma tarefa de construção ou de implantação recebe no seu ambiente as variáveis
`SEMAPHORE_TASK_TYPE`, `SEMAPHORE_TASK_TARGET_VERSION` (construção) e
`SEMAPHORE_TASK_INCOMING_VERSION` (implantação). O Ansible recebe os mesmos valores na variável
extra `semaphore_vars.task_details`.

## Versão Inicial {#templateForm-startVersion}

Exibido quando o tipo é **Construir**. Obrigatório.

A versão da primeira tarefa de construção, por exemplo `1.0.0` ou `v0.1`. Cada construção seguinte
aumenta em um o último número da versão anterior e mantém o texto ao redor dele e os zeros à
esquerda: `1.0.0`, `1.0.1`, `1.0.2`. Aumente o número aqui acima da última construção para
avançar, por exemplo para `2.0.0`; uma versão sem nenhum número nunca muda.

## Modelo de Construção {#templateForm-buildTemplateId}

Exibido quando o tipo é **Implantar**. Obrigatório.

O modelo cujas tarefas produzem os artefatos que este modelo implanta. A lista contém os modelos de
construção do projeto e, abaixo deles, os modelos de implantação: uma implantação pode seguir outra
implantação, e então recebe a versão que essa implantação recebeu da sua construção.

## Execução Automática {#templateForm-autorun}

Exibido quando o tipo é **Implantar**.

Inicia por conta própria uma tarefa deste modelo toda vez que uma tarefa do **Modelo de Construção**
é bem-sucedida, com a versão que essa tarefa produziu. Uma construção que falha não inicia nada.

## Nome {#templateForm-name}

O nome com que o modelo aparece na página Modelos de Tarefa. O campo é obrigatório.

## Repositório {#templateForm-repositoryId}

O repositório do projeto que contém o código a executar. O campo é obrigatório. Antes de cada
tarefa, o Semaphore atualiza a sua cópia do repositório e executa o código a partir dela.

O rótulo à direita do campo mostra o ramo de que as tarefas fazem o checkout. Clique nele para
abrir o campo **Ramo**.

## Ramo {#templateForm-gitBranch}

Exibido depois de um clique no rótulo do ramo à direita de **Repositório**.

O ramo ou a tag de que as tarefas deste modelo fazem o checkout, em vez do ramo definido no
repositório. Deixe o campo vazio para usar o ramo do repositório. Quando o Semaphore consegue ler
os ramos do repositório, o campo os lista.

O campo fica desativado quando o repositório é um monorepo: todos os modelos compartilham um único
checkout dele, no ramo do repositório.

## Caminho para o arquivo playbook {#templateForm-playbook}

O que executar, como um caminho a partir da raiz do repositório, por exemplo `deploy/site.yml`. O
caminho não pode ser absoluto nem levar para fora do repositório. A lista oferece os arquivos do
repositório; você também pode digitar um caminho que não está na lista. O botão à direita do campo
lê a lista novamente.

O rótulo e o significado mudam com o app:

- **Ansible**: o playbook. É obrigatório; a lista oferece arquivos `.yml` e `.yaml`.
- **Terraform**, **OpenTofu**, **Terragrunt**: **Subdirectory path**, o diretório com o código.
  Deixe vazio para executar o código na raiz do repositório. Digite `/` para listar o que um
  diretório contém.
- **Bash**, **Python**, **PowerShell** e outros apps: **Script Filename**, o script a executar. É
  obrigatório. O script é executado na raiz do repositório.

## Different working directory {#templateForm-workingDirectoryToggle}

Exibido para modelos de Ansible.

Executa o `ansible-playbook` em um diretório diferente da raiz do repositório. Marque para exibir
o campo **Working directory**. Desmarque para voltar a executar na raiz.

## Working directory {#templateForm-workingDirectory}

Exibido quando **Different working directory** está marcado. Obrigatório.

O diretório em que o `ansible-playbook` e o `ansible-galaxy` são executados, como um caminho a
partir da raiz do repositório, por exemplo `ansible`. Use quando o `ansible.cfg` do projeto fica em
um subdiretório, já que o Ansible lê o arquivo do diretório atual. O caminho do playbook continua
relativo à raiz do repositório. O caminho não pode ser absoluto nem levar para fora do
repositório.

## Inventário {#templateForm-inventoryId}

Exibido para modelos de Ansible, Terraform, OpenTofu e Terragrunt.

- **Ansible**: o inventário contra o qual o playbook é executado, com as credenciais para os
  hosts. O campo é obrigatório.
- **Terraform**, **OpenTofu**, **Terragrunt**: **Workspace**, o workspace em que o código é
  executado. O Semaphore o seleciona antes do plano e o cria quando ele não existe. Um novo modelo
  salvo sem workspace recebe um chamado `default`.

A lista contém os inventários do projeto feitos para o app do modelo.

## Grupos de Variáveis {#templateForm-environmentIds}

Opcional. Os grupos de variáveis cujas variáveis, variáveis de ambiente e segredos cada tarefa
deste modelo recebe. Você pode escolher vários: eles são mesclados na ordem do campo, e um grupo
posterior vence um anterior quando ambos definem o mesmo nome.

Como as variáveis chegam ao código depende do app: variáveis extras para o Ansible, `-var` para o
Terraform e o OpenTofu, argumentos `name=value` para um script. As variáveis de ambiente são
definidas para todos os apps.

## Visualizar {#templateForm-viewId}

Opcional. A visualização em que o modelo é listado: as visualizações são as abas da página Modelos
de Tarefa, que agrupam os modelos de um projeto. Deixe o campo vazio para não colocar o modelo em
nenhuma visualização.

## Tag do Executor {#templateForm-runnerTag}

Exibido quando há executores de projeto disponíveis.

As tarefas deste modelo são executadas apenas em um executor que tenha esta tag, mesmo quando o
servidor poderia executá-las por conta própria. A tag substitui a tag de executor do inventário.
Deixe o campo vazio para usar a tag do inventário, ou nenhuma.

## Executor image {#templateForm-executorImage}

A imagem de contêiner que o executor inicia para as tarefas deste modelo, por exemplo
`semaphoreui/job:latest`. Deixe o campo vazio para usar a imagem da configuração do executor.

Apenas os executores Docker e Kubernetes executam tarefas em um contêiner, então o campo só fica
habilitado quando um deles está disponível. Uma tarefa executada sem contêiner o ignora.

## Variáveis de Pesquisa {#surveyVars}

Os valores que são pedidos a um usuário quando uma tarefa deste modelo é iniciada. Clique em **+**
para adicionar uma variável, em um chip para alterar uma, no X de um chip para removê-la; arraste
os chips para definir a ordem em que as perguntas são feitas.

As respostas chegam ao código como as variáveis de um grupo de variáveis, ou como variáveis de
ambiente, conforme cada variável define. Elas vencem um grupo de variáveis que define o mesmo nome.

## Allow parallel tasks {#templateForm-allowParallelTasks}

Permite que as tarefas deste modelo sejam executadas ao mesmo tempo. Sem isso, uma tarefa do
modelo aguarda na fila até que a que está em execução termine. Os limites de tarefas paralelas do
projeto e do servidor se aplicam em ambos os casos.

## Issue JWT to task runner {#templateForm-jwtEnabled}

Exibido quando o administrador ativou a emissão de JWT no servidor.

Dá a cada tarefa deste modelo um JSON Web Token assinado pelo Semaphore, na variável de ambiente
`SEMAPHORE_JWT`. O código pode apresentar o token a outro serviço para provar qual tarefa ele é: o
token indica a tarefa, o projeto, o modelo e o usuário que a iniciou. O serviço verifica a
assinatura com as chaves públicas em `/.well-known/jwks.json` do servidor Semaphore.

## JWT audience {#templateForm-jwtAudience}

Exibido quando **Issue JWT to task runner** está marcado. Opcional.

Os serviços a que o token se destina, gravados na claim `aud` dele, por exemplo
`https://vault.example.com`. Pressione Enter depois de cada valor. Até 32 valores; deixe o campo
vazio para um token sem audience.

## JWT TTL {#templateForm-jwtTtl}

Exibido quando **Issue JWT to task runner** está marcado. Opcional.

Por quanto tempo o token permanece válido a partir do início da tarefa, como uma duração do tipo
`30s`, `15m` ou `1h`. Deixe o campo vazio para usar o padrão do servidor, uma hora, a menos que o
administrador tenha definido outro. O valor não pode exceder o máximo do servidor, 24 horas, a
menos que definido de outra forma.

## Quero executar uma tarefa pelo cron apenas para novos commits de algum repositório {#templateForm-cronVisible}

Inicia uma tarefa deste modelo quando um novo commit aparece em um repositório. O Semaphore
pergunta ao repositório qual é o último commit no intervalo escolhido e inicia uma tarefa apenas
quando o commit mudou desde a última verificação.

Marcar a caixa e salvar cria a agenda da verificação; desmarcá-la e salvar exclui essa agenda. As
outras agendas do modelo são mantidas.

## Repositório {#templateForm-cronRepositoryId}

Exibido quando **Quero executar uma tarefa pelo cron apenas para novos commits de algum
repositório** está marcado. Obrigatório.

O repositório a observar em busca de novos commits. O Semaphore observa o ramo definido nesse
repositório, não o **Ramo** do modelo.

## intervaloDeVerificação {#templateForm-cronFormat}

Exibido quando **Quero executar uma tarefa pelo cron apenas para novos commits de algum
repositório** está marcado.

Com que frequência o Semaphore verifica se há um novo commit no repositório: de a cada minuto a uma
vez por dia. O padrão é a cada minuto.

## Suprimir alertas de sucesso {#templateForm-suppressSuccessAlerts}

Impede que as tarefas deste modelo que são bem-sucedidas enviem alertas. Os alertas das tarefas
com falha continuam sendo enviados. Só tem efeito quando os alertas estão ativados para o projeto.

## Suprimir alertas de erro {#templateForm-suppressErrorAlerts}

Impede que as tarefas deste modelo que falham enviem alertas. Os alertas das tarefas bem-sucedidas
continuam sendo enviados. Só tem efeito quando os alertas estão ativados para o projeto.

## CLI args {#templateForm-args}

Argumentos adicionados ao comando de cada tarefa. Cada chip é um argumento, passado como está, sem
shell: coloque `--forks=10` em um chip, ou `--forks` e `10` em dois. Arraste os chips para mudar a
ordem deles. O botão à direita da lista abre os argumentos como um array JSON:

```json
["--forks=10", "-e", "env=staging"]
```

- **Ansible**: adicionados ao `ansible-playbook`, antes do playbook.
- **Terraform**, **OpenTofu**, **Terragrunt**: adicionados ao `plan` e ao `apply`, não ao `init`.
- Um script: adicionados depois do nome do script.

## Argumentos CLI {#templateForm-allowOverrideArgsInTask}

Em **Solicitações**. Permite que o usuário adicione argumentos próprios quando uma tarefa é
iniciada. Eles são adicionados depois dos argumentos do modelo.

## Ramo {#templateForm-allowOverrideBranchInTask}

Em **Solicitações**. Permite que o usuário escolha outro ramo, tag ou commit quando uma tarefa é
iniciada, em vez do ramo do modelo. Sem isso, a tarefa sempre faz o checkout do ramo do modelo.

A caixa fica desativada quando o repositório é um monorepo.

## Inventário {#templateForm-allowOverrideInventory}

Em **Solicitações**. Exibido para modelos de Ansible, Terraform, OpenTofu e Terragrunt.

Permite que o usuário escolha outro inventário quando uma tarefa é iniciada. Apenas os modelos de
Ansible usam o inventário escolhido ali: uma tarefa de Terraform, OpenTofu ou Terragrunt é
executada no workspace do modelo mesmo quando a caixa está marcada.

## Limite {#templateForm-limit}

Exibido para modelos de Ansible.

Os hosts ou grupos do inventário em que o playbook é executado, passados como `--limit`. Cada chip
é um padrão, por exemplo `web` ou `db01`; deixe a lista vazia para executar em todos os hosts.

## Tags {#templateForm-tags}

Exibido para modelos de Ansible.

Executa apenas as tarefas e roles com estas tags, passadas como `--tags`. Deixe a lista vazia para
executar tudo.

## Pular tags {#templateForm-skipTags}

Exibido para modelos de Ansible.

Pula as tarefas e roles com estas tags, passadas como `--skip-tags`.

## Cofres {#templateVaults}

Exibido para modelos de Ansible.

As senhas do Ansible Vault de que o playbook precisa para descriptografar os seus arquivos. Clique
em **+** para adicionar uma, em um chip para alterá-la, no X de um chip para removê-la. Cada cofre
tem um nome, o seu vault ID, e fornece a senha a partir de uma chave do Armazenamento de Chaves ou
de um script cliente do repositório.

## Aprovação automática {#templateForm-autoApprove}

Exibido para modelos de Terraform, OpenTofu e Terragrunt.

Aplica o plano sem perguntar. Sem isso, uma tarefa para depois do plano e aguarda até que um
usuário a confirme ou rejeite; um plano sem alterações encerra a tarefa sem nada a aplicar.

## Override backend settings {#templateForm-overrideBackend}

Exibido para modelos de Terraform, OpenTofu e Terragrunt quando o Semaphore pode guardar o estado
do Terraform.

Guarda o estado do código no Semaphore. Antes do `init`, o Semaphore grava no diretório do código
um arquivo que define o backend `http`, e o exclui depois da tarefa:

```hcl
terraform {
  backend "http" {
  }
}
```

## Backend filename {#templateForm-backendFilename}

Exibido com **Override backend settings**; habilitado quando essa caixa está marcada. Opcional.

O nome do arquivo em que o Semaphore grava o backend, por exemplo `semaphore_backend.tf`. O padrão
é `backend.tf`. O nome precisa terminar com `.tf` e conter apenas letras, dígitos, `_`, `-` e `.`.
Escolha um nome que nenhum arquivo do seu código tenha: o Semaphore sobrescreve esse arquivo e o
exclui depois da tarefa.

## Limite {#templateForm-allowOverrideLimit}

Em **Solicitações de Ansible**. Permite que o usuário defina os hosts em que executar quando uma
tarefa é iniciada, em vez do **Limite** do modelo.

## Tags {#templateForm-allowOverrideTags}

Em **Solicitações de Ansible**. Permite que o usuário defina as tags quando uma tarefa é iniciada,
em vez das **Tags** do modelo.

## Pular tags {#templateForm-allowOverrideSkipTags}

Em **Solicitações de Ansible**. Permite que o usuário defina as tags a pular quando uma tarefa é
iniciada, em vez das **Pular tags** do modelo.

## Depurar {#templateForm-allowDebug}

Em **Solicitações de Ansible**. Permite que o usuário ative a saída detalhada quando uma tarefa é
iniciada, de `-v` a `-vvvvvv`.

## Execução Simulada {#templateForm-showDryRun}

Em **Solicitações de Ansible**. Marcado por padrão. Oferece a opção **Execução Simulada** quando uma
tarefa é iniciada, que executa o playbook com `--check`: o Ansible informa o que alteraria e não
altera nada. Desmarque a caixa para retirar a opção.

## Diferença {#templateForm-showDiff}

Em **Solicitações de Ansible**. Marcado por padrão. Oferece a opção **Diferença** quando uma tarefa
é iniciada, que executa o playbook com `--diff` para mostrar as alterações feitas nos arquivos.
Desmarque a caixa para retirar a opção.

## Aprovação automática {#templateForm-allowAutoApprove}

Nas solicitações de Terraform, OpenTofu e Terragrunt. Permite que o usuário marque **Auto Approve**
quando uma tarefa é iniciada, para aplicar o plano sem confirmação. A caixa de diálogo da tarefa
oferece essa opção de qualquer forma, mas sem esta caixa o Semaphore a ignora: a tarefa aguarda uma
confirmação depois do plano, a menos que a **Aprovação automática** do modelo esteja marcada.

## Galaxy install options {#templateForm-galaxy}

Exibido para modelos de Ansible. Clique nele para exibir as opções da instalação do Galaxy; o
rótulo conta as opções que diferem dos padrões.

Antes de o playbook ser executado, o Semaphore instala com o `ansible-galaxy` as roles e coleções
que o repositório pede. Ele procura `requirements.yml`, `roles/requirements.yml` e
`collections/requirements.yml` ao lado do playbook e na raiz do repositório, e executa a
instalação apenas quando um arquivo mudou desde a última tarefa.

## Skip Galaxy install {#templateForm-skipGalaxyInstall}

Em **Galaxy install options**. Pula a instalação do Galaxy: o playbook é executado com as roles e
coleções instaladas antes. Use quando os requisitos são instalados de outra forma ou a instalação é
lenta e desnecessária.

## Role install args {#templateForm-galaxyRoleArgs}

Em **Galaxy install options**. Argumentos adicionados ao `ansible-galaxy role install`, um por chip,
por exemplo `--ignore-certs` ou `--server=https://galaxy.example.com`. O Semaphore aceita apenas as
flags da instalação que são seguras de passar: uma flag desconhecida, `--token`, `--api-key`, `-p`
ou `--roles-path` faz o modelo falhar ao ser salvo.

## Collection install args {#templateForm-galaxyCollectionArgs}

Em **Galaxy install options**. Argumentos adicionados ao `ansible-galaxy collection install`, um
por chip, por exemplo `--pre` ou `--upgrade`. A mesma regra de **Role install args** se aplica:
`--token`, `--api-key`, `-p` e `--collections-path` não são aceitos.

## Skip Galaxy install {#templateForm-allowOverrideSkipGalaxyInstall}

Em **Galaxy install options**, em **Solicitações**. Permite que o usuário escolha, quando uma
tarefa é iniciada, se a instalação do Galaxy deve ser pulada. A escolha da tarefa vence o **Skip
Galaxy install** do modelo.
