drop table `task__alert_send`;

alter table `task` drop column `alert_snapshot`;

alter table `project__schedule` drop column `alert_on_error`;
alter table `project__schedule` drop column `alert_on_success`;
alter table `project__schedule` drop column `alert_mode`;

alter table `project__template` drop column `alert_on_error`;
alter table `project__template` drop column `alert_on_success`;
alter table `project__template` drop column `alert_mode`;

drop table `project__schedule_alert`;
drop table `project__template_alert`;
drop table `project__alert`;
