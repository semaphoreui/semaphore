-- read_project (128): a member without it sees only the objects granted to their roles. Nobody loses access.
update `role` set `permissions` = `permissions` | 128;
update `project__template_role` set `permissions` = `permissions` | 128 where `permissions` <> 0;
update `project__workflow_role` set `permissions` = `permissions` | 128 where `permissions` <> 0;
