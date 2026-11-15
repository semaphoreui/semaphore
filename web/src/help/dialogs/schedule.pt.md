---
form: src/components/ScheduleForm.vue, src/components/TaskParamsForm.vue, src/components/TaskParamsAnsibleForm.vue, src/components/TaskParamsTerraformForm.vue
---

# Editar Agenda

Uma agenda inicia uma tarefa de um modelo de tarefa sem que ninguém prima **Executar**:
repetidamente, segundo uma expressão cron, ou uma única vez num momento definido. A tarefa
recebe os parâmetros guardados nesta caixa de diálogo. Todas as horas estão no fuso horário do
servidor Semaphore, apresentado junto aos campos de hora: `UTC`, salvo se o administrador definir
`schedule.timezone` ou `SEMAPHORE_SCHEDULE_TIMEZONE`.

## Name {#scheduleForm-name}

O nome com que a agenda aparece na página Agenda. O campo é obrigatório. Não altera o que a
agenda faz.

## Template {#scheduleForm-templateId}

O modelo de tarefa que a agenda inicia. O campo é obrigatório. A tarefa é executada como se
tivesse sido iniciada a partir do modelo, com o inventário, o repositório, o grupo de variáveis e
as restantes definições do próprio modelo, e com os parâmetros abaixo, nos casos em que o modelo
permite que uma tarefa os altere.

## Parâmetros da tarefa {#scheduleForm-taskParams}

Apresentado quando está selecionado um modelo.

Os parâmetros de todas as tarefas que a agenda inicia. Os campos apresentados dependem do modelo:
as suas variáveis de pesquisa, e as opções que permite que uma tarefa altere, como o ramo, o
inventário, **Limite**, **Tags**, **Pular tags** ou **Argumentos CLI**, e as opções do
Ansible ou do Terraform, OpenTofu e Terragrunt.

O servidor verifica o modelo sempre que a agenda é acionada: um valor guardado aqui é ignorado
quando o modelo deixa de permitir que uma tarefa o altere.

## Mensagem {#taskParams-message}

Opcional. Uma nota guardada com cada tarefa que a agenda inicia e apresentada no histórico de
tarefas, por exemplo `Nightly deploy`. Não altera a forma como a tarefa é executada.

## Variáveis de Pesquisa {#taskParams-survey}

Apresentado quando o modelo tem variáveis de pesquisa.

Os valores das variáveis de pesquisa do modelo para as tarefas agendadas. Chegam à tarefa da
mesma forma que quando alguém preenche o questionário manualmente. Uma variável obrigatória tem
de ter um valor antes de a agenda poder ser guardada.

As predefinições do modelo são copiadas quando a caixa de diálogo abre e são guardadas com a
agenda. Uma predefinição alterada mais tarde no modelo não chega a uma agenda guardada antes.

O valor de uma variável do tipo secret não é guardado com a agenda: a tarefa agendada é executada
sem ele.

## Ramo {#taskParams-gitBranch}

Apresentado quando o modelo permite que uma tarefa substitua o ramo.

O ramo ou a tag do repositório que a tarefa agendada obtém. O campo começa com o ramo do modelo.
O servidor só utiliza o valor enquanto o modelo permitir a substituição.

## Inventário {#taskParams-inventoryId}

Apresentado quando o modelo permite que uma tarefa escolha o inventário.

O inventário sobre o qual a tarefa agendada é executada, em vez do inventário do modelo. Num
modelo Terraform ou OpenTofu o campo chama-se **Workspace**, mas o servidor só aplica a escolha
aos modelos Ansible e mantém o workspace do modelo para os restantes.

## Depurar {#taskParams-debug}

Apresentado para um modelo Ansible que permite a depuração.

Executa o `ansible-playbook` com mais saída: `-v` repetido tantas vezes quantas as definidas em
**Nível de depuração**.

## Nível de depuração {#taskParams-debugLevel}

Apresentado para um modelo Ansible que permite a depuração; ativo quando **Depurar** está
assinalado.

O grau de detalhe da saída, de `1` (`-v`) a `6` (`-vvvvvv`). A predefinição é `4`.

## Execução Simulada {#taskParams-dryRun}

Apresentado para um modelo Ansible, salvo se o modelo o ocultar.

Executa o playbook com `--check`: o Ansible indica o que alteraria e não altera nada.

## Diferença {#taskParams-diff}

Apresentado para um modelo Ansible, salvo se o modelo o ocultar.

Executa o playbook com `--diff`: o Ansible mostra as alterações que faz aos ficheiros e aos
templates.

## Skip Galaxy install {#taskParams-skipGalaxyInstall}

Apresentado para um modelo Ansible que permite que uma tarefa o altere.

Ignora a instalação das roles e das coleções dos ficheiros `requirements.yml` antes de o playbook
ser executado. Assinale-o quando os requisitos já estão instalados e a tarefa não deve aceder ao
Ansible Galaxy.

## Plan {#taskParams-plan}

Apresentado para um modelo Terraform, OpenTofu ou Terragrunt.

Executa apenas o `plan`. A tarefa termina depois do plano e não aplica nada.

## Destroy {#taskParams-destroy}

Apresentado para um modelo Terraform, OpenTofu ou Terragrunt.

Acrescenta `-destroy` ao plano e à aplicação: a tarefa destrói os recursos do workspace em vez de
os criar ou atualizar.

## Auto Approve {#taskParams-autoApprove}

Apresentado para um modelo Terraform, OpenTofu ou Terragrunt.

Aplica o plano sem aguardar uma confirmação. Só funciona quando o modelo permite a aprovação
automática; um modelo configurado para aprovar todas as tarefas aplica o plano em qualquer caso.

