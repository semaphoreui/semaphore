ALTER TABLE `user` ADD PRIMARY KEY(id)

ALTER TABLE `project` ADD PRIMARY KEY(id)

ALTER TABLE `project__user` ADD PRIMARY KEY(project_id)

ALTER TABLE `access_key` ADD PRIMARY KEY(id)

ALTER TABLE `project__repository` ADD PRIMARY KEY(id)

ALTER TABLE `project__inventory` ADD PRIMARY KEY(id)

ALTER TABLE `project__environment` ADD PRIMARY KEY(id)

ALTER TABLE `project__template` ADD PRIMARY KEY(id)

ALTER TABLE `project__template_schedule` ADD PRIMARY KEY(template_id)

ALTER TABLE `task` ADD PRIMARY KEY(id)

ALTER TABLE `task__output` ADD PRIMARY KEY(task_id)

ALTER TABLE `user__token` ADD PRIMARY KEY(id)

ALTER TABLE `event` ADD PRIMARY KEY(project_id)

ALTER TABLE `session` ADD PRIMARY KEY(id)
