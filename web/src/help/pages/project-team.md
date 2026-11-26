---
form: src/views/project/Team.vue
skip: team-rowRolesSelect
---

# Team

The members of the project and the roles they hold. A role decides what a member may do in the
project: run tasks, change templates and other resources, manage the team. A member may hold
several roles and gets every permission any of them gives.

- **New Team Member** adds a user to the project with a role. The button is shown to members who
  may manage the team.
- The **Role** column changes the roles of a member at once, with no Save. A member who may not
  manage the team sees the roles without being able to change them.
- A role with a lock comes from a group of the user in LDAP or OIDC, mapped to that role by the
  administrator. It can not be removed here: it goes when the user leaves the group or the
  mapping is removed.
- The project needs at least one Owner. Semaphore refuses to remove its last Owner or to take the
  Owner role away from them, unless an administrator of Semaphore does it.
- **Leave Project** takes you out of the project. It is shown only when the administrator allows
  members to leave projects; the last Owner can not leave either.
