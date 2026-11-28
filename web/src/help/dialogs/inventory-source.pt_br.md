---
form: src/components/InventorySourceForm.vue
---

# Editar fonte

Uma fonte importa os hosts e os grupos de um inventário gerenciado a partir de um arquivo de um repositório Git. O arquivo é lido sempre que Sync é pressionado, e os hosts e grupos que ele descreve ficam alinhados com ele. O servidor precisa do Ansible e de acesso ao repositório.

## Nome {#inventorySourceForm-name}

O nome da fonte, mostrado no cartão dela e ao lado dos hosts que ela traz. O campo é obrigatório.

## Repositório {#inventorySourceForm-repository}

O repositório do projeto de onde o arquivo é lido. O campo é obrigatório. O branch dele é usado.

## Caminho {#inventorySourceForm-path}

O caminho de um arquivo de inventário no repositório, por exemplo `inventory/prod.yml`, ou de um diretório de arquivos de inventário. O campo é obrigatório.

Os arquivos são lidos como YAML ou INI estáticos, scripts e arquivos de configuração de plugins são recusados. Em um diretório, cada arquivo que o Ansible não ignora precisa ser um inventário válido, senão a sincronização falha e nada muda. O Ansible ignora algumas extensões em um diretório, por exemplo `.cfg`, `.md` e `.txt`. Um repositório com um link que leva para fora dele é recusado.
