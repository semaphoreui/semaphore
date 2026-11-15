---
form: src/components/TemplateForm.vue, src/components/SurveyVars.vue, src/components/TemplateVaults.vue
skip: surveyVarDialog-name, surveyVarDialog-title, surveyVarDialog-description, surveyVarDialog-type, surveyVarDialog-target, surveyVarDialog-valueName, surveyVarDialog-valueValue, surveyVarDialog-defaultValue, surveyVarDialog-required, templateVaultDialog-name, templateVaultDialog-type, templateVaultDialog-vaultKeyId, templateVaultDialog-script
---

# Editar Modelo 'Ansible Playbook'

Um modelo de tarefa indica o que o Semaphore executa e como: o repositório e o ficheiro nele, o
inventário e as variáveis, e as opções que um utilizador pode alterar quando uma tarefa é
iniciada. Todas as tarefas são iniciadas a partir de um modelo. Os campos da caixa de diálogo
dependem da aplicação do modelo (Ansible, Terraform, OpenTofu, Terragrunt, Bash, Python,
PowerShell e outras): um campo que só algumas aplicações têm indica-o. As alterações produzem
efeito nas tarefas iniciadas depois de **Salvar**.

## Tipo de modelo {#templateForm-typeTabs}

Os separadores no topo definem o tipo de modelo:

- **Tarefa**: executa o código. A maioria dos modelos é deste tipo.
- **Construir**: produz um artefacto. Cada tarefa de construção recebe uma versão, contada a
  partir de **Versão Inicial**.
- **Implantar**: implanta um artefacto produzido por uma construção. Ao iniciá-la, escolhe a
  tarefa de construção cuja versão é implantada.

Uma tarefa de construção ou de implantação recebe no seu ambiente as variáveis
`SEMAPHORE_TASK_TYPE`, `SEMAPHORE_TASK_TARGET_VERSION` (construção) e
`SEMAPHORE_TASK_INCOMING_VERSION` (implantação). O Ansible recebe os mesmos valores na variável
extra `semaphore_vars.task_details`.

## Versão Inicial {#templateForm-startVersion}

Apresentado quando o tipo é **Construir**. Obrigatório.

A versão da primeira tarefa de construção, por exemplo `1.0.0` ou `v0.1`. Cada construção
seguinte aumenta em um o último número da versão anterior e mantém o texto à sua volta e os zeros
à esquerda: `1.0.0`, `1.0.1`, `1.0.2`. Aumente aqui o número acima da última construção para
avançar, por exemplo para `2.0.0`; uma versão sem nenhum número nunca muda.

## Modelo de Construção {#templateForm-buildTemplateId}

Apresentado quando o tipo é **Implantar**. Obrigatório.

O modelo cujas tarefas produzem os artefactos que este modelo implanta. A lista contém os modelos
de construção do projeto e, abaixo deles, os modelos de implantação: uma implantação pode seguir
outra implantação e recebe, nesse caso, a versão que essa implantação recebeu da sua construção.

## Execução Automática {#templateForm-autorun}

Apresentado quando o tipo é **Implantar**.

Inicia por si só uma tarefa deste modelo sempre que uma tarefa do **Modelo de Construção** é
concluída com êxito, com a versão que essa tarefa produziu. Uma construção falhada não inicia
nada.

## Nome {#templateForm-name}

O nome com que o modelo aparece na página Modelos de Tarefa. O campo é obrigatório.

## Repositório {#templateForm-repositoryId}

O repositório do projeto que contém o código a executar. O campo é obrigatório. Antes de cada
tarefa, o Semaphore atualiza a sua cópia do repositório e executa o código a partir dela.

A etiqueta à direita do campo mostra o ramo que as tarefas obtêm. Clique nela para abrir o campo
**Ramo**.

## Ramo {#templateForm-gitBranch}

Apresentado depois de um clique na etiqueta do ramo à direita de **Repositório**.

O ramo ou a tag que as tarefas deste modelo obtêm, em vez do ramo definido no repositório. Deixe
o campo vazio para utilizar o ramo do repositório. Quando o Semaphore consegue ler os ramos do
repositório, o campo apresenta-os.

