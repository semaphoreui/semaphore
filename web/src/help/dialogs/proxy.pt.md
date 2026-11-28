---
form: src/components/ProxyForm.vue
---

# Editar Proxy

Um proxy é um host pelo qual as tarefas do projeto passam para chegar a servidores aos quais não se
podem ligar diretamente. Escolha-o no campo **Proxy** de um repositório, para obter o repositório
através dele, ou de um inventário, para ligar o Ansible aos hosts através dele.

## Nome {#proxyForm-name}

Obrigatório e único no projeto. Os repositórios e os inventários mostram o proxy com este nome.

## Tipo {#proxyForm-type}

- **SSH (jump host)**: o Semaphore liga-se ao host por SSH e abre a partir daí a ligação ao
  servidor. Só este tipo pode requerer outro proxy.
- **SOCKS5**, **HTTP**, **HTTPS**: um servidor proxy que transporta a ligação. O Git usa-o por si
  para um repositório `https`; o Semaphore passa uma ligação SSH por ele, pelo que não é preciso
  instalar outra ferramenta.

## Host {#proxyForm-host}

Obrigatório. Um nome de host ou um endereço IP, sem utilizador, porta nem esquema. Um nome de host só
pode conter letras, algarismos, pontos, hífenes e sublinhados.

## Port {#proxyForm-port}

Opcional, de 1 a 65535. Vazio significa `22` para um jump host SSH. Indique a porta nos outros
tipos: sem ela, o Git e as ligações SSH assumem portas diferentes.

## User {#proxyForm-user}

Só SSH: o login no jump host. Vazio significa o utilizador com que o Semaphore é executado. Letras,
algarismos, pontos, hífenes e sublinhados, sem hífen no início. Um proxy SOCKS5 ou HTTP obtém o
login da **Chave de Acesso**.

## Chave de Acesso {#proxyForm-sshKeyId}

Opcional. A chave do armazenamento de chaves com que se faz login no próprio proxy, normalmente não a
chave dos servidores por trás dele:

- uma **Chave SSH** para um jump host SSH;
- um **Login com senha** para SOCKS5, HTTP e HTTPS. A senha é passada no ambiente, nunca numa
  linha de comandos.

## Requires Proxy {#proxyForm-requiresProxyId}

Só SSH, opcional. Outro proxy SSH do projeto pelo qual é preciso passar primeiro, para um jump host
que só é alcançável a partir de outro. Uma cadeia tem no máximo 5 proxies e não pode voltar a um
proxy que já está nela.
