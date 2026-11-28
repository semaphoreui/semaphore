---
form: src/components/ProjectForm.vue
---

# Novo Projeto

Um projeto reúne os modelos de tarefa, inventários, repositórios, grupos de variáveis e chaves de
um trabalho, e as pessoas que nele trabalham. Torna-se proprietário do projeto que criar. Só um
administrador pode criar um projeto, a menos que o administrador do servidor o permita a todos
os utilizadores. Todas as configurações abaixo podem ser alteradas mais tarde na página
Configurações do projeto.

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

## Demo {#newProject-demo}

Preenche o novo projeto com exemplos para experimentar o Semaphore: o repositório público
`https://github.com/semaphoreui/semaphore-demo.git`, inventários, uma chave de palavra-passe do
vault e modelos de tarefa para Ansible, OpenTofu, Terragrunt, Bash e PowerShell, distribuídos
pelas visualizações **Build**, **Deploy** e **Tools**.

Deixe o interruptor desligado para um projeto sem exemplos: recebe apenas a chave **Nenhum** e a
visualização **All**. Os exemplos são elementos comuns do projeto: edite-os
ou elimine-os como quiser.
