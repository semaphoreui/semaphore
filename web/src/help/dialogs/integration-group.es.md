---
form: src/components/IntegrationGroupForm.vue, src/components/IntegrationAuthForm.vue
---

# Editar integración

Una integración es un conjunto de ejecuciones que se inician desde un endpoint. Cuando una solicitud
a una URL de ese endpoint supera la autorización, se inicia cada ejecución de la integración cuyos
matchers aceptan la solicitud.

## Nombre {#integrationGroupForm-name}

El nombre con el que la integración aparece en la página «Integraciones». El campo es obligatorio.

## Endpoint {#integrationGroupForm-endpointId}

El endpoint cuyas URL inician las ejecuciones de esta integración. Varias integraciones pueden usar
un mismo endpoint. El campo es obligatorio.

## Sobrescribir autorización {#integrationGroupForm-authOverride}

Déjelo desactivado para comprobar las solicitudes como lo hace el endpoint. Actívelo para comprobar
las solicitudes de esta integración con un método y un secreto propios, definidos en los campos que
aparecen debajo; la configuración del endpoint deja entonces de aplicarse a esta integración.

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
