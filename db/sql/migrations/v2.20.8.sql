alter table `role` add column `is_builtin` bool not null default false;

insert into `role` (`is_builtin`, `slug`, `name`, `permissions`, `project_id`)
values (true, 'owner', 'Owner', 15, null),
       (true, 'manager', 'Manager', 5, null),
       (true, 'task_runner', 'Task Runner', 1, null),
       (true, 'guest', 'Guest', 0, null);