O campo fica desativado quando o repositório é um monorepo: todos os modelos partilham uma única
cópia dele, no ramo do repositório.

## Caminho para o arquivo do playbook {#templateForm-playbook}

O que executar, como caminho a partir da raiz do repositório, por exemplo `deploy/site.yml`. O
caminho não pode ser absoluto nem sair do repositório. A lista propõe os ficheiros do
repositório; também pode escrever um caminho que não esteja na lista. O botão à direita do campo
volta a ler a lista.

A etiqueta e o significado mudam com a aplicação:

- **Ansible**: o playbook. É obrigatório; a lista propõe ficheiros `.yml` e `.yaml`.
- **Terraform**, **OpenTofu**, **Terragrunt**: **Subdirectory path**, o diretório com o código.
  Deixe-o vazio para executar o código na raiz do repositório. Escreva `/` para listar o conteúdo
  de um diretório.
- **Bash**, **Python**, **PowerShell** e outras aplicações: **Script Filename**, o script a
  executar. É obrigatório. O script é executado na raiz do repositório.

## Different working directory {#templateForm-workingDirectoryToggle}

Apresentado para modelos Ansible.

Executa o `ansible-playbook` num diretório que não a raiz do repositório. Assinale-o para mostrar
o campo **Working directory**. Desmarque-o para voltar a executar na raiz.

## Working directory {#templateForm-workingDirectory}

Apresentado quando **Different working directory** está assinalado. Obrigatório.

O diretório em que o `ansible-playbook` e o `ansible-galaxy` são executados, como caminho a partir
da raiz do repositório, por exemplo `ansible`. Utilize-o quando o `ansible.cfg` do projeto se
encontra num subdiretório, uma vez que o Ansible lê o ficheiro a partir do diretório atual. O
caminho do playbook continua relativo à raiz do repositório. O caminho não pode ser absoluto nem
sair do repositório.

## Inventário {#templateForm-inventoryId}

Apresentado para modelos Ansible, Terraform, OpenTofu e Terragrunt.

- **Ansible**: o inventário sobre o qual o playbook é executado, com as credenciais para os hosts.
  O campo é obrigatório.
- **Terraform**, **OpenTofu**, **Terragrunt**: **Workspace**, o workspace em que o código é
  executado. O Semaphore seleciona-o antes do plano e cria-o quando não existe. Um novo modelo
  guardado sem workspace recebe um chamado `default`.

A lista contém os inventários do projeto criados para a aplicação do modelo.

## Grupos de Variáveis {#templateForm-environmentIds}

Opcional. Os grupos de variáveis cujas variáveis, variáveis de ambiente e segredos todas as
tarefas deste modelo recebem. Pode escolher vários: são combinados pela ordem do campo, e um
grupo posterior prevalece sobre um anterior quando ambos definem o mesmo nome.

A forma como as variáveis chegam ao código depende da aplicação: variáveis extra para o Ansible,
`-var` para o Terraform e o OpenTofu, argumentos `name=value` para um script. As variáveis de
ambiente são definidas para todas as aplicações.

## Visualizar {#templateForm-viewId}

Opcional. A visualização em que o modelo aparece: as visualizações são os separadores da página
Modelos de Tarefa, que agrupam os modelos de um projeto. Deixe o campo vazio para não colocar o
modelo em nenhuma visualização.

## Tag do Executor {#templateForm-runnerTag}

Apresentado quando existem executores de projeto disponíveis.

As tarefas deste modelo são executadas apenas num executor que tenha esta tag, mesmo quando o
próprio servidor as poderia executar. A tag tem prioridade sobre a tag de executor do inventário.
Deixe o campo vazio para utilizar a tag do inventário, ou nenhuma.

## Executor image {#templateForm-executorImage}

A imagem de contentor que o executor inicia para as tarefas deste modelo, por exemplo
`semaphoreui/job:latest`. Deixe o campo vazio para utilizar a imagem da configuração do executor.

