---
form: src/components/InventoryForm.vue
---

# Editar Ansible Inventory

Um inventário indica ao Ansible os hosts em que deve executar um playbook e guarda as credenciais
que o Semaphore utiliza para se ligar a eles. Um modelo de tarefa faz referência a um inventário,
e todas as tarefas desse modelo são executadas sobre ele. O formato de um inventário está descrito
no [guia de inventário do Ansible](https://docs.ansible.com/ansible/latest/inventory_guide/intro_inventory.html).

## Nome {#inventoryForm-name}

O nome com que o inventário aparece na página Inventário e no seletor de inventário de um modelo
de tarefa. Escolha um nome que distinga os hosts, por exemplo `Production web servers`.

## Tag do Executor {#inventoryForm-runnerTag}

Apresentado quando existem executores de projeto disponíveis.

As tarefas que utilizam este inventário são executadas num executor que tenha esta tag. Uma tag
de executor definida no modelo de tarefa tem prioridade sobre a tag do inventário. Deixe o campo
vazio para que qualquer executor possa assumir as tarefas.

## Credenciais do Usuário {#inventoryForm-sshKeyId}

A chave do Armazenamento de Chaves com que o Ansible inicia sessão nos hosts. O campo é
obrigatório e aceita três tipos de chave:

- **Chave SSH**: o Ansible liga-se com a chave privada e, quando a chave o tem, com o respetivo
  nome de utilizador.
- **Login com senha**: o Ansible liga-se com o nome de utilizador e a palavra-passe.
- **Nenhum**: o Semaphore não transmite quaisquer credenciais. Utilize-o para hosts que não
  precisam delas, como `localhost`, ou quando o próprio inventário define como estabelecer a
  ligação.

## Credenciais Sudo {#inventoryForm-becomeKeyId}

Opcional. Uma chave do tipo **Login com senha** do Armazenamento de Chaves, utilizada quando um
play eleva privilégios com `become`: o seu nome de utilizador é o utilizador em que o Ansible se
torna, e a sua palavra-passe é a que o Ansible fornece quando lhe é pedida.

Deixe o campo vazio quando os plays não elevam privilégios, ou quando o utilizador acima não
precisa de palavra-passe para o fazer.

## Tipo {#inventoryForm-type}

De onde provém o inventário:

- **Static**: o inventário é escrito nesta caixa de diálogo no formato INI.
- **Static YAML**: o inventário é escrito nesta caixa de diálogo no formato YAML. Utilize-o também
  para a configuração de um plugin de inventário.
- **File**: o inventário é um ficheiro num repositório Git.
- **Managed**: os hosts e as suas variáveis são guardados como registos. Abra o inventário a partir
  da lista para os editar.

Os campos abaixo do tipo mudam consoante o tipo escolhido.

## Caminho para o arquivo de Inventário {#inventoryForm-inventoryPath}

Apresentado quando o tipo é **File**.

O caminho para o ficheiro de inventário a partir da raiz do respetivo repositório, por exemplo
`inventory/production.yml`. O repositório é o que está selecionado no campo Repositório ou,
quando esse campo está vazio, o repositório do modelo de tarefa que utiliza o inventário.

## Repositório {#inventoryForm-repositoryId}

Apresentado quando o tipo é **File**. Opcional.

O repositório que contém o ficheiro de inventário. O Semaphore atualiza a sua cópia do repositório
antes de cada tarefa. Deixe o campo vazio quando o ficheiro se encontra no mesmo repositório que o
playbook: nesse caso, o caminho é procurado no repositório do modelo de tarefa.

## Inventário {#inventoryForm-inventoryCode}

Apresentado quando o tipo é **Static** ou **Static YAML**.

O texto do inventário. O Semaphore guarda-o e escreve-o num ficheiro temporário para cada tarefa.
Um inventário estático no formato INI:

```ini
[web]
10.0.0.11
10.0.0.12

[db]
10.0.0.21 ansible_port=2222
```

Os mesmos hosts em **Static YAML**:

```yaml
web:
  hosts:
    10.0.0.11:
    10.0.0.12:
db:
  hosts:
    10.0.0.21:
      ansible_port: 2222
```
