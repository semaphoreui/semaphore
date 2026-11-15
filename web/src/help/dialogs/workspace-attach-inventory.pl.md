---
form: src/components/InventorySelectForm.vue
---

# Choose workspace to attach

Dołącza do tego szablonu workspace, który już istnieje w projekcie, tak aby pojawił się wśród
workspace szablonu wraz ze swoim stanem i aliasami. Workspace należy jednocześnie do jednego
szablonu. Dołączenie nie czyni workspace domyślnym: zadania nadal są wykonywane w domyślnym
workspace, dopóki nie naciśniesz **Make default**.

## Workspace {#workspaceSelect-inventoryId}

Workspace do dołączenia. Pole jest wymagane. Lista pokazuje workspace projektu, które są utworzone
dla aplikacji tego szablonu (Terraform, OpenTofu lub Terragrunt) i nie są dołączone do żadnego
szablonu; workspace odłączony od swojego szablonu przyciskiem **Detach** ponownie się tu pojawia.

Workspace zachowuje swoją nazwę, klucz i stan. Serwer odrzuca workspace, który w międzyczasie
zajął inny szablon.
