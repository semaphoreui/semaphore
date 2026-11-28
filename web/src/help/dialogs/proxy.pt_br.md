---
form: src/components/ProxyForm.vue
---

# Editar Proxy

Um proxy é um host pelo qual as tarefas do projeto passam para chegar a servidores aos quais não
conseguem se conectar diretamente. Escolha-o no campo **Proxy** de um repositório, para buscar o
repositório por ele, ou de um inventário, para conectar o Ansible aos hosts por ele.

## Nome {#proxyForm-name}

Obrigatório e único no projeto. Repositórios e inventários mostram o proxy com este nome.

## Tipo {#proxyForm-type}

- **SSH (jump host)**: o Semaphore se conecta ao host por SSH e abre dali a conexão com o servidor.
  Só este tipo pode exigir outro proxy.
- **SOCKS5**, **HTTP**, **HTTPS**: um servidor proxy que transporta a conexão. O Git o usa sozinho
  para um repositório `https`; o Semaphore passa uma conexão SSH por ele, então não é preciso
  instalar outra ferramenta.

## Host {#proxyForm-host}

Obrigatório. Um nome de host ou um endereço IP, sem usuário, porta nem esquema. Um nome de host só
pode conter letras, dígitos, pontos, hifens e sublinhados.

## Port {#proxyForm-port}

Opcional, de 1 a 65535. Vazio significa `22` para um jump host SSH. Informe a porta nos outros
tipos: sem ela, o Git e as conexões SSH supõem portas diferentes.

## User {#proxyForm-user}

Só SSH: o login no jump host. Vazio significa o usuário com que o Semaphore roda. Letras, dígitos,
pontos, hifens e sublinhados, sem hífen no início. Um proxy SOCKS5 ou HTTP pega o login da
**Chave de Acesso**.

## Chave de Acesso {#proxyForm-sshKeyId}

Opcional. A chave do armazenamento de chaves usada para fazer login no próprio proxy, normalmente
não a chave dos servidores por trás dele:

- uma **Chave SSH** para um jump host SSH;
- um **Login com senha** para SOCKS5, HTTP e HTTPS. A senha é passada no ambiente, nunca em uma
  linha de comando.

## Requires Proxy {#proxyForm-requiresProxyId}

Só SSH, opcional. Outro proxy SSH do projeto pelo qual é preciso passar antes, para um jump host
que só é alcançável a partir de outro. Uma cadeia tem no máximo 5 proxies e não pode voltar a um
proxy que já está nela.