Apenas os executores Docker e Kubernetes executam tarefas num contentor, pelo que o campo só fica
ativo quando um deles está disponível. Uma tarefa executada sem contentor ignora-o.

## Variáveis de Pesquisa {#surveyVars}

Os valores pedidos a um utilizador quando uma tarefa deste modelo é iniciada. Clique em **+** para
adicionar uma variável, num chip para alterar uma, na cruz de um chip para a remover; arraste os
chips para definir a ordem pela qual as perguntas são feitas.

As respostas chegam ao código como as variáveis de um grupo de variáveis, ou como variáveis de
ambiente, conforme definido em cada variável. Prevalecem sobre um grupo de variáveis que defina o
mesmo nome.

## Allow parallel tasks {#templateForm-allowParallelTasks}

Permite que as tarefas deste modelo sejam executadas em simultâneo. Sem esta opção, uma tarefa do
modelo fica em espera na fila até que a tarefa em execução termine. Os limites de tarefas
paralelas do projeto e do servidor aplicam-se em ambos os casos.

## Issue JWT to task runner {#templateForm-jwtEnabled}

Apresentado quando o administrador ativou a emissão de JWT no servidor.

Dá a cada tarefa deste modelo um JSON Web Token assinado pelo Semaphore, na variável de ambiente
`SEMAPHORE_JWT`. O código pode apresentar o token a outro serviço para provar de que tarefa se
trata: o token identifica a tarefa, o projeto, o modelo e o utilizador que a iniciou. O serviço
verifica a assinatura com as chaves públicas em `/.well-known/jwks.json` do servidor Semaphore.

## JWT audience {#templateForm-jwtAudience}

Apresentado quando **Issue JWT to task runner** está assinalado. Opcional.

Os serviços a que o token se destina, escritos na sua claim `aud`, por exemplo
`https://vault.example.com`. Prima Enter depois de cada valor. Até 32 valores; deixe o campo vazio
para um token sem audiência.

## JWT TTL {#templateForm-jwtTtl}

Apresentado quando **Issue JWT to task runner** está assinalado. Opcional.

Durante quanto tempo o token permanece válido a partir do início da tarefa, como uma duração do
tipo `30s`, `15m` ou `1h`. Deixe o campo vazio para utilizar a predefinição do servidor, uma hora,
salvo se o administrador tiver definido outra. O valor não pode exceder o máximo do servidor, 24
horas, salvo definição em contrário.

## Quero executar uma tarefa pelo cron apenas para novos commits de algum repositório {#templateForm-cronVisible}

Inicia uma tarefa deste modelo quando surge um novo commit num repositório. O Semaphore pede ao
repositório o seu último commit no intervalo escolhido e só inicia uma tarefa quando o commit
mudou desde a última verificação.

Assinalar a caixa e guardar cria a agenda da verificação; desmarcá-la e guardar elimina essa
agenda. As outras agendas do modelo são mantidas.

## Repositório {#templateForm-cronRepositoryId}

Apresentado quando **Quero executar uma tarefa pelo cron apenas para novos commits de algum
repositório** está assinalado. Obrigatório.

O repositório a vigiar para novos commits. O Semaphore vigia o ramo definido nesse repositório,
não o **Ramo** do modelo.

## intervaloDeVerificação {#templateForm-cronFormat}

Apresentado quando **Quero executar uma tarefa pelo cron apenas para novos commits de algum
repositório** está assinalado.

Com que frequência o Semaphore verifica se há um novo commit no repositório: de minuto a minuto
até uma vez por dia. A predefinição é de minuto a minuto.

## Suprimir alertas de sucesso {#templateForm-suppressSuccessAlerts}

Impede que as tarefas deste modelo concluídas com êxito enviem alertas. Os alertas das tarefas
falhadas continuam a ser enviados. Só produz efeito quando os alertas estão ativados para o
projeto.

## Suprimir alertas de erro {#templateForm-suppressErrorAlerts}

Impede que as tarefas deste modelo que falham enviem alertas. Os alertas das tarefas concluídas
com êxito continuam a ser enviados. Só produz efeito quando os alertas estão ativados para o
projeto.

