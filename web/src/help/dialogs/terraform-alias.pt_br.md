---
form: src/components/TerraformAliasForm.vue
---

# Editar Key

Um alias é um endereço do backend HTTP que o Semaphore disponibiliza para este workspace: o
Terraform ou o OpenTofu apontado para ele lê, grava, bloqueia e desbloqueia o estado do workspace
no Semaphore. O Semaphore cria o endereço quando você cria o alias; nesta caixa de diálogo você
apenas escolhe a chave que o protege. Os aliases são um recurso do Semaphore PRO.

## Auth key {#tfAlias-authKeyId}

A chave que protege o alias. A lista oferece apenas chaves do tipo **Login com senha** do
Armazenamento de Chaves do projeto; crie uma lá primeiro quando a lista estiver vazia.

Um cliente do alias envia o login e a senha desta chave. Depois de salvar, copie o endereço da
lista de aliases e coloque-o no backend `http` da sua configuração, com o login e a senha da chave:

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

O endereço responde aos métodos padrão do backend `http`, então um único endereço atende ao
estado, ao bloqueio e ao desbloqueio. Escolher outra chave para um alias existente mantém o
endereço dele: os clientes passam então a precisar do login e da senha da nova chave.
