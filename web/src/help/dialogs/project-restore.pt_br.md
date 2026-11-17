---
form: src/components/RestoreProjectForm.vue
---

# Restaurar Projeto...

Cria um novo projeto a partir de um arquivo de backup feito com **Backup do Projeto** na página
Configurações de um projeto. O novo projeto recebe os modelos de tarefa, inventários,
repositórios, grupos de variáveis, chaves, agendas, integrações e tudo o mais que estiver
configurado no backup, e você se torna o proprietário dele. Somente um administrador pode
restaurar um projeto, a menos que o administrador do servidor permita que todos os usuários criem
projetos.

## Nome do Projeto {#restoreProject-name}

O nome do novo projeto. O campo é obrigatório e substitui o nome gravado no arquivo de backup.

O nome precisa ser diferente dos nomes de todos os projetos do servidor: o servidor se recusa a
restaurar um projeto com um nome que já está em uso.

## Backup file {#restoreProject-file}

O arquivo JSON do backup. Escolha o arquivo antes de clicar em **Restore**.

Um backup deixa de fora os segredos das chaves e os segredos da aba **Secrets** dos grupos de
variáveis: após a restauração, o Semaphore informa quantas chaves estão vazias e precisam ter os
segredos informados novamente, e os segredos dos grupos de variáveis também precisam ser informados
novamente. Uma senha digitada diretamente em outro campo, como o endereço de um repositório, volta
como está no arquivo. Os membros do projeto e o histórico das tarefas dele não fazem parte do
backup. Um backup gravado por uma versão mais recente do Semaphore é recusado.
