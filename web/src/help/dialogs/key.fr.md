---
form: src/components/KeyForm.vue
---

# Modifier Key

Une clé du Magasin de clés contient un identifiant que le projet utilise sans l'afficher : la clé
SSH ou le mot de passe avec lequel Ansible se connecte aux hôtes, les identifiants d'un dépôt Git,
le mot de passe d'un coffre Ansible, le secret d'une intégration. Les inventaires, les dépôts, les
modèles de tâches et les intégrations font référence à une clé par son nom. Une fois enregistré, le
secret n'est plus jamais renvoyé au navigateur.

## Nom de la clé {#key-name}

Le nom sous lequel la clé apparaît dans le Magasin de clés et dans les sélecteurs de clé des
inventaires, des dépôts, des modèles de tâches et des intégrations. Le champ est obligatoire.
Choisissez un nom qui dit ce que la clé ouvre, par exemple `Deploy key of the web servers`.

## Source {#key-source}

L'endroit où le secret de la clé est conservé. Les onglets :

- **Local** : le secret est saisi dans cette boîte de dialogue et stocké dans la base de données de
  Semaphore, chiffré lorsque l'administrateur a défini une clé de chiffrement pour les clés d'accès.
  C'est la valeur par défaut.
- **Storage** : le secret est conservé dans un stockage de secrets externe du projet, choisi dans
  **Storage** ci-dessous. L'onglet n'existe que dans l'édition Pro. Les stockages eux-mêmes se
  configurent dans l'onglet **Storages** du Magasin de clés.
- **Env** : le secret est lu dans une variable d'environnement du processus du serveur Semaphore.
  Saisissez le nom de la variable dans le champ sous les onglets.
- **File** : le secret est lu dans un fichier sur le serveur Semaphore. Saisissez le chemin absolu
  du fichier dans le champ sous les onglets. Le fichier doit se trouver dans le répertoire des
  secrets du serveur (`dirs.secrets` dans sa configuration, `/tmp/semaphore` par défaut) ; un
  chemin situé en dehors, ou contenant `..`, échoue lorsque la clé est utilisée. Un saut de ligne
  final du fichier est supprimé.

Avec **Env** et **File**, Semaphore n'écrit jamais le secret : il lit la valeur chaque fois qu'une
tâche utilise la clé, vous modifiez donc le secret là où il se trouve. La valeur doit correspondre
au **Type** de la clé. Une clé **Chaîne** prend la valeur telle quelle. Une clé **Clé SSH** ou
**Connexion avec mot de passe** attend un JSON contenant les champs de ce type, par exemple :

```json
{"login": "deploy", "password": "s3cret"}
```

```json
{"login": "deploy", "passphrase": "", "private_key": "-----BEGIN OPENSSH PRIVATE KEY-----\n..."}
```

Lorsque vous modifiez une clé, les onglets sont verrouillés tant que **Remplacer** n'est pas coché.
Une clé synchronisée depuis un stockage de secrets conserve sa source, et son nom et son type ne
peuvent pas être modifiés.

## Storage {#key-sourceStorageId}

Affiché lorsque l'onglet **Storage** est choisi.

Le stockage de secrets du projet qui contient le secret : HashiCorp Vault, OpenBao, Devolutions
Server, AWS Secrets Manager ou Azure Key Vault, tel que configuré dans l'onglet **Storages** du
Magasin de clés.

- Un stockage qui n'est pas en lecture seule reçoit le secret que vous saisissez dans cette boîte
  de dialogue : Semaphore l'y écrit lors de l'enregistrement. Supprimer la clé supprime aussi le
  secret du stockage. Une clé conservée dans un tel stockage ne peut pas être déplacée ensuite
  vers un autre stockage.
- Un stockage en lecture seule est seulement lu. La boîte de dialogue n'affiche alors aucun champ
  pour le secret, et Semaphore le lit dans le stockage chaque fois qu'une tâche utilise la clé.

## Source Key {#key-sourceStorageKey}

Affiché lorsqu'un **Storage** est sélectionné.

L'emplacement du secret dans ce stockage, par exemple le chemin du secret dans HashiCorp Vault.

## Type {#key-type}

Le genre d'identifiant que contient la clé. Le champ est obligatoire, et les champs situés en
dessous changent en fonction du type.

- **Clé SSH** : une clé privée, avec un identifiant et une phrase secrète optionnels. Ansible se
  connecte aux hôtes avec elle, et Git clone avec elle un dépôt dont l'URL est `ssh`.
