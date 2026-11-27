---
form: src/components/ProjectForm.vue
---

# Novo Projeto

Um projeto reúne os modelos de tarefa, inventários, repositórios, grupos de variáveis e chaves de
um trabalho, e as pessoas que trabalham nele. Você se torna o proprietário do projeto que cria.
Somente um administrador pode criar um projeto, a menos que o administrador do servidor permita
isso a todos os usuários. Todas as configurações abaixo podem ser alteradas depois na página
Configurações do projeto.

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

## Demo {#newProject-demo}

Preenche o novo projeto com exemplos para experimentar o Semaphore: o repositório público
`https://github.com/semaphoreui/semaphore-demo.git`, inventários, uma chave de senha do vault e
modelos de tarefa para Ansible, OpenTofu, Terragrunt, Bash e PowerShell, organizados nas
visualizações **Build**, **Deploy** e **Tools**.

Deixe a opção desativada para um projeto sem exemplos: ele recebe apenas a chave **None** e a
visualização **All**. Os exemplos são itens comuns do projeto: edite-os
ou exclua-os como quiser.
