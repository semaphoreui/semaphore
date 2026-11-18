---
form: src/components/InventoryGroupForm.vue
---

# Upravit skupinu

Skupina hostů spravovaného inventáře. Playbook běží na části hostů podle názvu skupiny a proměnná
skupiny se nastaví jednou pro všechny její hosty. Hosté se do skupiny přidávají na kartě Hosts, v
dialogu hosta nebo akcemi pro vybrané hosty.

## Název {#inventoryGroupForm-name}

Název, pod kterým Ansible skupinu zná, například v `hosts: web` playbooku. Pole je povinné. Název je
v rámci inventáře jedinečný a záleží na velikosti písmen: `web` a `Web` jsou dvě skupiny. Nesmí
obsahovat mezery, čárky, hranaté závorky ani dvojtečky, má nejvýše 250 znaků a nesmí být `all` ani
`ungrouped`. Ansible varuje před názvy se znaky jinými než písmena, číslice a podtržítka, například
`web-eu`, takový název ale přesto funguje.

## Popis {#inventoryGroupForm-description}

Nepovinné. Poznámka pro lidi, zobrazuje se v seznamu skupin. Ansible ji nedostane.

## Child groups {#inventoryGroupForm-children}

Nepovinné. Skupiny uvnitř této skupiny. Hosté podřízené skupiny patří také do této skupiny. Skupina
nemůže být svým vlastním předkem.

## Proměnné {#inventoryGroupForm-variables}

Nepovinné. Proměnné této skupiny zapsané jako mapování YAML nebo JSON:

```yaml
http_port: 8080
ntp_server: ntp.example.com
```

Dostane je každý host skupiny. Proměnná hosta má přednost před stejnojmennou proměnnou skupiny a
proměnná skupiny má přednost před proměnnou inventáře. Kotvy, aliasy a slučovací klíče YAML nejsou
přijímány.