## CLI args {#templateForm-args}

Argumentos acrescentados ao comando de cada tarefa. Cada chip é um argumento, passado tal como
está, sem shell: coloque `--forks=10` num chip, ou `--forks` e `10` em dois. Arraste os chips para
alterar a sua ordem. O botão à direita da lista abre os argumentos como um array JSON:

```json
["--forks=10", "-e", "env=staging"]
```

- **Ansible**: acrescentados ao `ansible-playbook`, antes do playbook.
- **Terraform**, **OpenTofu**, **Terragrunt**: acrescentados ao `plan` e ao `apply`, não ao
  `init`.
- Um script: acrescentados depois do nome do script.

## Argumentos CLI {#templateForm-allowOverrideArgsInTask}

Em **Solicitações**. Permite que o utilizador acrescente argumentos seus quando uma tarefa é
iniciada. São acrescentados depois dos argumentos do modelo.

## Ramo {#templateForm-allowOverrideBranchInTask}

Em **Solicitações**. Permite que o utilizador escolha outro ramo, tag ou commit quando uma tarefa é
iniciada, em vez do ramo do modelo. Sem esta opção, a tarefa obtém sempre o ramo do modelo.

A caixa fica desativada quando o repositório é um monorepo.

## Inventário {#templateForm-allowOverrideInventory}

Em **Solicitações**. Apresentado para modelos Ansible, Terraform, OpenTofu e Terragrunt.

Permite que o utilizador escolha outro inventário quando uma tarefa é iniciada. Apenas os modelos
Ansible utilizam o inventário aí escolhido: uma tarefa Terraform, OpenTofu ou Terragrunt é
executada no workspace do modelo mesmo quando a caixa está assinalada.

## Limite {#templateForm-limit}

Apresentado para modelos Ansible.

Os hosts ou grupos do inventário em que o playbook é executado, passados como `--limit`. Cada chip
é um padrão, por exemplo `web` ou `db01`; deixe a lista vazia para executar em todos os hosts.

## Tags {#templateForm-tags}

Apresentado para modelos Ansible.

Executa apenas as tarefas e as roles com estas tags, passadas como `--tags`. Deixe a lista vazia
para executar tudo.

## Pular tags {#templateForm-skipTags}

Apresentado para modelos Ansible.

Ignora as tarefas e as roles com estas tags, passadas como `--skip-tags`.

## Cofres {#templateVaults}

Apresentado para modelos Ansible.

As palavras-passe do Ansible Vault de que o playbook precisa para desencriptar os seus ficheiros.
Clique em **+** para adicionar uma, num chip para a alterar, na cruz de um chip para a remover.
Cada cofre tem um nome, o seu vault ID, e obtém a palavra-passe de uma chave do Armazenamento de
Chaves ou de um script cliente do repositório.

## Aprovação automática {#templateForm-autoApprove}

Apresentado para modelos Terraform, OpenTofu e Terragrunt.

Aplica o plano sem perguntar. Sem esta opção, uma tarefa para depois do plano e aguarda até que um
utilizador o confirme ou rejeite; um plano sem alterações termina a tarefa sem nada a aplicar.

## Override backend settings {#templateForm-overrideBackend}

Apresentado para modelos Terraform, OpenTofu e Terragrunt quando o Semaphore pode guardar o
estado do Terraform.

Guarda o estado do código no Semaphore. Antes do `init`, o Semaphore escreve no diretório do
código um ficheiro que define o backend `http`, e elimina-o depois da tarefa:

```hcl
terraform {
  backend "http" {
  }
}
```

## Backend filename {#templateForm-backendFilename}

Apresentado com **Override backend settings**; fica ativo quando essa caixa está assinalada.
Opcional.

O nome do ficheiro em que o Semaphore escreve o backend, por exemplo `semaphore_backend.tf`. A
predefinição é `backend.tf`. O nome tem de terminar em `.tf` e conter apenas letras, algarismos,
`_`, `-` e `.`. Escolha um nome que nenhum ficheiro do seu código tenha: o Semaphore substitui
esse ficheiro e elimina-o depois da tarefa.