- **Connexion avec mot de passe** : un identifiant et un mot de passe. Ansible se connecte aux
  hôtes avec eux ou devient un autre utilisateur avec eux (**Identifiants sudo** d'un inventaire) ;
  Git clone avec eux un dépôt `https` ; le mot de passe ouvre un coffre Ansible d'un modèle de
  tâche.
- **Chaîne** : une seule valeur secrète sans identifiant : un jeton, une clé d'API, un mot de passe.
  Utilisez-la pour le secret d'une intégration. Elle ne peut pas servir de clé SSH, d'identifiants
  sudo, de mot de passe d'un coffre ni de clé d'accès d'un dépôt.
- **Aucun** : aucun identifiant. Utilisez-le pour un dépôt qui ne demande pas de connexion (un
  dépôt `https` public, une URL `file` ou un chemin local) et pour un inventaire dont les hôtes
  n'ont pas besoin d'identifiants ou définissent eux-mêmes la connexion.

## Nom d'utilisateur {#key-loginPasswordLogin}

Affiché lorsque le type est **Connexion avec mot de passe**. Optionnel.

Le nom d'utilisateur de l'identifiant.

- Ansible se connecte aux hôtes avec cet utilisateur (`--user`) ; en tant qu'**Identifiants sudo**
  d'un inventaire, c'est l'utilisateur qu'Ansible devient (`--become-user`). Lorsque le champ est
  vide, Ansible utilise l'utilisateur indiqué par l'inventaire ou par sa propre configuration.
- Pour un dépôt `https`, Git se connecte avec cet utilisateur et le mot de passe. Laissez le champ
  vide pour un jeton que le serveur Git accepte seul : le mot de passe est alors envoyé comme nom
  d'utilisateur.

## Mot de passe {#key-loginPasswordPassword}

Affiché lorsque le type est **Connexion avec mot de passe**. Obligatoire.

Le mot de passe de l'identifiant. Semaphore répond avec lui aux invites de mot de passe d'Ansible :
le mot de passe SSH (`--ask-pass`) pour les identifiants de l'utilisateur, le mot de passe `become`
(`--ask-become-pass`) pour les identifiants sudo, et le mot de passe du coffre pour un coffre
Ansible. Pour un dépôt `https`, c'est le mot de passe ou le jeton d'accès du serveur Git.

## Valeur {#key-stringValue}

Affiché lorsque le type est **Chaîne**. Obligatoire.

La valeur secrète elle-même, par exemple le jeton auquel une intégration compare les requêtes
qu'elle reçoit. Elle est stockée telle qu'elle est saisie.

## Nom d'utilisateur {#key-sshLogin}

Affiché lorsque le type est **Clé SSH**. Optionnel.

L'utilisateur avec lequel Ansible se connecte aux hôtes (`--user`). Lorsque le champ est vide,
Ansible utilise l'utilisateur indiqué par l'inventaire ou par sa propre configuration.

Pour un dépôt Git, placez plutôt l'utilisateur dans l'URL du dépôt, comme dans
`git@github.com:org/repo.git`.

## Passphrase {#key-sshPassphrase}

Affiché lorsque le type est **Clé SSH**. Optionnel.

La phrase secrète avec laquelle la clé privée est chiffrée. Laissez le champ vide pour une clé sans
phrase secrète. Une phrase secrète erronée fait échouer chaque tâche qui utilise la clé au moment
où la clé est chargée. Une clé générée par Semaphore n'a pas de phrase secrète : le contenu de ce
champ est alors ignoré.

## Generate SSH Key {#key-generateSshKey}

Affiché lorsque le type est **Clé SSH**.

Semaphore crée une nouvelle paire de clés lors de l'enregistrement : une clé privée RSA de 2048
bits, qui reste dans la clé, et sa clé publique, qui est affichée une fois dans une boîte de
dialogue après l'enregistrement, puis sous **Public key** chaque fois que vous modifiez la clé.
Ajoutez la clé publique à `~/.ssh/authorized_keys` sur les hôtes, ou comme clé de déploiement du
dépôt Git.

Le champ de la clé privée est verrouillé tant que la case est cochée. Sur une clé existante, cochez
d'abord **Remplacer** : l'ancienne paire de clés est remplacée. Une clé dont le secret est lu depuis
**Env**, **File** ou un stockage en lecture seule ne peut pas être générée, car Semaphore n'a nulle
part où l'écrire.

## Clé privée {#key-sshPrivateKey}

Affiché lorsque le type est **Clé SSH**. Obligatoire, sauf si **Generate SSH Key** est coché.

La clé privée, au format OpenSSH ou PEM, avec ses lignes `BEGIN` et `END`. Pour une tâche,
Semaphore la charge dans un agent SSH qui lui est propre, qu'Ansible et Git utilisent pendant la
durée de la tâche.

## Clé publique {#key-publicKey}

Affiché lorsque vous modifiez une clé SSH générée par Semaphore, tant que **Remplacer** n'est pas
coché.

La moitié publique de la paire de clés générée. **Show** la déplie, le bouton de copie la place
dans le presse-papiers. Ajoutez-la à `~/.ssh/authorized_keys` sur les hôtes, ou comme clé de
déploiement du dépôt Git.

C'est la clé publique de la dernière paire générée par Semaphore. Après avoir remplacé la clé
privée à la main avec **Remplacer**, la clé affichée ici ne lui correspond plus.

## Remplacer {#key-overrideSecret}

Affiché lorsque vous modifiez une clé existante.

Semaphore n'affiche plus jamais un secret enregistré : les champs du secret sont donc vides et
verrouillés. Cochez la case pour remplacer le secret : la source, le type et les champs du secret
s'ouvrent, et tout leur contenu est enregistré à nouveau ; saisissez donc le secret entier, pas
seulement la partie qui change. Sans elle, l'enregistrement ne modifie que le nom de la clé.

L'enregistrement d'une clé supprime aussi les copies que le serveur Semaphore conserve des dépôts
qui l'utilisent : la tâche suivante les clone à nouveau.
