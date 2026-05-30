alter table `runner` add column `is_default` boolean not null default false;

update `runner` set `is_default` = true where tag = '';

create table `runner__tag` (
  `id`        integer primary key autoincrement,
  `runner_id` int          not null,
  `tag`       varchar(200) not null,

  foreign key (`runner_id`) references `runner`(`id`) on delete cascade
);

create unique index `runner__tag_runner_id_tag_idx` on `runner__tag` (`runner_id`, `tag`);
create        index `runner__tag_tag_idx`          on `runner__tag` (`tag`);

insert into `runner__tag` (`runner_id`, `tag`)
  select `id`, `tag` from `runner` where `tag` <> '';

alter table `runner` drop column `tag`;
