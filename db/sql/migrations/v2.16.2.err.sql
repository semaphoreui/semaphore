alter table project__schedule drop task_params_id;
alter table project__integration drop task_params_id;
alter table project__integration add task_params TEXT;

drop table project__task_params;