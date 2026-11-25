---
form: src/views/Tokens.vue
---

# New Token

Token API pozwala skryptowi lub innej usłudze wywoływać API Semaphore w twoim imieniu, z takimi
samymi uprawnieniami jak twoje konto. Wyślij go w nagłówku `Authorization` żądania: `Bearer` i
token. Token jest wyświetlany raz, zaraz po utworzeniu: skopiuj go wtedy.

## Token Name {#token-name}

Opcjonalne. Nazwa, która odróżnia token na liście tokenów, na przykład nazwa skryptu, który go
używa. Nie jest częścią tokenu i nie zmienia tego, co token może robić.

## Expires {#token-expires}

Kiedy token przestaje działać. Wybierz okres liczony od teraz, **Never** dla tokenu, który działa do
momentu usunięcia, lub **Custom**, aby samodzielnie wybrać moment.

Token, który nigdy nie wygasa, działa, dopóki nie usuniesz go z listy. Dla tokenu przekazywanego
skryptowi lub innej usłudze lepiej wybrać okres.

## Expiration date {#token-expiresAt}

Wyświetlane, gdy **Expires** to **Custom**.

Data i godzina, o której token przestaje działać, w strefie czasowej twojej przeglądarki. Moment
musi leżeć w przyszłości: serwer odrzuca token, który byłby już wygasły.
