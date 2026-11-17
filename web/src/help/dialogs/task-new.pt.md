---
form: src/components/TaskForm.vue, src/components/TaskParamsAnsibleForm.vue, src/components/TaskParamsTerraformForm.vue
---

# Nova Tarefa

Inicia uma tarefa do modelo: uma execução do seu playbook, código Terraform ou script. O modelo
decide que campos esta caixa de diálogo apresenta, e cada secção abaixo indica quando o seu campo
está presente. O botão no fundo é **Executar**, **Construir** ou **Implantar**, consoante o tipo
do modelo.

Quando o modelo tem uma pesquisa, as suas perguntas são campos a seguir a **Mensagem**. Um campo
marcado com `*` é obrigatório, um campo Integer aceita apenas algarismos, e as respostas
predefinidas vêm preenchidas. O Semaphore passa as respostas à tarefa como variáveis: variáveis
extra para o Ansible, `-var` para o Terraform, argumentos `name=value` para um script, ou
variáveis de ambiente para uma pergunta assim configurada. Uma resposta substitui uma variável
com o mesmo nome nos grupos de variáveis do modelo. Uma resposta secreta fica oculta enquanto a
escreve e não é guardada com a tarefa.

## Commit {#newTask-commit}

Apresentado quando volta a executar uma tarefa anterior, e essa tarefa foi executada num commit
conhecido.

A tarefa faz checkout deste commit do repositório em vez do commit mais recente do ramo. Feche a
caixa para executar o commit mais recente.

O commit só é utilizado quando o modelo permite substituir o ramo, e nunca para um repositório
marcado como monorepo. Caso contrário, a tarefa executa o commit mais recente.

## Versão de Construção {#newTask-buildTaskId}

Apresentado quando o tipo do modelo é **Implantar**.

A construção a implantar: uma das últimas 20 tarefas bem-sucedidas do modelo de construção,
listadas por versão e mensagem. A mais recente vem selecionada. O campo é obrigatório.

A versão da construção chega à tarefa: como `semaphore_vars.task_details.incoming_version` no
Ansible, como a variável de ambiente `SEMAPHORE_TASK_DETAILS_INCOMING_VERSION` num script. A
lista fica vazia quando não pode ver o modelo de construção.

## Mensagem {#newTask-message}

Opcional. Uma nota sobre o motivo da execução da tarefa, por exemplo `Hotfix for the login page`.
É apresentada com a tarefa no histórico de tarefas e não altera nada na execução.

O playbook pode lê-la como `semaphore_vars.task_details.message`, e um script como a variável de
ambiente `SEMAPHORE_TASK_DETAILS_MESSAGE`.

## Ramo {#newTask-gitBranch}

Apresentado quando o modelo permite substituir o ramo, a menos que o repositório seja um
monorepo.

O ramo ou a tag do repositório de que a tarefa faz checkout. O campo começa com o ramo do modelo.
Quando está vazio, a tarefa utiliza o ramo do modelo, ou o ramo do repositório quando o modelo
não tem nenhum. Um nome que não seja um nome de ramo Git válido é recusado.

## Inventário {#newTask-inventoryId}

Apresentado quando o modelo permite escolher o inventário numa tarefa e pode ver o projeto. Num
modelo Terraform, OpenTofu ou Terragrunt, o campo chama-se **Workspace**, mas o servidor só usa a
escolha para modelos Ansible: uma tarefa Terraform, OpenTofu ou Terragrunt é executada no workspace
do modelo, seja o que for que escolha aqui.

O inventário sobre o qual a tarefa é executada. O campo começa com o inventário do modelo e lista
os inventários do projeto criados para a aplicação do modelo. Quando o inventário escolhido já não
existe no momento em que a tarefa é iniciada, a tarefa é executada sobre o inventário do modelo.

## Limite {#taskParams-limit}

Apresentado num modelo Ansible que permite substituir o limite.

Os hosts ou grupos do inventário em que o playbook é executado, um padrão por chip, por exemplo
`web` ou `10.0.0.11`. O Semaphore passa-os como `--limit`, separados por vírgulas. A lista começa
com o limite do modelo. Remova todos os chips para executar em todos os hosts do inventário.

## Tags {#taskParams-tags}

Apresentado num modelo Ansible que permite substituir as tags.

Só são executadas as tarefas e roles do playbook com uma destas tags: o Semaphore passa-as como
`--tags`, uma tag por chip. A lista começa com as tags do modelo. Remova todos os chips para
executar o playbook inteiro.

