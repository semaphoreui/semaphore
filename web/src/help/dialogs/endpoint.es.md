---
form: src/components/EndpointForm.vue, src/components/IntegrationAuthForm.vue
---

# Editar endpoint

Un endpoint es un punto de entrada para webhooks: las URL a las que llaman otros servicios y la
forma en que se comprueban sus solicitudes. Las integraciones se asocian a un endpoint e inician sus
ejecuciones con las solicitudes que este acepta. Las URL de un endpoint, sus alias, se añaden en la
pestaña «Endpoints».

## Nombre {#endpointForm-name}

El nombre con el que el endpoint aparece en la pestaña «Endpoints» y en el formulario de una
integración. El campo es obligatorio.

## Descripción {#endpointForm-description}

Opcional. Una nota sobre para qué sirve el endpoint; se muestra junto a su nombre.

## Método de autenticación {#integrationAuthForm-authMethod}

Cómo comprueba Semaphore que una solicitud procede del servicio que espera. Una solicitud que no
supera la comprobación no inicia nada, y el webhook le responde igual que a cualquier otra
solicitud.

- **Ninguno**: sin comprobación. Cualquiera que conozca la URL puede iniciar la tarea.
- **GitHub Webhooks**: el encabezado `X-Hub-Signature-256` debe contener la firma HMAC-SHA256 del
  cuerpo hecha con el secreto, tal como la envía GitHub.
- **Bitbucket Webhooks**: la misma comprobación sobre el encabezado `X-Hub-Signature`, tal como la
  envía Bitbucket.
- **Token**: el encabezado indicado en **Cabecera de autenticación** debe contener el propio secreto.
- **HMAC** y **HMAC (SHA-512)**: el encabezado indicado en **Cabecera de autenticación** debe contener el
  HMAC-SHA256 o el HMAC-SHA512 del cuerpo hecho con el secreto, en hexadecimal en minúsculas, sin un
  prefijo como `sha256=`.
- **BasicAuth**: la solicitud debe llevar el usuario y la contraseña de la clave de
  **Secreto de autenticación** en su encabezado `Authorization`.

## Cabecera de autenticación {#integrationAuthForm-authHeader}

Se muestra cuando **Método de autenticación** es **Token**, **HMAC** o **HMAC (SHA-512)**.

El nombre del encabezado de la solicitud que lleva el token o la firma, por ejemplo
`X-Webhook-Token`. El nombre no distingue entre mayúsculas y minúsculas. Una solicitud sin ese
encabezado se rechaza.

## Secreto de autenticación {#integrationAuthForm-authSecretId}

Se muestra cuando **Método de autenticación** no es **Ninguno**.

La clave del Almacén de Claves con la que se comprueba la solicitud. Para **BasicAuth** solo sirve
una clave del tipo **Iniciar sesión con contraseña**. Para los demás métodos también sirve una
clave del tipo **Cadena**: la comprobación usa su valor, o la contraseña de una clave
**Iniciar sesión con contraseña**.

Todos los métodos excepto **Ninguno** necesitan el campo: sin una clave, o con un secreto vacío, se
rechaza cada solicitud.
