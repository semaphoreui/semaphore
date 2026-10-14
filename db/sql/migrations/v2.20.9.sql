create table `audit_event` (
  `seq` bigint not null primary key,
  `event_id` varchar(36) not null,
  `created` {{if .Mysql}}datetime(6){{else}}datetime{{end}} not null,
  `schema_version` varchar(8) not null,
  `category` varchar(16) not null,
  `event_code` varchar(32) not null,
  `type` varchar(16) not null,
  `action` varchar(64) not null,
  `outcome` varchar(16) not null,
  `reason` varchar(64) not null,
  `actor_type` varchar(16) not null,
  `actor_id` varchar(64) not null,
  `actor_name` varchar(255) not null,
  `actor_auth` varchar(16) not null,
  `actor_token_fingerprint` varchar(16) not null,
  `source_ip` varchar(64) not null,
  `user_agent` varchar(1024) not null,
  `target_type` varchar(64) not null,
  `target_id` varchar(255) not null,
  `target_name` varchar(255) not null,
  `project_id` int null,
  `request_id` varchar(36) not null,
  `instance_id` varchar(255) not null,
  `node_id` varchar(255) not null,
  `metadata` text not null
){{if .Mysql}} default charset=utf8mb4{{end}};

create table `audit_event_sequence` (
  `id` int not null primary key,
  `last_seq` bigint not null
){{if .Mysql}} default charset=utf8mb4{{end}};

insert into `audit_event_sequence` (`id`, `last_seq`) values (1, 0);

create table `audit_export_state` (
  `destination_id` varchar(255) not null primary key,
  `cursor_seq` bigint not null
){{if .Mysql}} default charset=utf8mb4{{end}};
