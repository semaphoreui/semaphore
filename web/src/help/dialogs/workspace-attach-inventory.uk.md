---
form: src/components/InventorySelectForm.vue
---

# Choose workspace to attach

Прикріплює до цього шаблону workspace, який уже існує в проєкті, щоб він з’явився серед workspace
шаблону разом зі своїм станом і псевдонімами. Workspace належить одночасно лише одному шаблону.
Прикріплення не робить workspace типовим: завдання і далі виконуються у workspace за
замовчуванням, доки ви не натиснете **Make default**.

## Workspace {#workspaceSelect-inventoryId}

Workspace, який слід прикріпити. Поле обов’язкове. Список показує workspace проєкту, створені для
застосунку цього шаблону (Terraform, OpenTofu або Terragrunt), які не прикріплено до жодного
шаблону; workspace, від’єднаний від свого шаблону кнопкою **Detach**, знову з’являється тут.

Workspace зберігає свою назву, ключ і стан. Сервер відхиляє workspace, який тим часом зайняв
інший шаблон.
