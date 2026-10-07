create index audit_event_actor_idx on audit_event (actor_type, actor_id, seq);
create index audit_event_project_idx on audit_event (project_id, seq);
create index audit_event_kind_idx on audit_event (category, event_code, action, seq);
create index audit_event_outcome_idx on audit_event (outcome, seq);
create index audit_event_ip_idx on audit_event (source_ip, seq);
create index audit_event_target_idx on audit_event (target_type, target_id, seq);
