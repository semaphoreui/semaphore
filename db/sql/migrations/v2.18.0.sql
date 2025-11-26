create table `workflow` (
    `id` integer primary key autoincrement,
    `project_id` int not null,
    `name` varchar(255) not null,
    `description` text,
    `created_at` datetime not null,
    `updated_at` datetime not null,
    foreign key (`project_id`) references `project`(`id`) on delete cascade
);

create table `workflow_node` (
    `id` integer primary key autoincrement,
    `workflow_id` int not null,
    `project_template_id` int null,
    `type` varchar(50) not null,
    `position_x` int not null,
    `position_y` int not null,
    `config_json` text,
    foreign key (`workflow_id`) references `workflow`(`id`) on delete cascade,
    foreign key (`project_template_id`) references `project__template`(`id`) on delete set null
);

create table `workflow_link` (
    `id` integer primary key autoincrement,
    `workflow_id` int not null,
    `from_node_id` int not null,
    `to_node_id` int not null,
    `condition` varchar(50) not null,
    foreign key (`workflow_id`) references `workflow`(`id`) on delete cascade,
    foreign key (`from_node_id`) references `workflow_node`(`id`) on delete cascade,
    foreign key (`to_node_id`) references `workflow_node`(`id`) on delete cascade
);

create table `workflow_run` (
    `id` integer primary key autoincrement,
    `workflow_id` int not null,
    `status` varchar(50) not null,
    `created_at` datetime not null,
    `finished_at` datetime null,
    foreign key (`workflow_id`) references `workflow`(`id`) on delete cascade
);

create table `workflow_node_run` (
    `id` integer primary key autoincrement,
    `workflow_run_id` int not null,
    `workflow_node_id` int not null,
    `status` varchar(50) not null,
    `task_id` int null,
    `started_at` datetime not null,
    `finished_at` datetime null,
    foreign key (`workflow_run_id`) references `workflow_run`(`id`) on delete cascade,
    foreign key (`workflow_node_id`) references `workflow_node`(`id`) on delete cascade,
    foreign key (`task_id`) references `task`(`id`) on delete set null
);
