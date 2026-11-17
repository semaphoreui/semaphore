---
form: src/components/AppForm.vue
---

# Edit App

Um app é uma ferramenta com a qual o Semaphore executa tarefas: Ansible, Terraform, OpenTofu,
Terragrunt, Bash, Python, PowerShell ou um comando seu. Cada modelo de tarefa é feito para um app,
e esta caixa de diálogo define como o servidor o inicia e como o app é exibido na interface.

As configurações são armazenadas no banco de dados e entram em vigor imediatamente no servidor que
as salva; os outros nós de um cluster as leem quando são reiniciados. Um campo que você esvazia
mantém o valor antigo até o servidor ser reiniciado.

## ID {#appForm-id}

Obrigatório. A chave com que o app é armazenado; um modelo de tarefa se refere ao seu app por essa
chave. Os apps incorporados são `ansible`, `terraform`, `tofu`, `terragrunt`, `pulumi`, `bash`,
`python` e `powershell`. Qualquer outro ID adiciona um app seu: o Semaphore o executa como um
script, com o comando indicado em **Path** ou, quando **Path** está vazio, com o comando que tem o
nome do ID.

Use um ID sem pontos. Alterar o ID de um app existente não o renomeia: as configurações são salvas
como um novo app com o novo ID, e o app antigo permanece.

## Icon {#appForm-icon}

Opcional. O nome de um ícone do conjunto Material Design Icons sem o prefixo `mdi-`, por exemplo
`language-python`. Deixe vazio para manter o ícone incorporado do app;
um app seu exibe então um ponto de interrogação.

## Nome {#appForm-title}

Opcional. O nome do app na lista de aplicativos e no menu **Novo modelo**. Deixe vazio para manter
o nome incorporado, como `Ansible Playbook`; um app seu fica então sem nome, então dê um a ele.

## Path {#appForm-path}

Opcional. O comando que o Semaphore executa no lugar do padrão, seja um nome procurado no `PATH`
do servidor, seja um caminho completo, por exemplo `/usr/local/bin/terraform-1.9`. Os padrões são
`terraform`, `tofu`, `terragrunt`, `bash`, `python3` e `powershell`; um app seu executa o comando
que tem o nome do seu ID.

O Ansible ignora este campo: o Semaphore sempre executa `ansible-playbook` e `ansible-galaxy`.

## Priority {#appForm-priority}

A posição do app na lista de aplicativos e no menu **Novo modelo**: um app com um número maior vem
primeiro. Os apps incorporados vão de 1000 para o Ansible até 500 para o Python. Um app incorporado
com prioridade `0` ou menor recebe de volta a sua prioridade incorporada quando o servidor é
iniciado.

## Argumentos {#appForm-args}

Opcional. Argumentos colocados na linha de comando logo depois do comando, um por chip. Clique em
**+** para adicionar um, clique em um chip para alterá-lo, arraste os chips para mudar a ordem
deles.

- Terraform, OpenTofu e Terragrunt: os argumentos vêm antes dos argumentos que o Semaphore
  adiciona a todos os comandos, como `init` e `apply`.
- Apps de script: os argumentos vêm antes do script do modelo e substituem os padrões: o
  PowerShell é executado com `-File` por padrão, então coloque `-File` por último na sua lista.
- O Ansible ignora este campo.

## Active {#appForm-active}

Um app ativo é oferecido no menu **Novo modelo**. Desmarque a caixa para ocultar o app desse menu:
os modelos já criados com ele mantêm o seu app e continuam sendo executados.

A caixa vem desmarcada para um novo app: marque-a para tornar o app disponível.
