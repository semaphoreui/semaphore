create table project__invite
(
    id           integer primary key autoincrement,
    project_id   int not null,
    inviter_id   int not null,
    invitee_id   int not null,
    message      varchar(250),

    foreign key (`project_id`) references project (`id`) on delete cascade,
    foreign key (`inviter_id`) references user (`id`) on delete cascade,
    foreign key (`invitee_id`) references user (`id`) on delete cascade
);
