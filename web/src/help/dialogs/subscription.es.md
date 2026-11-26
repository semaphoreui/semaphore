---
form: src/components/SubscriptionForm.vue
---

# Subscription & Billing

La clave de suscripción desbloquea las funciones de Semaphore PRO y Enterprise en este servidor. El
diálogo activa una clave y muestra el plan de la clave activa: su titular, su estado, cuándo caduca
y cuántos usuarios Pro, ejecutores y otros elementos permite y están en uso. Lo abre un
administrador desde el menú de usuario de la edición Pro. El diálogo se llama
**Upgrade to Semaphore PRO** mientras no haya una suscripción activa.

## Enter your PRO or EE key {#subscription-key}

La clave de suscripción, tal como la recibió con su suscripción PRO o Enterprise. Péguela aquí, o
cárguela desde un archivo con el botón de carga, y pulse **Activate New Key**. El campo es
obligatorio para la activación.

El campo es de solo lectura cuando la clave está definida en la configuración del servidor
(`subscription.key` o `subscription.key_file`, o las variables `SEMAPHORE_SUBSCRIPTION_KEY`
y `SEMAPHORE_SUBSCRIPTION_KEY_FILE`): en ese caso la clave solo se puede cambiar allí.

## Cargar archivo de clave {#subscription-upload}

Se muestra cuando todavía no se ha activado ninguna suscripción.

Lee la clave de un archivo de su equipo (`.txt`, `.key`, `.pem` o `.lic`) en el campo de arriba.
La clave no se activa hasta que pulse **Activate New Key**.

## Acciones de la clave {#subscription-actionsMenu}

Se muestra cuando una suscripción está activa o ha caducado.

Un menú de acciones sobre la clave del servidor:

- **Reload**: vuelve a obtener el estado de la suscripción, por ejemplo después de renovarla.
- **Upload**: lee una nueva clave de un archivo en el campo de arriba, como hace el botón de carga.
- **Reset**: elimina la clave del servidor de inmediato, sin confirmación.

## Activate New Key {#subscription-activate}

Activa en este servidor la clave del campo. Cuando la clave es válida, el plan de la suscripción se
muestra debajo. Cuando su cuenta todavía no es un usuario Pro y el plan tiene una plaza libre de
usuario Pro, el diálogo pregunta si desea convertir su cuenta en usuario Pro.

Una clave que ya está activada se notifica como advertencia. El botón está deshabilitado cuando la
clave está definida en la configuración del servidor.

## Buy Pro {#subscription-buy}

Abre, en una nueva pestaña del navegador, la página del portal de Semaphore donde se compra una
suscripción PRO.

## Get 30-day free trial {#subscription-trial}

Se muestra mientras no haya una suscripción activa.

Abre, en una nueva pestaña del navegador, la página del portal de Semaphore que emite una clave
para una prueba gratuita de 30 días. Pegue la clave que reciba en el campo de arriba.
