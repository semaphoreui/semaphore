---
form: src/views/Tokens.vue
---

# New Token

Un token de API permite que un script u otro servicio llame a la API de Semaphore en su nombre, con
los mismos derechos que su cuenta. Envíelo en el encabezado `Authorization` de una solicitud:
`Bearer` y el token. El token se muestra una sola vez, justo después de crearlo: cópielo en ese
momento.

## Token Name {#token-name}

Opcional. Un nombre que distingue el token en la lista de tokens, por ejemplo el nombre del script
que lo usa. No forma parte del token y no cambia lo que el token puede hacer.

## Expires {#token-expires}

Cuándo deja de funcionar el token. Elija un periodo contado desde ahora, **Never** para un token que
funciona hasta que se elimina, o **Custom** para elegir usted mismo el momento.

Un token que nunca caduca funciona hasta que lo elimine de la lista. Prefiera un periodo para un
token que se entrega a un script o a otro servicio.

## Expiration date {#token-expiresAt}

Se muestra cuando **Expires** es **Custom**.

La fecha y la hora en que el token deja de funcionar, en la zona horaria de su navegador. El momento
debe estar en el futuro: el servidor rechaza un token que ya estaría caducado.
