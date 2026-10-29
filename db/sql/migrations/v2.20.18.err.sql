update `role` set `permissions` = `permissions` & 127;
update `project__template_role` set `permissions` = `permissions` & 127;
update `project__workflow_role` set `permissions` = `permissions` & 127;
