---
form: src/components/HostConfigForm.vue
---

# Edit mapping

Um mapeamento informa ao Semaphore qual credencial do Armazenamento de Chaves usar para um host
SSH ou uma URL Git. Um repositório leva uma chave, mas uma tarefa também pode acessar outros
servidores: submódulos, roles e coleções do `requirements.yml`, módulos do Terraform, repositórios
de inventário. Os mapeamentos do projeto se aplicam a todas as conexões Git e SSH das tarefas dele
e aos comandos Git que o Semaphore executa para o projeto, como a verificação de novos commits em
um repositório. Nenhum arquivo do seu repositório é alterado.

## Tipo {#hostConfigForm-type}

O que o mapeamento corresponde:

- **Host**: um nome de host SSH. Toda conexão SSH de uma tarefa a esse host usa a chave do
  mapeamento: o Git por SSH e o Ansible ao se conectar a um host de inventário com esse nome. O
  mapeamento tem precedência sobre a configuração SSH do servidor para o mesmo host.
- **URL**: uma URL Git ou o início de uma. O Git se conecta a todos os repositórios sob essa URL
  com a credencial do mapeamento.

Mudar para **Host** limpa uma **Credential** que um mapeamento de host não pode usar.

## Host ou URL {#hostConfigForm-host}

O campo é obrigatório. Pode existir um mapeamento de cada tipo para cada host ou URL.

Para **Host**: o nome do host exatamente como o Git ou o Ansible se conectam a ele, por exemplo
`github.com` ou `10.0.0.21`. Apenas letras, dígitos, pontos, hifens e sublinhados: sem usuário,
sem porta, sem curingas.

Para **URL**: um endereço que começa com `https://` ou `http://`, sem login e sem espaços, aspas ou
`=`. Informe um repositório inteiro ou um prefixo que termine com `/` para abranger um grupo de
repositórios:

```
https://github.com/acme/
https://github.com/acme/infra.git
```

Quando vários mapeamentos correspondem, o mais longo vence: o mapeamento do repositório acima
substitui o mapeamento de `acme` para esse repositório específico.

Com uma chave SSH, um mapeamento de URL faz o Git se conectar por SSH ao host da URL na porta
`22`, seja qual for a porta indicada na URL.

## Credential {#hostConfigForm-sshKeyId}

A chave do Armazenamento de Chaves usada para as conexões às quais o mapeamento corresponde. O
campo é obrigatório.

- **Host** aceita apenas uma **Chave SSH**. Quando a chave tem um login, o SSH se conecta como esse
  usuário.
- **URL** aceita uma **Chave SSH** ou um **Login com senha**. Uma chave SSH se conecta com o login
  dela, ou como `git` quando não tem um. Um login com senha funciona apenas com uma URL `https://`;
  para um token, coloque o token na senha e deixe o login vazio.

Um login com senha chega ao Git pelo ambiente dele, não pela linha de comando, e o log da tarefa
mostra apenas a URL original.
