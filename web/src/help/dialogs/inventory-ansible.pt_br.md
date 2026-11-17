---
form: src/components/InventoryForm.vue
---

# Editar Ansible Inventory

Um inventário informa ao Ansible em quais hosts executar um playbook e guarda as credenciais que
o Semaphore usa para se conectar a eles. Um modelo de tarefa faz referência a um inventário, e
todas as tarefas desse modelo são executadas nele. O formato de um inventário é descrito no
[guia de inventário do Ansible](https://docs.ansible.com/ansible/latest/inventory_guide/intro_inventory.html).

## Nome {#inventoryForm-name}

O nome com que o inventário aparece na página Inventário e no seletor de inventário de um modelo
de tarefa. Escolha um nome que diferencie os hosts, por exemplo `Production web servers`.

## Tag do Executor {#inventoryForm-runnerTag}

Exibido quando há executores de projeto disponíveis.

As tarefas que usam este inventário são executadas em um executor que tenha esta tag. Uma tag de
executor definida no modelo de tarefa tem precedência sobre a tag do inventário. Deixe o campo
vazio para que qualquer executor possa assumir as tarefas.

## Credenciais do Usuário {#inventoryForm-sshKeyId}

A chave do Armazenamento de Chaves com a qual o Ansible faz login nos hosts. O campo é
obrigatório e aceita três tipos de chave:

- **Chave SSH**: o Ansible se conecta com a chave privada e, quando a chave tem um, com o login
  dela.
- **Login com senha**: o Ansible se conecta com o login e a senha.
- **Nenhum**: o Semaphore não passa nenhuma credencial. Use para hosts que não precisam de
  credenciais, como `localhost`, ou quando o próprio inventário define como se conectar.

## Credenciais Sudo {#inventoryForm-becomeKeyId}

Opcional. Uma chave do tipo **Login com senha** do Armazenamento de Chaves, usada quando um play
eleva privilégios com `become`: o login dela é o usuário que o Ansible assume, e a senha dela é a
senha com que o Ansible responde ao prompt.

Deixe o campo vazio quando os plays não elevam privilégios, ou quando o usuário acima não precisa
de senha para isso.

## Tipo {#inventoryForm-type}

De onde vem o inventário:

- **Static**: o inventário é digitado nesta caixa de diálogo no formato INI.
- **Static YAML**: o inventário é digitado nesta caixa de diálogo no formato YAML. Use-o também
  para a configuração de um plugin de inventário.
- **File**: o inventário é um arquivo em um repositório Git.
- **Managed**: os hosts e suas variáveis são guardados como registros. Abra o inventário pela lista
  para editá-los.

Os campos abaixo do tipo mudam de acordo com ele.

## Caminho para o arquivo de Inventário {#inventoryForm-inventoryPath}

Exibido quando o tipo é **File**.

O caminho para o arquivo de inventário a partir da raiz do seu repositório, por exemplo
`inventory/production.yml`. O repositório é o selecionado no campo Repositório ou, quando esse
campo está vazio, o repositório do modelo de tarefa que usa o inventário.

## Repositório {#inventoryForm-repositoryId}

Exibido quando o tipo é **File**. Opcional.

O repositório que contém o arquivo de inventário. O Semaphore atualiza sua cópia do repositório
antes de cada tarefa. Deixe o campo vazio quando o arquivo estiver no mesmo repositório que o
playbook: o caminho é então procurado no repositório do modelo de tarefa.

## Inventário {#inventoryForm-inventoryCode}

Exibido quando o tipo é **Static** ou **Static YAML**.

O texto do inventário. O Semaphore o armazena e o grava em um arquivo temporário para cada
tarefa. Um inventário estático no formato INI:

```ini
[web]
10.0.0.11
10.0.0.12

[db]
10.0.0.21 ansible_port=2222
```

Os mesmos hosts como **Static YAML**:

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
