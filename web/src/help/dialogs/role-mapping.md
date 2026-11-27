---
form: src/components/RoleMappingForm.vue
---

# New mapping

A mapping gives a role in a project to every user who signs in through an LDAP or OpenID Connect
provider and belongs to a group there. Semaphore reads the groups of the user at each sign-in
through the provider and gives or takes away the mapped roles to match, so a user who leaves the
group loses the role at their next sign-in. A mapped role makes the user a member of the project.
A mapping can not be edited: delete it and add a new one.

## Provider {#roleMapping-provider}

The LDAP directory or the OpenID Connect provider the group comes from. The list holds the
providers set up in the configuration of the server. The field is required.

The groups are read only when the provider is set up for it: an LDAP provider needs a group
search base (`group_search_dn`), an OpenID Connect provider names the claim with the groups
(`groups_claim`). A mapping of a provider that is later removed from the configuration stays in
the list, marked **Provider not configured**.

## Group {#roleMapping-group}

The name of the group as the provider gives it. The field is required.

- LDAP: the value of the group name attribute of the group entry, `cn` by default, for example
  `devops`. A full DN such as `cn=devops,ou=groups,dc=example,dc=com` is reduced to its `cn`.
- OpenID Connect: one of the values of the groups claim, as the provider sends it.

## Project {#roleMapping-projectId}

The project the role is given in. The field is required. An administrator sees every project; any
other user sees only the projects they are a member of.

## Role {#roleMapping-role}

The role the members of the group get in the project. The field is required and is available once
a project is selected. The list holds the roles of that project, the shared roles and the
built-in roles. When you pick another project, a role that does not exist there is cleared.

An **Owner** role given by a mapping does not count when Semaphore keeps you from taking the role
from the last owner of a project, because it may go away at the next sign-in. Keep at least one
owner added by hand.
