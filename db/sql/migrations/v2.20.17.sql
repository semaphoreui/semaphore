-- A user token is accepted only where its type is: `api` by the REST API, `mcp` by the MCP endpoint.
alter table `user__token` add `type` varchar(16) not null default 'api';
