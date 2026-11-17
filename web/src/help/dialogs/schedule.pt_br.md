---
form: src/components/ScheduleForm.vue, src/components/TaskParamsForm.vue, src/components/TaskParamsAnsibleForm.vue, src/components/TaskParamsTerraformForm.vue
---

# Editar Agenda

Uma agenda inicia uma tarefa de um modelo de tarefa sem que ninguém clique em **Executar**:
repetidamente, por uma expressão cron, ou uma única vez em um momento definido. A tarefa recebe os
parâmetros salvos nesta caixa de diálogo. Todos os horários estão no fuso horário do servidor
Semaphore, exibido ao lado dos campos de horário: `UTC`, a menos que o administrador defina
`schedule.timezone` ou `SEMAPHORE_SCHEDULE_TIMEZONE`.

## Name {#scheduleForm-name}

O nome com que a agenda aparece na página Agenda. O campo é obrigatório. Ele não altera o que a
agenda faz.

## Template {#scheduleForm-templateId}

O modelo de tarefa que a agenda inicia. O campo é obrigatório. A tarefa é executada como se tivesse
sido iniciada a partir do modelo, com o inventário, o repositório, o grupo de variáveis e as
outras configurações do próprio modelo, e com os parâmetros abaixo onde o modelo permite que uma
tarefa os altere.

## Parâmetros da tarefa {#scheduleForm-taskParams}

Exibido quando um modelo está selecionado.

Os parâmetros de cada tarefa que a agenda inicia. Os campos que aparecem dependem do modelo: as
variáveis de pesquisa dele e as opções que ele permite que uma tarefa altere, como o ramo, o
inventário, **Limite**, **Tags**, **Pular tags** ou **CLI args**, e as opções do Ansible ou do
Terraform, do OpenTofu e do Terragrunt.

O servidor verifica o modelo toda vez que a agenda dispara: um valor salvo aqui é ignorado assim
que o modelo deixa de permitir que uma tarefa o altere.

## Mensagem {#taskParams-message}

Opcional. Uma nota armazenada com cada tarefa que a agenda inicia e exibida no histórico de
tarefas, por exemplo `Nightly deploy`. Ela não altera a forma como a tarefa é executada.

## Variáveis de Pesquisa {#taskParams-survey}

Exibido quando o modelo tem variáveis de pesquisa.

Os valores das variáveis de pesquisa do modelo para as tarefas agendadas. Eles chegam à tarefa da
mesma forma que quando alguém preenche a pesquisa manualmente. Uma variável obrigatória precisa ter
um valor antes que a agenda possa ser salva.

Os padrões do modelo são copiados quando a caixa de diálogo abre e são salvos com a agenda. Um
padrão alterado no modelo depois não chega a uma agenda salva antes.

O valor de uma variável do tipo secret não é salvo com a agenda: a tarefa agendada é executada sem
ele.

## Ramo {#taskParams-gitBranch}

Exibido quando o modelo permite que uma tarefa substitua o ramo.

O ramo ou a tag do repositório de que a tarefa agendada faz o checkout. O campo começa com o ramo do
modelo. O servidor usa o valor apenas enquanto o modelo permitir a substituição.

## Inventário {#taskParams-inventoryId}

Exibido quando o modelo permite que uma tarefa escolha o inventário.

O inventário contra o qual a tarefa agendada é executada, no lugar do inventário do modelo. Para um
modelo de Terraform ou OpenTofu, o campo se chama **Workspace**, mas o servidor aplica a escolha
apenas aos modelos de Ansible e mantém o workspace do modelo para os outros.

## Depurar {#taskParams-debug}

Exibido para um modelo de Ansible que permite depuração.

Executa o `ansible-playbook` com mais saída: `-v` repetido tantas vezes quanto define o **Nível de
depuração**.

## Nível de depuração {#taskParams-debugLevel}

Exibido para um modelo de Ansible que permite depuração; ativo quando **Depurar** está marcado.

Quão detalhada é a saída, de `1` (`-v`) a `6` (`-vvvvvv`). O padrão é `4`.

## Execução Simulada {#taskParams-dryRun}

Exibido para um modelo de Ansible, a menos que o modelo o oculte.

Executa o playbook com `--check`: o Ansible informa o que alteraria e não altera nada.

## Diferença {#taskParams-diff}

Exibido para um modelo de Ansible, a menos que o modelo o oculte.

Executa o playbook com `--diff`: o Ansible mostra as alterações que faz em arquivos e templates.

## Skip Galaxy install {#taskParams-skipGalaxyInstall}

Exibido para um modelo de Ansible que permite que uma tarefa o altere.

Pula a instalação das roles e coleções dos arquivos `requirements.yml` antes de o playbook ser
executado. Marque quando os requisitos já estiverem instalados e a tarefa não deve acessar o
Ansible Galaxy.

## Plan {#taskParams-plan}

Exibido para um modelo de Terraform, OpenTofu ou Terragrunt.

Executa apenas o `plan`. A tarefa termina depois do plano e não aplica nada.

## Destroy {#taskParams-destroy}

Exibido para um modelo de Terraform, OpenTofu ou Terragrunt.

Adiciona `-destroy` ao plan e ao apply: a tarefa destrói os recursos do workspace em vez de
criá-los ou atualizá-los.

## Auto Approve {#taskParams-autoApprove}

Exibido para um modelo de Terraform, OpenTofu ou Terragrunt.

