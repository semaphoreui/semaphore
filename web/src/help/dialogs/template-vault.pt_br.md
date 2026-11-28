---
form: src/components/TemplateVaults.vue
---

# Cofre

Um cofre fornece ao Ansible a senha dos arquivos e das variáveis criptografados com o Ansible Vault
no repositório do playbook. Cada cofre do modelo se torna uma opção `--vault-id` do
`ansible-playbook`. O cofre é salvo com o modelo, quando você salva a caixa de diálogo do modelo.

## ID do Cofre {#templateVaultDialog-name}

Opcional. O vault ID ao qual a senha se refere, o rótulo com que o conteúdo foi criptografado
(`ansible-vault encrypt --vault-id prod@prompt`). Deixe-o vazio para o vault ID `default`.

O ID é salvo em letras minúsculas. Cada cofre do modelo precisa do seu próprio ID, e apenas um pode
ficar vazio.

## Tipo {#templateVaultDialog-type}

De onde vem a senha:

- **Senha**: uma chave do Armazenamento de Chaves. O Semaphore digita a senha dela quando o Ansible
  a solicita, de modo que ela nunca é gravada em um arquivo.
- **Script do Cliente**: um script cliente de senha do vault que imprime a senha; o Ansible o
  executa.

## Senha do Cofre {#templateVaultDialog-vaultKeyId}

Exibido quando o tipo é **Senha**.

A chave que guarda a senha do cofre. Apenas chaves do tipo **Login com senha** são listadas; o
Semaphore usa a senha delas e ignora o login. O campo é obrigatório.

## Caminho do Script {#templateVaultDialog-script}

Exibido quando o tipo é **Script do Cliente**.

O caminho para o script, por exemplo `scripts/vault-pass-client.py`. Um caminho relativo é
considerado a partir do diretório em que o Ansible é executado: a raiz do repositório, ou o
diretório de trabalho do modelo, quando ele tem um. O campo é obrigatório.

O nome do arquivo precisa terminar em `-client`, com ou sem extensão: só assim o Ansible executa o
arquivo como script cliente e passa a ele o vault ID com `--vault-id`. O script é executado na
máquina que executa a tarefa, como o usuário com que o Semaphore é executado.
