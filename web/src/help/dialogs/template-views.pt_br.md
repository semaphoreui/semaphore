---
form: src/components/EditViewsForm.vue
---

# Editar Visualizações

As visualizações são as abas acima da lista de modelos de tarefa. Uma visualização mostra os
modelos que você colocou nela, ou todos os modelos do projeto. A caixa de diálogo salva cada
alteração imediatamente: não há botão Salvar para a caixa de diálogo inteira. Arraste uma
visualização pela alça para alterar a ordem das abas.

## Título {#viewsForm-title}

O nome da aba. Digite-o e confirme-o com a marca de verificação ao lado do campo; o X restaura o
título salvo. Uma nova visualização é salva quando tem um título e você o confirma ou altera uma
das configurações dela. Fechar a caixa de diálogo não salva um título que você não confirmou.

A engrenagem abre as configurações da visualização abaixo do título dela.

## Excluir {#viewsForm-viewDelete}

O ícone de lixeira exclui a visualização imediatamente, sem perguntar. Os modelos da visualização
não são excluídos: eles permanecem no projeto e não estão mais em nenhuma visualização.

## Type {#viewsForm-type}

Quais modelos a aba mostra:

- **Custom**: os modelos cujo campo **Visualizar**, na caixa de diálogo do modelo, indica esta
  visualização. Este é o padrão.
- **All**: todos os modelos do projeto, seja qual for a visualização deles.

Um membro vê apenas os modelos que ele pode ver, em qualquer visualização.

## Sort by {#viewsForm-sortColumn}

A coluna pela qual os modelos da aba são ordenados; **Name** é a única. Uma aba sem ela também é
ordenada pelo nome, então o campo importa junto com **Reverse**.

## Reverse {#viewsForm-sortReverse}

Ordena a aba em ordem decrescente. Funciona apenas quando **Sort by** está definido.

## Hidden {#viewsForm-hidden}

Oculta a aba dos membros que não podem gerenciar os recursos do projeto. Quem pode continua a
vê-la, riscada. Ocultar uma aba não oculta os modelos dela: uma visualização do tipo **All**
continua a listá-los.

## Show task counts {#viewsForm-showTemplateStats}

Mostra na aba quantos dos modelos dela estão em cada estado, em um chip de segmentos coloridos:
verde para uma última tarefa bem-sucedida, vermelho para uma que falhou, cinza para uma interrompida
ou para um modelo que nunca foi executado, laranja para uma tarefa que aguarda confirmação, azul
para uma na fila ou em execução. Fica desativado por padrão, e as contagens são calculadas apenas
para as visualizações que o têm ativado.

## Adicionar visualização {#viewsForm-add}

Adiciona uma visualização ao final da lista, com um título vazio para digitar.
