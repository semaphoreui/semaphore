---
form: src/components/InventoryHostForm.vue
---

# Editar host

Um host de um inventário gerenciado. Uma tarefa que usa o inventário é executada nos hosts que estão
ativados na lista de hosts.

## Nome {#inventoryHostForm-name}

O nome pelo qual o Ansible conhece o host e o endereço ao qual ele se conecta, a menos que as
variáveis definam `ansible_host`. O campo é obrigatório. Um nome é único em seu inventário e
diferencia maiúsculas de minúsculas: `web1` e `Web1` são dois hosts. Não pode conter espaços,
vírgulas, colchetes nem dois-pontos, e tem no máximo 250 caracteres.

## Descrição {#inventoryHostForm-description}

Opcional. Uma nota para as pessoas, exibida na lista de hosts. O Ansible não a recebe.

## Variáveis {#inventoryHostForm-variables}

Opcional. As variáveis deste host, escritas como um mapeamento YAML ou JSON:

```yaml
ansible_host: 10.0.0.11
ansible_port: 2222
```

Uma variável do host tem precedência sobre uma variável do inventário com o mesmo nome. Âncoras,
aliases e chaves de mesclagem de YAML não são aceitos.
