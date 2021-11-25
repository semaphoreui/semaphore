-- ALTER TABLE `user__token` CHANGE COLUMN `id` `id` VARCHAR(44) NOT NULL;

alter table user__token rename to user__token_backup;

create table user__token
(
    id varchar(44) not null primary key,
    created datetime not null,
    expired boolean default false not null,
    user_id int not null,

    foreign key (`user_id`) references `user`(`id`) on delete cascade
);

insert into user__token select * from user__token_backup;

drop table user__token_backup;
