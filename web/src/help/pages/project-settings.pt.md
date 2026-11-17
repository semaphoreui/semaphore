---
form: src/views/project/Settings.vue, src/components/ProjectForm.vue
skip: newProject-demo
---

# Configurações do Projeto

As configurações do projeto no seu conjunto: o nome, quantas das suas tarefas podem ser executadas
em simultâneo e se envia alertas. A Zona de Perigo, abaixo do formulário, reúne as ações que
abrangem todo o projeto: o backup, a cache e a eliminação do projeto. As alterações do formulário
produzem efeito depois de **Salvar**.

## Nome do Projeto {#newProject-name}

O nome com que o projeto aparece no menu de projetos, no topo da barra lateral. O campo é
obrigatório.

## Número máximo de tarefas paralelas {#newProject-maxParallelTasks}

Quantas tarefas deste projeto podem ser executadas em simultâneo. Uma tarefa iniciada acima do
limite fica em espera na fila até que uma das tarefas em execução termine.

Deixe o campo vazio ou defina-o como `0` para não ter um limite próprio. O limite de todo o
servidor, definido pelo respetivo administrador, aplica-se em ambos os casos.

## ID do Chat do Telegram {#newProject-tg}

Opcional. O chat do Telegram que recebe os alertas deste projeto. Quando o campo está vazio, os
alertas são enviados para o chat definido para todo o servidor.

O campo só produz efeito quando o administrador ativou os alertas do Telegram no servidor e os
alertas estão permitidos para este projeto.

## Permitir alertas para este projeto {#newProject-alert}

Ativa os alertas para as tarefas deste projeto. O destino dos alertas é definido para todo o
servidor pelo respetivo administrador: e-mail, Telegram, Slack, Rocket.Chat, Microsoft Teams,
DingTalk ou Gotify.

- É enviado um e-mail quando uma tarefa falha, aos membros do projeto que podem ver a tarefa e
  que têm os alertas ativados na sua conta.
- É enviada uma mensagem para os serviços de mensagens quando uma tarefa é concluída com êxito,
  falha ou aguarda confirmação.

Um modelo de tarefa pode impedir que as suas tarefas bem-sucedidas ou falhadas enviem alertas.

## Test Alerts {#settings-testAlerts}

Envia um alerta de teste através de todos os canais ativados no servidor, para que possa
verificar se os alertas deste projeto chegam.

O botão só funciona quando **Permitir alertas para este projeto** está assinalado e guardado.
Assinale-o e prima **Salvar** primeiro.

## Backup do Projeto {#settings-exportProject}

Transfere um ficheiro JSON com a configuração do projeto: modelos de tarefa, inventários,
repositórios, grupos de variáveis, chaves, agendas, integrações e tudo o resto que nele estiver
configurado. O ficheiro pode voltar a ser transformado num projeto com **Restaurar Projeto** no
menu de projetos, no topo da barra lateral.

O ficheiro deixa de fora os segredos das chaves e os segredos do separador **Secrets** dos grupos de
variáveis: têm de ser introduzidos novamente após um restauro. Tudo o resto é copiado tal como está,
também uma palavra-passe escrita diretamente noutro campo: o endereço de um repositório que contenha
uma palavra-passe, o texto de um inventário estático, os valores das variáveis comuns e das
variáveis de ambiente. Verifique se o ficheiro contém tais valores antes de o partilhar. O histórico
de tarefas não faz parte do backup.

## Limpar cache {#settings-clearCache}

Elimina os ficheiros que o Semaphore guarda no servidor para este projeto, como as suas cópias dos
repositórios. Estes voltam a ser transferidos quando a tarefa seguinte for iniciada.

Utilize-o quando uma tarefa falhar devido a uma cópia desatualizada ou danificada de um
repositório. A ação não pode ser anulada e não altera as configurações do projeto.

## Excluir Projeto {#settings-deleteProject}

Elimina o projeto com tudo o que contém: modelos de tarefa, o histórico de tarefas, inventários,
repositórios, grupos de variáveis, chaves e agendas. O Semaphore pede primeiro uma confirmação.

Um projeto eliminado não pode ser recuperado. Transfira um backup antes de eliminar um projeto de
que possa voltar a precisar.
