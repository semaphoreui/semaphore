-- The type of a user token: `api` is accepted by the REST API, `mcp` by the MCP endpoint and the REST API.
alter table `user__token` add `type` varchar(16) not null default 'api';
