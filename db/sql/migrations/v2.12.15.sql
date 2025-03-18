alter table `project__schedule` change `last_commit_hash` `last_commit_hash` varchar(64);
alter table `task` change `commit_hash` `commit_hash` varchar(64);
