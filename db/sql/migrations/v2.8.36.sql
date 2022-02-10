alter table `project__template` add allow_override_args_in_task bool not null default false;
alter table `task` add arguments text;
alter table `project__template` drop column `override_args`;
