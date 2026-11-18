---
form: src/components/InventoryHostForm.vue
---

# Upravit hosta

Host spravovaného inventáře. Úloha, která inventář používá, běží na těch hostech, které jsou v
seznamu hostů zapnuté.

## Název {#inventoryHostForm-name}

Název, pod kterým Ansible hosta zná, a adresa, ke které se připojuje, pokud proměnné nenastavují
`ansible_host`. Pole je povinné. Název je v rámci inventáře jedinečný a záleží na velikosti písmen:
`web1` a `Web1` jsou dva hosté. Nesmí obsahovat mezery, čárky, hranaté závorky ani dvojtečky a má
nejvýše 250 znaků.

## Popis {#inventoryHostForm-description}

Nepovinné. Poznámka pro lidi, zobrazuje se v seznamu hostů. Ansible ji nedostane.

## Proměnné {#inventoryHostForm-variables}

Nepovinné. Proměnné tohoto hosta zapsané jako mapování YAML nebo JSON:

```yaml
ansible_host: 10.0.0.11
ansible_port: 2222
```

Proměnná hosta má přednost před stejnojmennou proměnnou inventáře. Kotvy, aliasy a slučovací klíče
YAML nejsou přijímány.
