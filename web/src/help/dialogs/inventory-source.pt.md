---
form: src/components/InventorySourceForm.vue
---

# Editar fonte

Uma fonte importa os hosts e os grupos de um inventário gerido a partir de um ficheiro de um repositório Git. O ficheiro é lido sempre que se prime Sync, e os hosts e grupos que descreve ficam alinhados com ele. O servidor precisa de Ansible e de acesso ao repositório.

## Nome {#inventorySourceForm-name}

O nome da fonte, mostrado no seu cartão e junto dos hosts que ela traz. O campo é obrigatório.

## Repositório {#inventorySourceForm-repository}

O repositório do projeto de onde o ficheiro é lido. O campo é obrigatório. É usado o seu branch.

## Caminho {#inventorySourceForm-path}

O caminho de um ficheiro de inventário no repositório, por exemplo `inventory/prod.yml`, ou de um diretório de ficheiros de inventário. O campo é obrigatório.

Os ficheiros são lidos como YAML ou INI estáticos, scripts e ficheiros de configuração de plugins são recusados. Num diretório, cada ficheiro que o Ansible não ignora tem de ser um inventário válido, caso contrário a sincronização falha e nada muda. O Ansible ignora algumas extensões num diretório, por exemplo `.cfg`, `.md` e `.txt`. Um repositório com uma ligação que aponta para fora dele é recusado.
