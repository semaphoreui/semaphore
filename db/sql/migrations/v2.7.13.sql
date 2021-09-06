alter table project__template_schedule rename to project__schedule;
alter table `project__schedule` add `id` integer primary key autoincrement;
alter table `project__schedule` add `project_id` int not null references project(`id`);