## Limite {#templateForm-allowOverrideLimit}

Em **Solicitações de Ansible**. Permite que o utilizador defina os hosts em que a tarefa é
executada quando é iniciada, em vez do **Limite** do modelo.

## Tags {#templateForm-allowOverrideTags}

Em **Solicitações de Ansible**. Permite que o utilizador defina as tags quando uma tarefa é
iniciada, em vez das **Tags** do modelo.

## Pular tags {#templateForm-allowOverrideSkipTags}

Em **Solicitações de Ansible**. Permite que o utilizador defina as tags a ignorar quando uma
tarefa é iniciada, em vez de **Pular tags** do modelo.

## Depurar {#templateForm-allowDebug}

Em **Solicitações de Ansible**. Permite que o utilizador ative uma saída detalhada quando uma
tarefa é iniciada, de `-v` a `-vvvvvv`.

## Execução Simulada {#templateForm-showDryRun}

Em **Solicitações de Ansible**. Assinalado por predefinição. Disponibiliza a opção **Execução
Simulada** quando uma tarefa é iniciada, que executa o playbook com `--check`: o Ansible indica o
que alteraria e não altera nada. Desmarque a caixa para retirar a opção.

## Diferença {#templateForm-showDiff}

Em **Solicitações de Ansible**. Assinalado por predefinição. Disponibiliza a opção **Diferença**
quando uma tarefa é iniciada, que executa o playbook com `--diff` para mostrar as alterações feitas
aos ficheiros. Desmarque a caixa para retirar a opção.

## Aprovação automática {#templateForm-allowAutoApprove}

Nas solicitações de Terraform, OpenTofu e Terragrunt. Permite que o utilizador assinale **Auto
Approve** quando uma tarefa é iniciada, para aplicar o plano sem confirmação. A caixa de diálogo
da tarefa disponibiliza essa opção em qualquer caso, mas sem esta caixa o Semaphore ignora-a: a
tarefa aguarda uma confirmação depois do plano, salvo se a **Aprovação automática** do modelo
estiver assinalada.

## Galaxy install options {#templateForm-galaxy}

Apresentado para modelos Ansible. Clique nele para mostrar as opções da instalação Galaxy; a
etiqueta conta as opções que diferem das predefinições.

Antes de o playbook ser executado, o Semaphore instala com o `ansible-galaxy` as roles e as
coleções que o repositório pede. Procura `requirements.yml`, `roles/requirements.yml` e
`collections/requirements.yml` junto ao playbook e na raiz do repositório, e só executa a
instalação quando um ficheiro mudou desde a última tarefa.

## Skip Galaxy install {#templateForm-skipGalaxyInstall}

Em **Galaxy install options**. Ignora a instalação Galaxy: o playbook é executado com as roles e
as coleções instaladas anteriormente. Utilize-o quando os requisitos são instalados de outra forma
ou quando a instalação é lenta e desnecessária.

## Role install args {#templateForm-galaxyRoleArgs}

Em **Galaxy install options**. Argumentos acrescentados ao `ansible-galaxy role install`, um por
chip, por exemplo `--ignore-certs` ou `--server=https://galaxy.example.com`. O Semaphore só aceita
as flags da instalação que são seguras de passar: uma flag desconhecida, `--token`, `--api-key`,
`-p` ou `--roles-path` impede que o modelo seja guardado.

## Collection install args {#templateForm-galaxyCollectionArgs}

Em **Galaxy install options**. Argumentos acrescentados ao `ansible-galaxy collection install`, um
por chip, por exemplo `--pre` ou `--upgrade`. Aplica-se a mesma regra que para **Role install
args**: `--token`, `--api-key`, `-p` e `--collections-path` não são aceites.

## Skip Galaxy install {#templateForm-allowOverrideSkipGalaxyInstall}

Em **Galaxy install options**, em **Prompts**. Permite que o utilizador escolha, quando uma tarefa
é iniciada, se a instalação Galaxy é ignorada. A escolha da tarefa prevalece sobre **Skip Galaxy
install** do modelo.
