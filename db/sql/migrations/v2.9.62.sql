alter table project add `type` varchar(20) default '';

alter table task add `inventory_id` int null references project__inventory(`id`) on delete set null;

alter table project__inventory add `holder_id` int null references project__template(`id`) on delete set null;

create table `option` (
    `key` varchar(255) primary key not null,
    `value` varchar(255) not null
);
