drop table `project__workflow_role`;

update `role` set `permissions` = `permissions` & 15;
