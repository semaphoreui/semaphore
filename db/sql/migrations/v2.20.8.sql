create table `project__workflow_revision` (
  `id` integer primary key autoincrement,
  `project_id` int not null,
  `workflow_template_id` int not null,
  `number` int not null,
  `created` datetime not null,
  `created_by_user_id` int null,

  foreign key (`project_id`) references `project`(`id`) on delete cascade,
  foreign key (`workflow_template_id`) references `project__workflow_template`(`id`) on delete cascade,
  foreign key (`created_by_user_id`) references `user`(`id`) on delete set null
);

create unique index `project__workflow_revision__template_number` on `project__workflow_revision`(`workflow_template_id`, `number`);
create index `project__workflow_revision__project_id` on `project__workflow_revision`(`project_id`);

{{if .Sqlite}}
alter table `project__workflow_node` add `revision_id` int null references `project__workflow_revision`(`id`) on delete cascade;
alter table `project__workflow_edge` add `revision_id` int null references `project__workflow_revision`(`id`) on delete cascade;
alter table `project__workflow_run` add `revision_id` int null references `project__workflow_revision`(`id`);
{{else}}
alter table `project__workflow_node` add `revision_id` int null;
alter table `project__workflow_node` add constraint `project__workflow_node__revision_fk` foreign key (`revision_id`) references `project__workflow_revision`(`id`) on delete cascade;
alter table `project__workflow_edge` add `revision_id` int null;
alter table `project__workflow_edge` add constraint `project__workflow_edge__revision_fk` foreign key (`revision_id`) references `project__workflow_revision`(`id`) on delete cascade;
alter table `project__workflow_run` add `revision_id` int null;
alter table `project__workflow_run` add constraint `project__workflow_run__revision_fk` foreign key (`revision_id`) references `project__workflow_revision`(`id`);
{{end}}

create index `project__workflow_node__revision_id` on `project__workflow_node`(`revision_id`);
create index `project__workflow_edge__revision_id` on `project__workflow_edge`(`revision_id`);
create index `project__workflow_run__revision_id` on `project__workflow_run`(`revision_id`);
