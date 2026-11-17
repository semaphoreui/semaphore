---
form: src/components/TerraformAliasForm.vue
---

# Editar Key

Um alias é um endereço do backend HTTP que o Semaphore disponibiliza para este workspace: o
Terraform ou o OpenTofu apontado para ele lê, escreve, bloqueia e desbloqueia o estado do
workspace no Semaphore. O Semaphore cria o endereço quando cria o alias; nesta caixa de diálogo
escolhe apenas a chave que o protege. Os aliases são uma funcionalidade do Semaphore PRO.

## Auth key {#tfAlias-authKeyId}

A chave que protege o alias. A lista oferece apenas chaves do tipo **Login com senha** do
Armazenamento de Chaves do projeto; crie primeiro uma lá quando a lista estiver vazia.

Um cliente do alias envia o nome de utilizador e a palavra-passe desta chave. Depois de guardar,
copie o endereço da lista de aliases e coloque-o no backend `http` da sua configuração, com o
nome de utilizador e a palavra-passe da chave:

```hcl
terraform {
  backend "http" {
    address        = "ALIAS_URL"
    lock_address   = "ALIAS_URL"
    unlock_address = "ALIAS_URL"
    username       = "LOGIN"
    password       = "PASSWORD"
  }
}
```

O endereço responde aos métodos predefinidos do backend `http`, pelo que um só endereço serve o
estado, o bloqueio e o desbloqueio. Escolher outra chave para um alias existente mantém o seu
endereço: os clientes passam então a precisar do nome de utilizador e da palavra-passe da nova
chave.
