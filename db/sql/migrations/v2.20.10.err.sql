{{if .Mysql}}
drop index audit_event_created_idx on audit_event;
{{else}}
drop index if exists audit_event_created_idx;
{{end}}
