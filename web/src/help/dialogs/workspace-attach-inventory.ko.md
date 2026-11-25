---
form: src/components/InventorySelectForm.vue
---

# Choose workspace to attach

프로젝트에 이미 있는 워크스페이스를 이 템플릿에 연결하여, 템플릿의 워크스페이스 사이에 상태 및 별칭과 함께
표시되게 합니다. 워크스페이스는 한 번에 하나의 템플릿에만 속합니다. 연결해도 워크스페이스가 기본 워크스페이스가
되지는 않습니다. **Make default**를 누를 때까지 작업은 계속 기본 워크스페이스에서 실행됩니다.

## Workspace {#workspaceSelect-inventoryId}

연결할 워크스페이스입니다. 이 필드는 필수입니다. 목록에는 이 템플릿의 앱(Terraform, OpenTofu 또는 Terragrunt)용으로
만들어졌고 어떤 템플릿에도 연결되지 않은 프로젝트의 워크스페이스가 표시됩니다. **Detach**로 템플릿에서 분리한
워크스페이스는 여기에 다시 나타납니다.

워크스페이스는 이름, 키, 상태를 유지합니다. 그사이 다른 템플릿이 가져간 워크스페이스는 서버가 거부합니다.
