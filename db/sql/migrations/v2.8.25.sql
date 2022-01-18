alter table `project__template` add survey_vars longtext;
alter table `project__template` add autorun boolean default false;
alter table `project__schedule` add repository_id int null references project__repository(`id`) on delete set null;
alter table `project__schedule` add last_commit_hash varchar(40);
