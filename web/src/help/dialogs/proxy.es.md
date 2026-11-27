---
form: src/components/ProxyForm.vue
---

# Editar Proxy

Un proxy es un host por el que pasan las tareas del proyecto para llegar a servidores a los que no
pueden conectarse directamente. Elíjalo en el campo **Proxy** de un repositorio, para obtener el
repositorio a través de él, o de un inventario, para conectar Ansible a los hosts a través de él.

## Nombre {#proxyForm-name}

Obligatorio y único en el proyecto. Los repositorios y los inventarios muestran el proxy con este nombre.

## Tipo {#proxyForm-type}

- **SSH (jump host)**: Semaphore se conecta al host por SSH y abre desde allí la conexión al
  servidor. Solo este tipo puede requerir otro proxy.
- **SOCKS5**, **HTTP**, **HTTPS**: un servidor proxy que transporta la conexión. Git lo usa por sí
  mismo para un repositorio `https`; Semaphore pasa una conexión SSH a través de él, así que no hay
  que instalar ninguna otra herramienta.

## Host {#proxyForm-host}

Obligatorio. Un nombre de host o una dirección IP, sin usuario, puerto ni esquema. Un nombre de host
solo puede contener letras, dígitos, puntos, guiones y guiones bajos.

## Port {#proxyForm-port}

Opcional, de 1 a 65535. Vacío significa `22` para un jump host SSH. Indique el puerto en los demás
tipos: sin él, Git y las conexiones SSH suponen puertos distintos.

## User {#proxyForm-user}

Solo SSH: el usuario en el jump host. Vacío significa el usuario con el que se ejecuta Semaphore.
Letras, dígitos, puntos, guiones y guiones bajos, sin un guion al principio. Un proxy SOCKS5 o HTTP
toma su usuario de la **Clave de Acceso**.

## Clave de Acceso {#proxyForm-sshKeyId}

Opcional. La clave del almacén de claves con la que se inicia sesión en el propio proxy,
normalmente no la clave de los servidores que hay detrás:

- una **Clave SSH** para un jump host SSH;
- un **Iniciar sesión con contraseña** para SOCKS5, HTTP y HTTPS. La contraseña se pasa en el
  entorno, nunca en una línea de comandos.

## Requires Proxy {#proxyForm-requiresProxyId}

Solo SSH, opcional. Otro proxy SSH del proyecto por el que hay que pasar primero, para un jump host
al que solo se llega desde otro. Una cadena tiene como máximo 5 proxies y no puede volver a un
proxy que ya está en ella.
