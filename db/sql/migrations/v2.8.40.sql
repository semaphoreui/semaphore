alter table `project` change `alert_chat` `alert_chat` varchar(30);

alter table `project__template` change `alias` `name` varchar(100) not null;

alter table `project__inventory` drop column `removed`;

alter table `project__environment` drop column `removed`;

alter table `access_key` drop column `removed`;

alter table `project__repository` drop column `removed`;
