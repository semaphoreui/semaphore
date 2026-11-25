---
form: src/components/RestoreProjectForm.vue
---

# Restaurar Projeto...

Cria um novo projeto a partir de um ficheiro de backup feito com **Backup do Projeto** na página
Configurações de um projeto. O novo projeto recebe os modelos de tarefa, inventários,
repositórios, grupos de variáveis, chaves, agendas, integrações e as restantes coisas
configuradas no backup, e passa a ser o respetivo dono. Apenas um administrador pode restaurar
um projeto, a menos que o administrador do servidor permita que todos os utilizadores criem
projetos.

## Nome do Projeto {#restoreProject-name}

O nome do novo projeto. O campo é obrigatório e substitui o nome escrito no ficheiro de backup.

O nome tem de ser diferente dos nomes de todos os projetos do servidor: o servidor recusa
restaurar um projeto com um nome que já esteja ocupado.

## Ficheiro de backup {#restoreProject-file}

O ficheiro JSON do backup. Escolha o ficheiro antes de premir **Restore**.

Um backup deixa de fora os segredos das chaves e os segredos do separador **Secrets** dos grupos de
variáveis: depois do restauro, o Semaphore indica quantas chaves estão vazias e precisam que os seus
segredos sejam introduzidos novamente, e os segredos dos grupos de variáveis também têm de ser
introduzidos novamente. Uma palavra-passe escrita diretamente noutro campo, como o endereço de um
repositório, volta tal como está no ficheiro. Os membros do projeto e o histórico das suas tarefas
não fazem parte do backup. Um backup escrito por uma versão mais recente do Semaphore é recusado.
