alter table `project__schedule` add column if not exists `run_at` datetime null;
alter table `project__schedule` add column if not exists `one_off` boolean not null default false;
