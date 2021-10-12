alter table project__template add `type` varchar(10) not null default 'task';
alter table `task` add `message` varchar(250);
alter table project__template add start_version varchar(20);
alter table project__template add build_template_id int references project__template(id);
alter table `task` add `version` varchar(20);
alter table `task` add commit_hash varchar(40);
