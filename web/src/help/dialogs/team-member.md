---
form: src/components/TeamMemberForm.vue
---

# New Team Member

Adds a user to the project with a role. The role decides what the member may see and do in the
project; nothing else of their account changes. When the administrator has turned invitations
on, the user gets an invitation instead and becomes a member once they accept it.

## Email or Username {#teamMember-inviteType}

Shown when users can be invited both by e-mail and by username.

How you pick the person: **Email** to invite an e-mail address, **Username** to pick a user who
has an account on this Semaphore server already.

## Email {#teamMember-email}

Shown when the person is invited by e-mail.

The e-mail address of the person to invite. They become a member of the project once they
accept the invitation.

## User {#teamMember-userId}

Shown when the person is picked by username.

The user to add, listed as the username and the full name. The field is required. The list
holds every user of the Semaphore server who is not a member of the project yet: a person
without an account has to get one from an administrator first.

## Role {#teamMember-role}

What the member may do in the project. The field is required. The built-in roles:

- **Owner**: everything, including the settings of the project, its team and its removal.
- **Manager**: runs tasks and workflows, approves workflows, and creates, changes and deletes
  task templates, inventories, keys, repositories, variable groups, schedules, integrations and
  workflows. A manager can not change the settings of the project or its team.
- **Task Runner**: sees the project, runs tasks and workflows and approves workflows. A task
  runner can not change anything in the project.
- **Guest**: sees the project, its tasks and their logs, and changes nothing.

The list also holds the roles made for this project (**Project roles**) and the roles shared by
every project (**Shared roles**). Such a role grants the permissions set for it on the Roles
page. An invitation can carry a built-in role only.

A member who can manage the team may give any role, **Owner** included. The dialog gives one
role; add more in the **Role** column of the team list afterwards.
