alter table `project__workflow_edge` add column `input_mode` varchar(16) not null default 'by_name';
alter table `project__workflow_edge` add column `input_mappings` text null;
create index audit_event_created_idx on audit_event (created);