Sem aprovação automática, uma tarefa agendada fica parada depois do plano e aguarda até que
alguém o confirme ou rejeite na página da tarefa.

## Upgrade {#taskParams-upgrade}

Apresentado para um modelo Terraform, OpenTofu ou Terragrunt.

Executa o `init` com `-upgrade`: os providers e os módulos são atualizados para as versões mais
recentes que a configuração permite.

## Reconfigure {#taskParams-reconfigure}

Apresentado para um modelo Terraform, OpenTofu ou Terragrunt.

Executa o `init` com `-reconfigure` em vez de `-migrate-state`: o backend é configurado de novo e
o estado existente não é migrado.

## Run at {#scheduleForm-runAt}

Apresentado para uma agenda que é executada uma única vez.

A data e a hora em que a tarefa é iniciada, no fuso horário do servidor apresentado no fim do
campo. Uma nova agenda começa com a próxima hora certa. O momento tem de estar no futuro: o
servidor recusa uma hora que já passou.

Depois de a tarefa ter sido iniciada, a agenda é desativada, ou eliminada quando **Delete after
run** está assinalado. Quando o Semaphore não está em execução nesse momento, a tarefa não é
iniciada mais tarde: no arranque seguinte, a agenda é desativada, ou eliminada, sem execução.

## Delete after run {#scheduleForm-deleteAfterRun}

Apresentado para uma agenda que é executada uma única vez.

Elimina a agenda depois de a sua tarefa ter sido iniciada. Deixe-o por assinalar para manter a
agenda, desativada, de modo a poder definir uma nova hora e voltar a ativá-la.

## Show cron format {#scheduleForm-rawCron}

Apresentado para uma agenda cron.

Alterna entre escrever a expressão cron (ligado) e construí-la a partir de **Timing** e das caixas
de seleção de hora (desligado). Ambos editam a mesma expressão. O navegador memoriza a escolha.

O interruptor fica bloqueado na posição ligada quando a expressão não pode ser apresentada como
caixas de seleção: quando é inválida, ou utiliza uma forma como `@daily` ou `CRON_TZ=`.

## Cron {#scheduleForm-cronFormat}

Apresentado para uma agenda cron quando **Show cron format** está ligado.

A expressão cron em cinco campos: minuto, hora, dia do mês, mês, dia da semana (`0` é domingo). O
campo é obrigatório, e o servidor verifica a expressão à medida que escreve. Uma nova agenda
começa com `* * * * *`, a cada minuto.

```
0 3 * * *       every day at 03:00
*/15 * * * *    every 15 minutes
30 8 * * 1-5    at 08:30 from Monday to Friday
0 0 1 * *       at midnight on the first day of every month
```

O servidor aceita também `@hourly`, `@daily`, `@weekly`, `@monthly`, `@yearly` e
`@every 1h30m`, e um prefixo como `CRON_TZ=Europe/Berlin` para outro fuso horário. A
pré-visualização abaixo não consegue mostrar a próxima execução destas formas.

## Timing {#scheduleForm-timing}

Apresentado para uma agenda cron quando **Show cron format** está desligado.

Com que frequência a tarefa é executada. As caixas de seleção abaixo mudam em conformidade:

- **Hourly**: **Minutes**.
- **Daily**: **Hours** e **Minutes**.
- **Weekly**: **Weekdays**, **Hours** e **Minutes**.
- **Monthly**: **Days**, **Hours** e **Minutes**.
- **Yearly**: **Months**, **Days**, **Hours** e **Minutes**.

## Months {#scheduleForm-months}

Apresentado quando **Timing** é **Yearly**.

Os meses em que a tarefa é executada. Assinale vários para a executar em cada um deles. Sem nenhum
assinalado, a tarefa é executada em todos os meses.

## Weekdays {#scheduleForm-weekdays}

Apresentado quando **Timing** é **Weekly**.

Os dias da semana em que a tarefa é executada. Sem nenhum assinalado, a tarefa é executada todos
os dias.

## Days {#scheduleForm-days}

Apresentado quando **Timing** é **Monthly** ou **Yearly**.

Os dias do mês em que a tarefa é executada. Um mês sem esse dia é ignorado: uma tarefa definida
para o dia 31 não é executada em abril. Sem nenhum assinalado, a tarefa é executada todos os dias.

## Hours {#scheduleForm-hours}

Apresentado quando **Timing** é **Daily**, **Weekly**, **Monthly** ou **Yearly**.

As horas a que a tarefa é executada, no fuso horário apresentado à direita. Sem nenhuma
assinalada, a tarefa é executada a todas as horas.

## Minutes {#scheduleForm-minutes}

Apresentado para uma agenda cron quando **Show cron format** está desligado.

Os minutos da hora a que a tarefa é executada, em intervalos de cinco. Sem nenhum assinalado, a
tarefa é executada a cada minuto: assinale pelo menos um, por exemplo `:00`, para uma tarefa que
deva ser executada uma vez por hora ou uma vez por dia.

## Next run time {#scheduleForm-nextRun}

Quando seria iniciada a próxima tarefa, segundo a agenda tal como está no formulário: no fuso
horário do servidor e na hora local do seu navegador. Um traço significa que a expressão é
inválida, ou é uma que a pré-visualização não consegue ler, como `@daily`.

## Habilitado {#scheduleForm-active}

Se a agenda inicia tarefas. Desmarque-o para manter uma agenda sem a executar. O interruptor na
lista de agendas faz o mesmo. Uma nova agenda está habilitada.
