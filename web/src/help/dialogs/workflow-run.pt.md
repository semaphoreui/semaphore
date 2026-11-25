---
form: src/components/WorkflowRunDialog.vue
---

# Executar workflow

Inicia uma execução do workflow depois de responder ao seu questionário. A caixa de diálogo só
abre para um workflow com questionário: um workflow sem questionário é iniciado de imediato.
Depois de **Executar** abre-se a página da nova execução; quando o servidor recusa os valores, a
caixa de diálogo permanece aberta com as suas respostas.

## Variáveis {#workflowRunDialog-survey}

Um campo por cada variável do questionário do workflow, pela ordem do questionário, com o seu
valor predefinido já preenchido. Um campo marcado com `*` é obrigatório, um campo Integer aceita
apenas algarismos, e um campo de lista aceita um valor ou, quando a variável permite vários,
qualquer número deles. **Executar** não inicia nada enquanto todos os campos não forem válidos.

Os valores são guardados com a execução e apresentados na respetiva página em **Variables**. Uma
tarefa do workflow não os recebe por si só: uma variável do seu modelo recebe um valor apenas
quando o seu nó a associa a uma variável do workflow. O nome, por si só, não transmite nada.

Um questionário de workflow não guarda segredos, porque a execução conserva os seus valores em
texto simples. Coloque palavras-passe e tokens numa chave ou num grupo de variáveis.
