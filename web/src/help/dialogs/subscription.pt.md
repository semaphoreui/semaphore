---
form: src/components/SubscriptionForm.vue
---

# Subscription & Billing

A chave de subscrição desbloqueia as funcionalidades do Semaphore PRO e Enterprise neste
servidor. A caixa de diálogo ativa uma chave e mostra o plano da chave ativa: o seu titular, o seu
estado, quando expira e quantos utilizadores Pro, executores e outros elementos permite e estão em
uso. É aberta por um administrador a partir do menu do utilizador da edição Pro. A caixa de
diálogo chama-se **Upgrade to Semaphore PRO** enquanto não houver uma subscrição ativa.

## Enter your PRO or EE key {#subscription-key}

A chave de subscrição, tal como a recebeu com a sua subscrição PRO ou Enterprise. Cole-a aqui, ou
carregue-a a partir de um ficheiro com o botão de carregamento, e depois prima
**Activate New Key**. O campo é obrigatório para a ativação.

O campo é só de leitura quando a chave está definida na configuração do servidor
(`subscription.key` ou `subscription.key_file`, ou as variáveis `SEMAPHORE_SUBSCRIPTION_KEY`
e `SEMAPHORE_SUBSCRIPTION_KEY_FILE`): nesse caso, a chave só pode ser alterada aí.

## Carregar ficheiro de chave {#subscription-upload}

Apresentado quando ainda não foi ativada nenhuma subscrição.

Lê a chave de um ficheiro no seu computador (`.txt`, `.key`, `.pem` ou `.lic`) para o campo
acima. A chave só é ativada quando prime **Activate New Key**.

## Ações da chave {#subscription-actionsMenu}

Apresentado quando uma subscrição está ativa ou expirou.

Um menu de ações sobre a chave do servidor:

- **Reload**: obtém novamente o estado da subscrição, por exemplo depois de ter sido renovada.
- **Upload**: lê uma nova chave de um ficheiro para o campo acima, tal como o botão de
  carregamento.
- **Reset**: remove a chave do servidor de imediato, sem pedir confirmação.

## Activate New Key {#subscription-activate}

Ativa neste servidor a chave do campo. Quando a chave é válida, o plano da subscrição é
apresentado abaixo. Quando a sua conta ainda não é um utilizador Pro e o plano tem um lugar de
utilizador Pro livre, a caixa de diálogo pergunta se deve tornar a sua conta um utilizador Pro.

Uma chave que já está ativada é assinalada com um aviso. O botão está desativado quando a chave
está definida na configuração do servidor.

## Buy Pro {#subscription-buy}

Abre a página do portal do Semaphore onde se compra uma subscrição PRO, num novo separador do
navegador.

## Get 30-day free trial {#subscription-trial}

Apresentado enquanto não houver uma subscrição ativa.

Abre a página do portal do Semaphore que emite uma chave para um período experimental gratuito de
30 dias, num novo separador do navegador. Cole a chave que receber no campo acima.
