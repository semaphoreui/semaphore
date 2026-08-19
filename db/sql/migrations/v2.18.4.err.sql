drop table project__template_environment;
alter table `project__template` add `environment_id` int references `project__environment`(`id`) on delete set null;
