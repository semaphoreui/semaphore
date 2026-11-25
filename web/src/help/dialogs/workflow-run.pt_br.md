---
form: src/components/WorkflowRunDialog.vue
---

# Run Workflow

Inicia uma execução do workflow depois que você responde à pesquisa dele. A caixa de diálogo só
abre para um workflow com pesquisa: um workflow sem pesquisa é iniciado imediatamente. Depois de
**Executar**, a página da nova execução se abre; quando o servidor recusa os valores, a caixa de
diálogo permanece aberta com as suas respostas.

## Variáveis {#workflowRunDialog-survey}

Um campo por variável da pesquisa do workflow, na ordem da pesquisa, com o valor padrão já
preenchido. Um campo marcado com `*` é obrigatório, um campo Integer aceita apenas dígitos, e um
campo de lista aceita um valor ou, quando a variável permite vários, qualquer quantidade deles.
**Executar** não inicia nada até que todos os campos sejam válidos.

Os valores são armazenados com a execução e exibidos na página dela em **Variables**. Uma tarefa
do workflow não os recebe por conta própria: uma variável do modelo dela recebe um valor apenas
quando o nó dela a vincula a uma variável do workflow. O nome sozinho não entrega nada.

Uma pesquisa de workflow não guarda segredos, porque a execução mantém os valores dela em texto
claro. Coloque senhas e tokens em uma chave ou em um grupo de variáveis.
