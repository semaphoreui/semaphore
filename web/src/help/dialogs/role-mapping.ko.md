---
form: src/components/RoleMappingForm.vue
---

# New mapping

매핑은 LDAP 또는 OpenID Connect 공급자를 통해 로그인하고 그곳의 그룹에 속한 모든 사용자에게 프로젝트의
역할을 부여합니다. Semaphore는 공급자를 통해 로그인할 때마다 사용자의 그룹을 읽고, 그에 맞게 매핑된 역할을
부여하거나 회수합니다. 따라서 그룹을 떠난 사용자는 다음 로그인 때 역할을 잃습니다. 매핑된 역할을 받은
사용자는 프로젝트의 멤버가 됩니다. 매핑은 수정할 수 없습니다. 삭제하고 새로 추가하세요.

## Provider {#roleMapping-provider}

그룹을 가져오는 LDAP 디렉터리 또는 OpenID Connect 공급자입니다. 목록에는 서버 구성에 설정된 공급자가
있습니다. 이 필드는 필수입니다.

그룹은 공급자가 그에 맞게 설정된 경우에만 읽힙니다. LDAP 공급자에는 그룹 검색 기준(`group_search_dn`)이
필요하고, OpenID Connect 공급자는 그룹이 담긴 클레임(`groups_claim`)을 지정해야 합니다. 나중에 구성에서
제거된 공급자의 매핑은 **Provider not configured**로 표시된 채 목록에 남습니다.

## Group {#roleMapping-group}

공급자가 제공하는 그대로의 그룹 이름입니다. 이 필드는 필수입니다.

- LDAP: 그룹 항목의 그룹 이름 속성 값으로, 기본값은 `cn`입니다. 예: `devops`. `cn=devops,ou=groups,dc=example,dc=com`과
  같은 전체 DN은 `cn`으로 줄어듭니다.
- OpenID Connect: 공급자가 보내는 그대로의 그룹 클레임 값 중 하나입니다.

## Project {#roleMapping-projectId}

역할이 부여되는 프로젝트입니다. 이 필드는 필수입니다. 관리자는 모든 프로젝트를 보며, 다른 사용자는 자신이
멤버인 프로젝트만 봅니다.

## 역할 {#roleMapping-role}

그룹의 멤버가 프로젝트에서 받는 역할입니다. 이 필드는 필수이며, 프로젝트를 선택하면 사용할 수 있습니다.
목록에는 그 프로젝트의 역할, 공유 역할, 기본 제공 역할이 있습니다. 다른 프로젝트를 선택하면 그곳에 없는
역할은 지워집니다.

매핑으로 부여된 **Owner** 역할은 다음 로그인 때 사라질 수 있으므로, Semaphore가 프로젝트의 마지막 소유자에게서
역할을 빼앗지 못하게 할 때 집계되지 않습니다. 직접 추가한 소유자를 적어도 한 명 유지하세요.
