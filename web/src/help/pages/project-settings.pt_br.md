---
form: src/views/project/Settings.vue, src/components/ProjectForm.vue
skip: newProject-demo
---

# Configurações do Projeto

As configurações do projeto como um todo: o nome, quantas tarefas dele podem ser executadas ao
mesmo tempo e se ele envia alertas. A Zona de Perigo abaixo do formulário reúne as ações que
afetam o projeto inteiro: o backup, o cache e a exclusão. As alterações do formulário entram em
vigor depois de **Salvar**.

## Nome do Projeto {#newProject-name}

O nome com que o projeto aparece no menu de projetos, no topo da barra lateral. O campo é
obrigatório.

## Número máximo de tarefas paralelas {#newProject-maxParallelTasks}

Quantas tarefas deste projeto podem ser executadas ao mesmo tempo. Uma tarefa iniciada acima do
limite aguarda na fila até que uma das tarefas em execução termine.

Deixe o campo vazio ou defina-o como `0` para não ter um limite próprio. O limite do servidor
inteiro, definido pelo administrador, se aplica em ambos os casos.

## ID do Chat do Telegram {#newProject-tg}

Opcional. O chat do Telegram que recebe os alertas deste projeto. Quando o campo está vazio, os
alertas vão para o chat definido para o servidor inteiro.

O campo só tem efeito quando o administrador ativou os alertas do Telegram no servidor e os
alertas estão permitidos para este projeto.

## Permitir alertas para este projeto {#newProject-alert}

Ativa os alertas para as tarefas deste projeto. Para onde os alertas vão é definido para o
servidor inteiro pelo administrador: e-mail, Telegram, Slack, Rocket.Chat, Microsoft Teams,
DingTalk ou Gotify.

- Um e-mail é enviado quando uma tarefa falha, para os membros do projeto que podem ver a tarefa
  e que ativaram os alertas na própria conta.
- Uma mensagem para os mensageiros é enviada quando uma tarefa é concluída com sucesso, falha ou
  aguarda uma confirmação.

Um modelo de tarefa pode impedir que suas tarefas bem-sucedidas ou com falha enviem alertas.

## Test Alerts {#settings-testAlerts}

Envia um alerta de teste por todos os canais que o servidor ativou, para que você possa verificar
se os alertas deste projeto estão chegando.

O botão só funciona quando **Permitir alertas para este projeto** está marcado e salvo. Marque a
opção e clique em **Salvar** primeiro.

## Backup do Projeto {#settings-exportProject}

Baixa um arquivo JSON com a configuração do projeto: modelos de tarefa, inventários,
repositórios, grupos de variáveis, chaves, agendas, integrações e tudo o mais que estiver
configurado nele. O arquivo pode ser transformado de volta em um projeto com **Restaurar Projeto**
no menu de projetos, no topo da barra lateral.

O arquivo deixa de fora os segredos das chaves e os segredos da aba **Secrets** dos grupos de
variáveis: eles precisam ser informados novamente após uma restauração. Todo o resto é copiado como
está, inclusive uma senha digitada diretamente em outro campo: o endereço de um repositório com uma
senha nele, o texto de um inventário estático, os valores das variáveis comuns e das variáveis de
ambiente. Procure esses valores no arquivo antes de compartilhá-lo. O histórico de tarefas não faz
parte do backup.

## Limpar cache {#settings-clearCache}

Exclui os arquivos que o Semaphore mantém no servidor para este projeto, como as cópias dos
repositórios. Eles são baixados novamente quando a próxima tarefa é iniciada.

Use quando uma tarefa falhar por causa de uma cópia desatualizada ou corrompida de um
repositório. A ação não pode ser desfeita e não altera as configurações do projeto.

## Excluir Projeto {#settings-deleteProject}

Exclui o projeto com tudo o que há nele: modelos de tarefa, o histórico de tarefas, inventários,
repositórios, grupos de variáveis, chaves e agendas. O Semaphore pede uma confirmação antes.

Um projeto excluído não pode ser recuperado. Baixe um backup antes de excluir um projeto de que
você possa precisar novamente.
