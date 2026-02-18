create table `app` (
    `id`        varchar(100) primary key not null,
    `title`     varchar(100) not null default '',
    `icon`      varchar(100) not null default '',
    `color`     varchar(100) not null default '',
    `dark_color` varchar(100) not null default '',
    `active`    boolean not null default true,
    `priority`  int not null default 0
);

create table `app__version` (
    `id`       integer primary key autoincrement,
    `app_id`   varchar(100) not null,
    `name`     varchar(100) not null default '',
    `path`     varchar(255) not null default '',
    `args`     varchar(1000),
    `active`   boolean not null default true,
    `priority` int not null default 0,
    foreign key (`app_id`) references `app` (`id`) on delete cascade
);
