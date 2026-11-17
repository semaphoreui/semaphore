---
form: src/components/TemplateSelectForm.vue
---

# Choose workspace to attach

Dołącza ten workspace do szablonu, tak aby pojawił się wśród workspace szablonu wraz ze swoim
stanem i aliasami. Workspace należy jednocześnie do jednego szablonu: po dołączeniu znika ze strony
Inwentarz i jest zarządzany w szablonie. Dołączenie nie czyni go domyślnym workspace szablonu: aby
to zrobić, otwórz szablon i naciśnij **Make default**.

## Template {#inventoryAttach-templateId}

Szablon, do którego ma zostać dołączony workspace. Pole jest wymagane. Lista pokazuje szablony
projektu utworzone dla tej samej aplikacji co workspace, Terraform lub OpenTofu.

Workspace zachowuje swoją nazwę, klucz i stan. Aby później przenieść go do innego szablonu,
naciśnij **Detach** w szablonie, do którego jest dołączony: wraca on wtedy na stronę Inwentarz.
Domyślnego workspace szablonu nie można odłączyć.
