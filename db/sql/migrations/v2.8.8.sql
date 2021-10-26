create table `project__view` (
    `id` integer primary key autoincrement,
    `title` varchar(100) not null,
    `project_id` int not null,
    `position` int not null,
    foreign key (`project_id`) references project(`id`) on delete cascade
);

alter table `project__template` add view_id int references `project__view`(id) on delete set null;