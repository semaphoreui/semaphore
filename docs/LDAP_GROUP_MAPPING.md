# LDAP Group-Based Role Mapping

This document describes the LDAP group-based role mapping functionality added to Semaphore UI to automatically assign roles to users based on their Active Directory/LDAP group memberships.

## Overview

When LDAP authentication is enabled, Semaphore can now:
1. Automatically grant admin privileges based on group membership
2. Log group-to-role mappings for future project assignment functionality
3. Retrieve and process multiple group memberships from LDAP

## Configuration

### Environment Variables

Add these new environment variables to your Docker Compose or configuration:

```bash
# LDAP group membership attribute (default: memberOf)
SEMAPHORE_LDAP_MAPPING_MEMBEROF=memberOf

# Groups that grant admin privileges (JSON array)
SEMAPHORE_LDAP_ADMIN_GROUPS=["CN=Semaphore_Admins,OU=Groups,DC=company,DC=com", "CN=IT_Admins,OU=Groups,DC=company,DC=com"]

# Group to role mappings (JSON object) - for future project assignment
SEMAPHORE_LDAP_GROUP_MAPPINGS={"CN=Semaphore_Users,OU=Groups,DC=company,DC=com": "manager", "CN=Developers,OU=Groups,DC=company,DC=com": "task_runner"}
```

### Docker Compose Example

```yaml
version: '3.8'
services:
  semaphore:
    image: semaphoreui/semaphore
    environment:
      # Existing LDAP configuration
      SEMAPHORE_LDAP_ENABLE: "true"
      SEMAPHORE_LDAP_SERVER: "ldap.company.com:389"
      SEMAPHORE_LDAP_BIND_DN: "cn=service_account,ou=Service Accounts,dc=company,dc=com"
      SEMAPHORE_LDAP_BIND_PASSWORD: "service_password"
      SEMAPHORE_LDAP_SEARCH_DN: "dc=company,dc=com"
      SEMAPHORE_LDAP_SEARCH_FILTER: "(&(sAMAccountName=%s))"
      SEMAPHORE_LDAP_MAPPING_DN: "dn"
      SEMAPHORE_LDAP_MAPPING_MAIL: "mail"
      SEMAPHORE_LDAP_MAPPING_UID: "sAMAccountName"
      SEMAPHORE_LDAP_MAPPING_CN: "cn"
      
      # New group mapping configuration
      SEMAPHORE_LDAP_MAPPING_MEMBEROF: "memberOf"
      SEMAPHORE_LDAP_ADMIN_GROUPS: '["CN=Semaphore_Admins,OU=Groups,DC=company,DC=com"]'
      SEMAPHORE_LDAP_GROUP_MAPPINGS: '{"CN=Semaphore_Users,OU=Groups,DC=company,DC=com": "manager"}'
```

### JSON Configuration File

If using a JSON configuration file instead of environment variables:

```json
{
  "ldap_enable": true,
  "ldap_server": "ldap.company.com:389",
  "ldap_binddn": "cn=service_account,ou=Service Accounts,dc=company,dc=com",
  "ldap_bindpassword": "service_password",
  "ldap_searchdn": "dc=company,dc=com",
  "ldap_searchfilter": "(&(sAMAccountName=%s))",
  "ldap_mappings": {
    "dn": "dn",
    "mail": "mail",
    "uid": "sAMAccountName",
    "cn": "cn",
    "memberof": "memberOf"
  },
  "ldap_admin_groups": [
    "CN=Semaphore_Admins,OU=Groups,DC=company,DC=com"
  ],
  "ldap_group_mappings": {
    "CN=Semaphore_Users,OU=Groups,DC=company,DC=com": "manager",
    "CN=Developers,OU=Groups,DC=company,DC=com": "task_runner"
  }
}
```

## Functionality

### Admin Role Assignment

Users who belong to any group listed in `SEMAPHORE_LDAP_ADMIN_GROUPS` will automatically receive admin privileges when they log in. This happens for both new and existing users.

### Group Processing

The system handles various LDAP group membership formats:
- Single group (string)
- Multiple groups (array of strings)
- LDAP multi-value attributes

### Role Mapping (Future Enhancement)

The group mapping infrastructure is implemented and will log intended role assignments. Project-specific role assignment will be available in a future update when additional Store interface methods are implemented.

Available roles for mapping:
- `owner` - Full project permissions
- `manager` - Can run tasks and manage resources
- `task_runner` - Can only run tasks
- `guest` - Read-only access

## Active Directory Example

For a typical Active Directory setup where users are in groups like "Semaphore_Users":

```bash
SEMAPHORE_LDAP_ADMIN_GROUPS='["CN=Semaphore_Admins,OU=Security Groups,DC=company,DC=local"]'
SEMAPHORE_LDAP_GROUP_MAPPINGS='{"CN=Semaphore_Users,OU=Security Groups,DC=company,DC=local": "task_runner"}'
```

## Testing and Validation

1. Enable LDAP with group mapping configuration
2. Log in with a user who belongs to the configured admin groups
3. Verify the user receives admin privileges
4. Check server logs for group mapping information

### Example Log Output

```
INFO[0000] User john.doe with email john.doe@company.com authorized via LDAP correctly
INFO[0000] LDAP user 123 has group CN=Semaphore_Users,OU=Groups,DC=company,DC=com mapped to role task_runner (project assignment not yet implemented)
```

## Troubleshooting

### Common Issues

1. **Groups not detected**: Verify `SEMAPHORE_LDAP_MAPPING_MEMBEROF` matches your LDAP schema
2. **Admin rights not granted**: Check that group DNs in `SEMAPHORE_LDAP_ADMIN_GROUPS` exactly match LDAP
3. **JSON parsing errors**: Ensure proper JSON escaping in environment variables

### Debug Steps

1. Check server logs for LDAP search results
2. Verify LDAP service account has read access to group membership attributes
3. Test LDAP queries manually using `ldapsearch`

## Migration from Existing LDAP Setup

This feature is fully backwards compatible. Existing LDAP configurations will continue to work without any changes. Simply add the new environment variables to enable group-based functionality.

## Security Considerations

- Ensure LDAP service account has minimal required permissions
- Regularly audit group memberships that grant admin access
- Use secure LDAP connections when possible (LDAPS or StartTLS)