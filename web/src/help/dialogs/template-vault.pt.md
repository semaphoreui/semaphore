---
form: src/components/TemplateVaults.vue
---

# Cofre

Um cofre fornece ao Ansible a palavra-passe dos ficheiros e das variáveis cifrados com o Ansible
Vault no repositório do playbook. Cada cofre do modelo torna-se uma opção `--vault-id` de
`ansible-playbook`. O cofre é guardado com o modelo, quando guarda a caixa de diálogo do modelo.

## ID do Cofre {#templateVaultDialog-name}

Opcional. O vault ID a que a palavra-passe se destina, a etiqueta com que o conteúdo foi cifrado
(`ansible-vault encrypt --vault-id prod@prompt`). Deixe-o vazio para o vault ID `default`.

O ID é guardado em minúsculas. Cada cofre do modelo precisa do seu próprio ID, e apenas um pode
estar vazio.

## Tipo {#templateVaultDialog-type}

De onde provém a palavra-passe:

- **Senha**: uma chave do Armazenamento de Chaves. O Semaphore escreve a sua palavra-passe quando
  o Ansible a pede, pelo que nunca é gravada num ficheiro.
- **Script do Cliente**: um script cliente de palavra-passe do vault que imprime a palavra-passe;
  o Ansible executa-o.

## Senha do Cofre {#templateVaultDialog-vaultKeyId}

Apresentado quando o tipo é **Senha**.

A chave que contém a palavra-passe do cofre. Só são listadas as chaves do tipo **Login com senha**;
o Semaphore utiliza a sua palavra-passe e ignora o login. O campo é obrigatório.

## Caminho do Script {#templateVaultDialog-script}

Apresentado quando o tipo é **Script do Cliente**.

O caminho para o script, por exemplo `scripts/vault-pass-client.py`. Um caminho relativo é tomado a
partir do diretório em que o Ansible é executado: a raiz do repositório, ou o diretório de
trabalho do modelo quando este tem um. O campo é obrigatório.

O nome do ficheiro tem de terminar em `-client`, com ou sem extensão: só assim o Ansible executa o
ficheiro como script cliente e lhe transmite o vault ID com `--vault-id`. O script é executado na
máquina que executa a tarefa, como o utilizador com que o Semaphore é executado.
