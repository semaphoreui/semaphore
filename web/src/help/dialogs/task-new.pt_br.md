---
form: src/components/TaskForm.vue, src/components/TaskParamsAnsibleForm.vue, src/components/TaskParamsTerraformForm.vue
---

# Nova Tarefa

Inicia uma tarefa do modelo: uma execução do seu playbook, código Terraform ou script. O modelo
decide quais campos esta caixa de diálogo mostra, e cada seção abaixo diz quando o seu campo
aparece. O botão na parte inferior é **Executar**, **Construir** ou **Implantar**, conforme o tipo
do modelo.

Quando o modelo tem uma pesquisa, as perguntas dela são campos depois de **Mensagem**. Um campo
marcado com `*` é obrigatório, um campo Integer aceita apenas dígitos, e as respostas padrão já vêm
preenchidas. O Semaphore passa as respostas para a tarefa como variáveis: extra vars para o
Ansible, `-var` para o Terraform, argumentos `name=value` para um script, ou variáveis de ambiente
para uma pergunta configurada assim. Uma resposta substitui uma variável com o mesmo nome nos
grupos de variáveis do modelo. Uma resposta secreta fica oculta enquanto você digita e não é
guardada com a tarefa.

## Commit {#newTask-commit}

Exibido quando você executa novamente uma tarefa anterior, e essa tarefa foi executada em um
commit conhecido.

A tarefa faz o checkout deste commit do repositório em vez do commit mais recente do ramo. Feche a
caixa para executar o commit mais recente.

O commit é usado apenas quando o modelo permite substituir o ramo, e nunca para um repositório
marcado como monorepo. Caso contrário, a tarefa executa o commit mais recente.

## Versão de Construção {#newTask-buildTaskId}

Exibido quando o tipo do modelo é **Implantar**.

A construção a implantar: uma das últimas 20 tarefas bem-sucedidas do modelo de construção,
listadas por versão e mensagem. A mais recente vem selecionada. O campo é obrigatório.

A versão da construção chega à tarefa: como `semaphore_vars.task_details.incoming_version` no
Ansible, como a variável de ambiente `SEMAPHORE_TASK_DETAILS_INCOMING_VERSION` em um script. A
lista fica vazia quando você não pode ver o modelo de construção.

## Mensagem {#newTask-message}

Opcional. Uma nota sobre por que a tarefa é executada, por exemplo `Hotfix for the login page`. Ela
é exibida com a tarefa no histórico de tarefas e não altera nada na execução.

O playbook pode lê-la como `semaphore_vars.task_details.message`, e um script como a variável de
ambiente `SEMAPHORE_TASK_DETAILS_MESSAGE`.

## Ramo {#newTask-gitBranch}

Exibido quando o modelo permite substituir o ramo, a menos que o repositório seja um monorepo.

O ramo ou a tag do repositório de que a tarefa faz o checkout. O campo começa com o ramo do modelo.
Quando está vazio, a tarefa usa o ramo do modelo, ou o ramo do repositório quando o modelo não tem
um. Um nome que não é um nome de ramo Git válido é recusado.

## Inventário {#newTask-inventoryId}

Exibido quando o modelo permite escolher o inventário em uma tarefa e você pode ver o projeto.
Para um modelo de Terraform, OpenTofu ou Terragrunt, o campo se chama **Workspace**, mas o servidor
usa a escolha apenas para modelos de Ansible: uma tarefa de Terraform, OpenTofu ou Terragrunt é
executada no workspace do modelo, seja qual for a escolha feita aqui.

O inventário contra o qual a tarefa é executada. O campo começa com o inventário do modelo e lista
os inventários do projeto feitos para o app do modelo. Quando o inventário escolhido não existe
mais no momento em que a tarefa é iniciada, a tarefa é executada contra o inventário do modelo.

## Limite {#taskParams-limit}

Exibido para um modelo de Ansible que permite substituir o limite.

Os hosts ou grupos do inventário em que o playbook é executado, um padrão por chip, por exemplo
`web` ou `10.0.0.11`. O Semaphore os passa como `--limit`, unidos por vírgulas. A lista começa com
o limite do modelo. Remova todos os chips para executar em todos os hosts do inventário.

## Tags {#taskParams-tags}

Exibido para um modelo de Ansible que permite substituir as tags.

