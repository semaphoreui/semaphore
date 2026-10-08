{{if .Mysql}}
drop index audit_event_actor_idx on audit_event;
drop index audit_event_project_idx on audit_event;
drop index audit_event_kind_idx on audit_event;
drop index audit_event_outcome_idx on audit_event;
drop index audit_event_ip_idx on audit_event;
drop index audit_event_target_idx on audit_event;
drop index audit_event_created_idx on audit_event;
{{else}}
drop index if exists audit_event_actor_idx;
drop index if exists audit_event_project_idx;
drop index if exists audit_event_kind_idx;
drop index if exists audit_event_outcome_idx;
drop index if exists audit_event_ip_idx;
drop index if exists audit_event_target_idx;
drop index if exists audit_event_created_idx;
{{end}}
create index audit_event_created_idx on audit_event (created);
