---
form: src/components/AppForm.vue
---

# Edit App

Uma aplicação é uma ferramenta com que o Semaphore executa tarefas: Ansible, Terraform, OpenTofu,
Terragrunt, Bash, Python, PowerShell ou um comando à sua escolha. Cada modelo de tarefa é feito
para uma aplicação, e esta caixa de diálogo define como o servidor a inicia e como a aplicação é
apresentada na interface.

As configurações são guardadas na base de dados e produzem efeito de imediato no servidor que as
guarda; os outros nós de um cluster leem-nas quando são reiniciados. Um campo que esvazie mantém
o valor anterior até o servidor ser reiniciado.

## ID {#appForm-id}

Obrigatório. A chave sob a qual a aplicação é guardada; um modelo de tarefa faz referência à sua
aplicação por esta chave. As aplicações incorporadas são `ansible`, `terraform`, `tofu`,
`terragrunt`, `pulumi`, `bash`, `python` e `powershell`. Qualquer outro ID acrescenta uma
aplicação própria: o Semaphore executa-a como um script, com o comando indicado em **Path** ou,
quando **Path** está vazio, com o comando com o mesmo nome que o ID.

Utilize um ID sem pontos. Alterar o ID de uma aplicação existente não lhe muda o nome: as
configurações são guardadas como uma nova aplicação com o novo ID, e a aplicação antiga mantém-se.

## Icon {#appForm-icon}

Opcional. O nome de um ícone do conjunto Material Design Icons sem o prefixo `mdi-`, por exemplo
`language-python`. Deixe o campo vazio para manter o ícone incorporado da aplicação;
uma aplicação própria apresenta então um ponto de interrogação.

## Nome {#appForm-title}

Opcional. O nome da aplicação na lista de aplicações e no menu **Novo modelo**. Deixe o campo
vazio para manter o nome incorporado, como `Ansible Playbook`; uma aplicação própria fica então
sem nome, por isso dê-lhe um.

## Path {#appForm-path}

Opcional. O comando que o Semaphore executa em vez do predefinido: um nome procurado no `PATH`
do servidor ou um caminho completo, por exemplo `/usr/local/bin/terraform-1.9`. Os predefinidos
são `terraform`, `tofu`, `terragrunt`, `bash`, `python3` e `powershell`; uma aplicação própria
executa o comando com o mesmo nome que o seu ID.

O Ansible ignora este campo: o Semaphore executa sempre `ansible-playbook` e `ansible-galaxy`.

## Priority {#appForm-priority}

A posição da aplicação na lista de aplicações e no menu **Novo modelo**: uma aplicação com um
número mais alto aparece primeiro. As aplicações incorporadas vão de 1000 para o Ansible até 500
para o Python. Uma aplicação incorporada com uma prioridade de `0` ou inferior recupera a sua
prioridade incorporada quando o servidor é iniciado.

## Args {#appForm-args}

Opcional. Argumentos colocados na linha de comandos logo a seguir ao comando, um por chip. Clique
em **+** para acrescentar um, clique num chip para o alterar e arraste os chips para mudar a sua
ordem.

- Terraform, OpenTofu e Terragrunt: os argumentos vêm antes dos argumentos que o Semaphore
  acrescenta a cada comando, como `init` e `apply`.
- Aplicações de script: os argumentos vêm antes do script do modelo e substituem os
  predefinidos: o PowerShell é executado com `-File` por predefinição, por isso coloque `-File`
  em último lugar na sua lista.
- O Ansible ignora este campo.

## Active {#appForm-active}

Uma aplicação ativa é proposta no menu **Novo modelo**. Desmarque a caixa para ocultar a
aplicação desse menu: os modelos já criados com ela mantêm a sua aplicação e continuam a ser
executados.

A caixa está desmarcada numa aplicação nova: assinale-a para disponibilizar a aplicação.