Aplica o plano sem aguardar uma confirmação. Funciona apenas quando o modelo permite a aprovação
automática; um modelo configurado para aprovar todas as tarefas aplica o plano em ambos os casos.

Sem a aprovação automática, uma tarefa agendada para depois do plano e aguarda até que alguém a
confirme ou rejeite na página da tarefa.

## Upgrade {#taskParams-upgrade}

Exibido para um modelo de Terraform, OpenTofu ou Terragrunt.

Executa o `init` com `-upgrade`: os providers e módulos são atualizados para as versões mais
recentes que a configuração permite.

## Reconfigure {#taskParams-reconfigure}

Exibido para um modelo de Terraform, OpenTofu ou Terragrunt.

Executa o `init` com `-reconfigure` em vez de `-migrate-state`: o backend é configurado de novo e
o estado existente não é migrado.

## Run at {#scheduleForm-runAt}

Exibido para uma agenda que é executada uma única vez.

A data e a hora em que a tarefa é iniciada, no fuso horário do servidor exibido no final do campo.
Uma nova agenda começa com a próxima hora cheia. O momento precisa estar no futuro: o servidor
recusa um horário que já passou.

Depois que a tarefa é iniciada, a agenda é desativada, ou excluída quando **Delete after run** está
marcado. Quando o Semaphore não está em execução nesse momento, a tarefa não é iniciada depois: na
próxima inicialização, a agenda é desativada, ou excluída, sem execução.

## Delete after run {#scheduleForm-deleteAfterRun}

Exibido para uma agenda que é executada uma única vez.

Exclui a agenda depois que a tarefa dela é iniciada. Deixe desmarcado para manter a agenda,
desativada, de modo que você possa definir um novo horário e ativá-la novamente.

## Show cron format {#scheduleForm-rawCron}

Exibido para uma agenda cron.

Alterna entre digitar a expressão cron (ativado) e montá-la a partir de **Timing** e das caixas de
seleção de horário (desativado). Ambos editam a mesma expressão. O navegador lembra a escolha.

O interruptor fica travado em ativado quando a expressão não pode ser exibida como caixas de
seleção: quando é inválida, ou usa uma forma como `@daily` ou `CRON_TZ=`.

## Cron {#scheduleForm-cronFormat}

Exibido para uma agenda cron quando **Show cron format** está ativado.

A expressão cron em cinco campos: minuto, hora, dia do mês, mês, dia da semana (`0` é domingo). O
campo é obrigatório, e o servidor verifica a expressão enquanto você digita. Uma nova agenda começa
com `* * * * *`, a cada minuto.

```
0 3 * * *       every day at 03:00
*/15 * * * *    every 15 minutes
30 8 * * 1-5    at 08:30 from Monday to Friday
0 0 1 * *       at midnight on the first day of every month
```

O servidor também aceita `@hourly`, `@daily`, `@weekly`, `@monthly`, `@yearly` e `@every 1h30m`, e
um prefixo como `CRON_TZ=Europe/Berlin` para outro fuso horário. A prévia abaixo não consegue
mostrar a próxima execução destes.

## Timing {#scheduleForm-timing}

Exibido para uma agenda cron quando **Show cron format** está desativado.

Com que frequência a tarefa é executada. As caixas de seleção abaixo mudam de acordo:

- **Hourly**: **Minutes**.
- **Daily**: **Hours** e **Minutes**.
- **Weekly**: **Weekdays**, **Hours** e **Minutes**.
- **Monthly**: **Days**, **Hours** e **Minutes**.
- **Yearly**: **Months**, **Days**, **Hours** e **Minutes**.

## Months {#scheduleForm-months}

Exibido quando **Timing** é **Yearly**.

Os meses em que a tarefa é executada. Marque vários para executar em cada um deles. Sem nenhum
marcado, a tarefa é executada em todos os meses.

## Weekdays {#scheduleForm-weekdays}

Exibido quando **Timing** é **Weekly**.

Os dias da semana em que a tarefa é executada. Sem nenhum marcado, a tarefa é executada todos os
dias.

## Days {#scheduleForm-days}

Exibido quando **Timing** é **Monthly** ou **Yearly**.

Os dias do mês em que a tarefa é executada. Um mês que não tem o dia é pulado: uma tarefa definida
para o dia 31 não é executada em abril. Sem nenhum marcado, a tarefa é executada todos os dias.

## Hours {#scheduleForm-hours}

Exibido quando **Timing** é **Daily**, **Weekly**, **Monthly** ou **Yearly**.

As horas em que a tarefa é executada, no fuso horário exibido à direita. Sem nenhuma marcada, a
tarefa é executada a cada hora.

## Minutes {#scheduleForm-minutes}

Exibido para uma agenda cron quando **Show cron format** está desativado.

Os minutos da hora em que a tarefa é executada, de cinco em cinco. Sem nenhum marcado, a tarefa é
executada a cada minuto: marque pelo menos um, por exemplo `:00`, para uma tarefa que deve ser
executada uma vez por hora ou uma vez por dia.

## Next run time {#scheduleForm-nextRun}

Quando a próxima tarefa seria iniciada, pela agenda como está no formulário: no fuso horário do
servidor e no horário local do seu navegador. Um traço significa que a expressão é inválida, ou é
uma que a prévia não consegue ler, como `@daily`.

## Habilitado {#scheduleForm-active}

Se a agenda inicia tarefas. Desmarque para manter uma agenda sem executá-la. O interruptor na lista
de agendas faz o mesmo. Uma nova agenda vem habilitada.
