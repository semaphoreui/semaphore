---
form: src/components/EditViewsForm.vue
---

# Editar Visualizações

As visualizações são os separadores acima da lista de modelos de tarefa. Uma visualização mostra
os modelos que lhe atribuir, ou todos os modelos do projeto. A caixa de diálogo guarda cada
alteração de imediato: não há botão para guardar a caixa de diálogo inteira. Arraste uma
visualização pela pega para mudar a ordem dos separadores.

## Título {#viewsForm-title}

O nome do separador. Escreva-o e confirme-o com o visto ao lado do campo; a cruz repõe o título
guardado. Uma nova visualização é guardada quando tem um título e o confirma ou altera uma das
suas configurações. Fechar a caixa de diálogo não guarda um título que não tenha confirmado.

A roda dentada abre as configurações da visualização por baixo do seu título.

## Eliminar {#viewsForm-viewDelete}

O ícone do caixote do lixo elimina a visualização de imediato, sem pedir confirmação. Os modelos
da visualização não são eliminados: ficam no projeto e deixam de estar em qualquer visualização.

## Type {#viewsForm-type}

Que modelos o separador mostra:

- **Custom**: os modelos cujo campo **Visualizar**, na caixa de diálogo do modelo, indica esta
  visualização. É a predefinição.
- **All**: todos os modelos do projeto, seja qual for a sua visualização.

Um membro só vê os modelos que pode ver, em qualquer visualização.

## Sort by {#viewsForm-sortColumn}

A coluna pela qual os modelos do separador são ordenados; **Name** é a única. Um separador sem
ela também é ordenado por nome, por isso o campo só importa em conjunto com **Reverse**.

## Reverse {#viewsForm-sortReverse}

Ordena o separador por ordem decrescente. Só funciona quando **Sort by** está definido.

## Hidden {#viewsForm-hidden}

Oculta o separador aos membros que não podem gerir os recursos do projeto. Quem pode continua a
vê-lo, riscado. Ocultar um separador não oculta os seus modelos: uma visualização do tipo **All**
continua a listá-los.

## Show task counts {#viewsForm-showTemplateStats}

Mostra no separador quantos dos seus modelos estão em cada estado, num chip de segmentos coloridos:
verde para uma última tarefa bem-sucedida, vermelho para uma falhada, cinzento para uma parada ou
para um modelo que nunca foi executado, laranja para uma tarefa que aguarda confirmação, azul para
uma em fila ou em execução. Está desativado por omissão, e as contagens só são calculadas para as
visualizações que o têm ativado.

## Adicionar visualização {#viewsForm-add}

Acrescenta uma visualização no fim da lista, com um título vazio para preencher.
