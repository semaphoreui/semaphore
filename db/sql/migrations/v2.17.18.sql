{{if not .Sqlite}}
alter table `user__token` add `name` varchar(255) not null default '';
{{end}}
