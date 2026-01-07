{{if .Sqlite}}
{{else}}
alter table `project__terraform_inventory_state` change `state` `state` text not null;
{{end}}