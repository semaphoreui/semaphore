---
form: src/components/SubscriptionForm.vue
---

# Subscription & Billing

A chave de assinatura libera os recursos do Semaphore PRO e Enterprise neste servidor. A caixa de
diálogo ativa uma chave e mostra o plano da chave ativa: o titular, o status, quando expira e
quantos usuários Pro, executores e outros itens ela permite e estão em uso. Ela é aberta por um
administrador a partir do menu do usuário da edição Pro. A caixa de diálogo se chama
**Upgrade to Semaphore PRO** até que uma assinatura esteja ativa.

## Enter your PRO or EE key {#subscription-key}

A chave de assinatura, como você a recebeu com a sua assinatura PRO ou Enterprise. Cole-a aqui, ou
carregue-a de um arquivo com o botão de upload, e clique em **Activate New Key**. O campo é
obrigatório para a ativação.

O campo fica somente leitura quando a chave está definida na configuração do servidor
(`subscription.key` ou `subscription.key_file`, ou as variáveis `SEMAPHORE_SUBSCRIPTION_KEY` e
`SEMAPHORE_SUBSCRIPTION_KEY_FILE`): a chave então só pode ser alterada lá.

## Enviar arquivo de chave {#subscription-upload}

Exibido quando nenhuma assinatura foi ativada ainda.

Lê a chave de um arquivo no seu computador (`.txt`, `.key`, `.pem` ou `.lic`) para o campo acima.
A chave não é ativada até você clicar em **Activate New Key**.

## Ações da chave {#subscription-actionsMenu}

Exibido quando uma assinatura está ativa ou expirou.

Um menu de ações sobre a chave do servidor:

- **Reload**: busca novamente o estado da assinatura, por exemplo depois que ela foi renovada.
- **Upload**: lê uma nova chave de um arquivo para o campo acima, como faz o botão de upload.
- **Reset**: remove a chave do servidor imediatamente, sem confirmação.

## Activate New Key {#subscription-activate}

Ativa neste servidor a chave do campo. Quando a chave é válida, o plano da assinatura é exibido
abaixo. Quando a sua conta ainda não é um usuário Pro e o plano tem uma vaga livre de usuário Pro,
a caixa de diálogo pergunta se a sua conta deve se tornar um usuário Pro.

Uma chave que já está ativada é informada com um aviso. O botão fica desativado quando a chave está
definida na configuração do servidor.

## Buy Pro {#subscription-buy}

Abre, em uma nova aba do navegador, a página do portal do Semaphore em que uma assinatura PRO é
comprada.

## Get 30-day free trial {#subscription-trial}

Exibido enquanto nenhuma assinatura está ativa.

Abre, em uma nova aba do navegador, a página do portal do Semaphore que emite uma chave para um
teste gratuito de 30 dias. Cole a chave que você receber no campo acima.
