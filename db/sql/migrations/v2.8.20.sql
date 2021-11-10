alter table `event` rename to `event_backup_7534878`;
create table `event`
(
    `id` integer primary key autoincrement,
    `project_id` int,
    `object_id` int,
    `object_type` varchar(20) DEFAULT '',
    `description` text,
    `created` datetime NOT NULL,
    `user_id` int,
    foreign key (`project_id`) references `project`(`id`) on delete cascade,
    foreign key (`user_id`) references `user`(`id`) on delete set null
);

insert into `event`(project_id, object_id, description, created, user_id) select project_id, object_id, description, created, user_id from `event_backup_7534878`;
drop table `event_backup_7534878`;