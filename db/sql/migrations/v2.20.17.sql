-- Workflow verbs: run_workflows (16), edit_workflows (32), approve_workflows (64). Nobody loses access.
update `role` set `permissions` = `permissions` | 16 | 64 where `permissions` & 1 <> 0;
update `role` set `permissions` = `permissions` | 32 where `permissions` & 4 <> 0;

-- The twin of project__template_role.
create table `project__workflow_role`
(
    `id`          integer primary key autoincrement,
    `workflow_id` int    not null,
    `role_id`     int    not null,
    `project_id`  int    not null,
    `permissions` bigint not null default 0,

    foreign key (`workflow_id`) references project__workflow_template (`id`) on delete cascade,
    foreign key (`role_id`) references `role` (`id`) on delete restrict,
    foreign key (`project_id`) references project (`id`) on delete cascade,
    unique (`workflow_id`, `role_id`)
);
