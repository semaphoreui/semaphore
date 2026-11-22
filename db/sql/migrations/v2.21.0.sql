-- MCP servers of an inventory of type "mcp": one row per server, deleted with the inventory.
create table `project__inventory_mcp`
(
    `id`           integer primary key autoincrement,
    `inventory_id` int           not null,
    `name`         varchar(64){{if .Mysql}} character set utf8mb4 collate utf8mb4_bin{{end}} not null,
    `url`          varchar(2048) not null,
    `enabled`      boolean       not null default true,

    foreign key (`inventory_id`) references project__inventory (`id`) on delete cascade,
    unique (`inventory_id`, `name`)
);

-- The authorization token of an MCP server is an access key with owner "mcp" pointing at the server;
-- it is deleted with the server. A server has at most one token.
{{if .Sqlite}}
alter table `access_key` add `mcp_id` int null references project__inventory_mcp (`id`) on delete cascade;
{{else}}
alter table `access_key` add `mcp_id` int null;

alter table `access_key` add constraint `access_key__mcp_fk` foreign key (`mcp_id`) references `project__inventory_mcp`(`id`) on delete cascade;
{{end}}

create unique index `access_key__mcp_id` on `access_key` (`mcp_id`);