## Pular tags {#taskParams-skipTags}

Apresentado num modelo Ansible que permite substituir as tags ignoradas.

As tarefas e roles do playbook com uma destas tags não são executadas: o Semaphore passa-as como
`--skip-tags`, uma tag por chip. A lista começa com as tags ignoradas do modelo.

## Depurar {#taskParams-debug}

Apresentado num modelo Ansible que permite a depuração.

Executa o `ansible-playbook` com uma opção de verbosidade, para que o log da tarefa mostre mais
detalhe. O rótulo mostra a opção: `-v` tantas vezes quantas indicar **Nível de depuração**.

## Nível de depuração {#taskParams-debugLevel}

Apresentado com **Depurar**, e ativo quando **Depurar** está assinalado.

O grau de detalhe do log, de 1 (`-v`) a 6 (`-vvvvvv`). A predefinição é 4 (`-vvvv`), que mostra
também os detalhes das ligações aos hosts.

## Execução Simulada {#taskParams-dryRun}

Apresentado num modelo Ansible, a menos que o modelo o oculte.

Executa o playbook com `--check`: o Ansible indica o que alteraria nos hosts e não altera nada.

## Diferença {#taskParams-diff}

Apresentado num modelo Ansible, a menos que o modelo o oculte.

Executa o playbook com `--diff`: o log mostra as alterações feitas a ficheiros e templates nos
hosts. Assinale também **Execução Simulada** para ver as alterações sem as fazer.

## Skip Galaxy install {#taskParams-skipGalaxyInstall}

Apresentado num modelo Ansible que permite substituir a instalação do Galaxy.

Antes de o playbook ser executado, o Semaphore instala as roles e coleções indicadas nos
ficheiros `requirements.yml` do repositório, quando um ficheiro foi alterado desde a última
instalação. Assinale a caixa para saltar esse passo nesta tarefa.

A caixa começa desmarcada, seja qual for a configuração do modelo, e uma caixa desmarcada instala
os requisitos.

## Plan {#taskParams-plan}

Apresentado num modelo Terraform, OpenTofu ou Terragrunt.

Executa `plan` e termina a tarefa aí: nada é aplicado. Utilize-o para rever primeiro as
alterações. Sem esta opção, a tarefa continua e aplica o plano quando este tem alterações.

## Destroy {#taskParams-destroy}

Apresentado num modelo Terraform, OpenTofu ou Terragrunt.

Acrescenta `-destroy` a `plan` e `apply`: a tarefa planeia e depois aplica a remoção de todos os
recursos no estado do workspace. Assinale também **Plan** para ver o que seria destruído sem o
destruir.

## Auto Approve {#taskParams-autoApprove}

Apresentado num modelo Terraform, OpenTofu ou Terragrunt.

Aplica o plano logo que é criado. Sem esta opção, a tarefa para depois do plano e aguarda que
alguém o confirme ou rejeite na página da tarefa.

A caixa só produz efeito quando o modelo permite a aprovação automática. Um modelo configurado
para aprovar aplica sempre sem perguntar, seja qual for o estado da caixa.

## Upgrade {#taskParams-upgrade}

Apresentado num modelo Terraform, OpenTofu ou Terragrunt.

Executa `init` com `-upgrade`: os providers e módulos são atualizados para as versões mais
recentes que as suas restrições de versão permitem.

## Reconfigure {#taskParams-reconfigure}

Apresentado num modelo Terraform, OpenTofu ou Terragrunt.

Executa `init` com `-reconfigure`: o backend é configurado de novo, e a configuração guardada do
backend é ignorada. Sem esta opção, o Semaphore executa `init -migrate-state`, que move o estado
existente para um backend alterado.

## CLI args {#newTask-args}

Apresentado quando o modelo permite substituir os argumentos CLI.

Argumentos extra para a linha de comandos da tarefa, um por chip: `--forks=10` é um chip,
`--forks` e `10` são dois. Clique num chip para o editar, arraste-o para mudar a ordem e utilize o
último chip para acrescentar um.

A lista começa com os argumentos do modelo. O Semaphore passa sempre os argumentos do modelo e
acrescenta esta lista a seguir a eles, por isso, depois de alterar a lista, os argumentos do
modelo que nela ficarem são passados duas vezes. Remova-os e mantenha apenas os argumentos que
acrescenta.
