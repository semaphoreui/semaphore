{{if .Sqlite}}
{{else}}
alter table `project__terraform_inventory_state` change `state` `state` longtext not null;
{{end}}