alter table `runner` add column `tag` varchar(200) not null default '';

update `runner` set `tag` = (
  select rt.`tag` from `runner__tag` rt where rt.`runner_id` = `runner`.`id` limit 1
) where exists (
  select 1 from `runner__tag` rt where rt.`runner_id` = `runner`.`id`
);

drop index if exists `runner__tag_tag_idx`;
drop index if exists `runner__tag_runner_id_tag_idx`;
drop table `runner__tag`;