Apenas as tarefas e roles do playbook com uma destas tags são executadas: o Semaphore as passa como
`--tags`, uma tag por chip. A lista começa com as tags do modelo. Remova todos os chips para
executar o playbook inteiro.

## Pular tags {#taskParams-skipTags}

Exibido para um modelo de Ansible que permite substituir as tags puladas.

As tarefas e roles do playbook com uma destas tags não são executadas: o Semaphore as passa como
`--skip-tags`, uma tag por chip. A lista começa com as tags puladas do modelo.

## Depurar {#taskParams-debug}

Exibido para um modelo de Ansible que permite depuração.

Executa o `ansible-playbook` com uma flag de verbosidade, para que o log da tarefa mostre mais
detalhes. O rótulo mostra a flag: `-v` tantas vezes quanto indica o **Nível de depuração**.

## Nível de depuração {#taskParams-debugLevel}

Exibido com **Depurar**, e habilitado quando **Depurar** está marcado.

Quão detalhado é o log, de 1 (`-v`) a 6 (`-vvvvvv`). O padrão é 4 (`-vvvv`), que também mostra os
detalhes das conexões com os hosts.

## Execução Simulada {#taskParams-dryRun}

Exibido para um modelo de Ansible, a menos que o modelo o oculte.

Executa o playbook com `--check`: o Ansible informa o que alteraria nos hosts e não altera nada.

## Diferença {#taskParams-diff}

Exibido para um modelo de Ansible, a menos que o modelo o oculte.

Executa o playbook com `--diff`: o log mostra as alterações feitas em arquivos e templates nos
hosts. Marque também **Execução Simulada** para ver as alterações sem fazê-las.

## Skip Galaxy install {#taskParams-skipGalaxyInstall}

Exibido para um modelo de Ansible que permite substituir a instalação do Galaxy.

Antes de o playbook ser executado, o Semaphore instala as roles e coleções listadas nos arquivos
`requirements.yml` do repositório, quando um arquivo mudou desde a última instalação. Marque a
caixa para pular essa etapa nesta tarefa.

A caixa começa desmarcada, seja qual for a configuração do modelo, e uma caixa desmarcada instala
os requisitos.

## Plan {#taskParams-plan}

Exibido para um modelo de Terraform, OpenTofu ou Terragrunt.

Executa o `plan` e encerra a tarefa ali: nada é aplicado. Use para revisar as alterações primeiro.
Sem ele, a tarefa prossegue e aplica o plano quando ele tem alterações.

## Destroy {#taskParams-destroy}

Exibido para um modelo de Terraform, OpenTofu ou Terragrunt.

Adiciona `-destroy` ao `plan` e ao `apply`: a tarefa planeja, e depois aplica, a remoção de todos
os recursos no estado do workspace. Marque também **Plan** para ver o que seria destruído sem
destruir.

## Auto Approve {#taskParams-autoApprove}

Exibido para um modelo de Terraform, OpenTofu ou Terragrunt.

Aplica o plano assim que ele é feito. Sem ele, a tarefa para depois do plano e aguarda que alguém
o confirme ou rejeite na página da tarefa.

A caixa só tem efeito quando o modelo permite a aprovação automática. Um modelo configurado para
aprovar sempre aplica sem perguntar, seja qual for o estado da caixa.

## Upgrade {#taskParams-upgrade}

Exibido para um modelo de Terraform, OpenTofu ou Terragrunt.

Executa o `init` com `-upgrade`: os providers e módulos são atualizados para as versões mais
recentes que as suas restrições de versão permitem.

## Reconfigure {#taskParams-reconfigure}

Exibido para um modelo de Terraform, OpenTofu ou Terragrunt.

Executa o `init` com `-reconfigure`: o backend é configurado de novo, e a configuração salva do
backend é ignorada. Sem ele, o Semaphore executa `init -migrate-state`, que move o estado
existente para um backend alterado.

## CLI args {#newTask-args}

Exibido quando o modelo permite substituir os argumentos CLI.

Argumentos extras para a linha de comando da tarefa, um por chip: `--forks=10` é um chip,
`--forks` e `10` são dois. Clique em um chip para editá-lo, arraste-o para mudar a ordem e use o
último chip para adicionar um.

A lista começa com os argumentos do modelo. O Semaphore sempre passa os argumentos do modelo e
adiciona esta lista depois deles, então, depois que você altera a lista, os argumentos do modelo
que ficaram nela são passados duas vezes. Remova-os e mantenha apenas os argumentos que você
adiciona.
