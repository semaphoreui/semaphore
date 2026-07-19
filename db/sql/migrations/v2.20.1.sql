alter table `project__template` add `allow_override_env_in_task` boolean not null default false;
alter table `task` add `environment_ids` text;
