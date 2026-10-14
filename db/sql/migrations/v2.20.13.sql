{{if .Mysql}}
create unique index `role__scope_slug` on `role` (`scope`, `slug`);
{{else}}
create unique index `role__scope_slug` on `role` (`scope`, lower(`slug`));
{{end}}
