---
form: src/components/HostConfigForm.vue
---

# Edit mapping

Um mapeamento indica ao Semaphore que credencial do Armazenamento de Chaves deve utilizar para um
host SSH ou um URL Git. Um repositório tem uma chave, mas uma tarefa pode também chegar a outros
servidores: submódulos, funções e coleções de `requirements.yml`, módulos do Terraform,
repositórios de inventário. Os mapeamentos do projeto aplicam-se a todas as ligações Git e SSH das
suas tarefas, e aos comandos Git que o Semaphore executa para o projeto, como a verificação de
novos commits num repositório. Nenhum ficheiro do seu repositório é alterado.

## Tipo {#hostConfigForm-type}

O que o mapeamento faz corresponder:

- **Host**: um nome de host SSH. Todas as ligações SSH de uma tarefa a esse host utilizam a chave
  do mapeamento: o Git por SSH e o Ansible a ligar-se a um host do inventário com esse nome. O
  mapeamento tem prioridade sobre a configuração SSH do servidor para o mesmo host.
- **URL**: um URL Git ou o início de um. O Git liga-se a todos os repositórios sob esse URL com a
  credencial do mapeamento.

Mudar para **Host** limpa uma **Credential** que um mapeamento de host não pode utilizar.

## Host ou URL {#hostConfigForm-host}

O campo é obrigatório. Pode existir um mapeamento de cada tipo para cada host ou URL.

Para **Host**: o nome do host exatamente como o Git ou o Ansible se ligam a ele, por exemplo
`github.com` ou `10.0.0.21`. Apenas letras, algarismos, pontos, hífenes e sublinhados: sem
utilizador, sem porta, sem caracteres universais.

Para **URL**: um endereço que começa por `https://` ou `http://`, sem credenciais de acesso e sem
espaços, aspas ou `=`. Indique um repositório inteiro, ou um prefixo terminado em `/` para
abranger um grupo de repositórios:

```
https://github.com/acme/
https://github.com/acme/infra.git
```

Quando vários mapeamentos correspondem, ganha o mais longo: o mapeamento do repositório acima
substitui o mapeamento de `acme` para esse repositório.

Com uma chave SSH, um mapeamento de URL faz com que o Git se ligue por SSH ao host do URL na porta
`22`, seja qual for a porta indicada no URL.

## Credential {#hostConfigForm-sshKeyId}

A chave do Armazenamento de Chaves utilizada para as ligações a que o mapeamento corresponde. O
campo é obrigatório.

- **Host** aceita apenas uma **Chave SSH**. Quando a chave tem um login, o SSH liga-se como esse
  utilizador.
- **URL** aceita uma **Chave SSH** ou um **Login com senha**. Uma chave SSH liga-se com o seu
  login, ou como `git` quando não tem nenhum. Um login com palavra-passe funciona apenas com um
  URL `https://`; para um token, coloque o token na palavra-passe e deixe o login vazio.

Um login com palavra-passe chega ao Git através do seu ambiente, não da sua linha de comandos, e o
log da tarefa mostra apenas o URL original.
