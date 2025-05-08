update `option` set `value` = '' where `value` is null;
alter table `option` change `value` `value` varchar(1000) not null;
