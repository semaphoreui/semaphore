alter table session rename to session_backup;

create table session
(
    id integer primary key autoincrement,
    user_id int not null,
    created datetime not null,
    last_active datetime not null,
    ip varchar(39) default '' not null,
    user_agent text not null,
    expired boolean default false not null
);

insert into session select * from session_backup;

drop table session_backup;

create index expired
    on session (expired);

create index user_id
    on session (user_id);
