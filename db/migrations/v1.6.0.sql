# add deleted column

alter table project__environment add `removed` tinyint(1) default 0 comment 'marks as deleted';
alter table project__inventory add `removed` tinyint(1) default 0 comment 'marks as deleted';
alter table project__repository add `removed` tinyint(1) default 0 comment 'marks as deleted';
alter table access_key add `removed` tinyint(1) default 0 comment 'marks as deleted';