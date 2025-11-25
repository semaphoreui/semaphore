-- Create workflows table
create table `project__workflow` (
    `id` integer primary key autoincrement,
    `project_id` int not null,
    `name` varchar(255) not null,
    `description` text,
    `created_at` datetime not null default current_timestamp,
    `updated_at` datetime not null default current_timestamp,

    foreign key (`project_id`) references `project`(`id`) on delete cascade
);

-- Create workflow_nodes table
create table `project__workflow_node` (
    `id` integer primary key autoincrement,
    `workflow_id` int not null,
    `task_template_id` int,
    `type` varchar(20) not null default 'task',
    `name` varchar(255) not null,
    `position_x` float not null default 0,
    `position_y` float not null default 0,
    `config` text,

    foreign key (`workflow_id`) references `project__workflow`(`id`) on delete cascade,
    foreign key (`task_template_id`) references `project__template`(`id`) on delete set null
);

-- Create workflow_links table (edges between nodes)
create table `project__workflow_link` (
    `id` integer primary key autoincrement,
    `workflow_id` int not null,
    `from_node_id` int not null,
    `to_node_id` int not null,
    `condition` varchar(20) not null default 'success',

    foreign key (`workflow_id`) references `project__workflow`(`id`) on delete cascade,
    foreign key (`from_node_id`) references `project__workflow_node`(`id`) on delete cascade,
    foreign key (`to_node_id`) references `project__workflow_node`(`id`) on delete cascade
);

-- Create workflow_runs table (execution history)
create table `project__workflow_run` (
    `id` integer primary key autoincrement,
    `workflow_id` int not null,
    `project_id` int not null,
    `user_id` int,
    `status` varchar(20) not null default 'pending',
    `start` datetime,
    `end` datetime,
    `message` text,

    foreign key (`workflow_id`) references `project__workflow`(`id`) on delete cascade,
    foreign key (`project_id`) references `project`(`id`) on delete cascade,
    foreign key (`user_id`) references `user`(`id`) on delete set null
);

-- Create workflow_node_runs table (track individual node execution)
create table `project__workflow_node_run` (
    `id` integer primary key autoincrement,
    `workflow_run_id` int not null,
    `node_id` int not null,
    `task_id` int,
    `status` varchar(20) not null default 'pending',
    `start` datetime,
    `end` datetime,
    `message` text,

    foreign key (`workflow_run_id`) references `project__workflow_run`(`id`) on delete cascade,
    foreign key (`node_id`) references `project__workflow_node`(`id`) on delete cascade,
    foreign key (`task_id`) references `task`(`id`) on delete set null
);

-- Create indexes for better performance
create index `idx_workflow_project` on `project__workflow`(`project_id`);
create index `idx_workflow_node_workflow` on `project__workflow_node`(`workflow_id`);
create index `idx_workflow_link_workflow` on `project__workflow_link`(`workflow_id`);
create index `idx_workflow_link_from_node` on `project__workflow_link`(`from_node_id`);
create index `idx_workflow_link_to_node` on `project__workflow_link`(`to_node_id`);
create index `idx_workflow_run_workflow` on `project__workflow_run`(`workflow_id`);
create index `idx_workflow_run_project` on `project__workflow_run`(`project_id`);
create index `idx_workflow_node_run_workflow_run` on `project__workflow_node_run`(`workflow_run_id`);
create index `idx_workflow_node_run_node` on `project__workflow_node_run`(`node_id`);
