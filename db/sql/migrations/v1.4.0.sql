CREATE TABLE `event` (
  `project_id` int DEFAULT NULL,
  `object_id` int DEFAULT NULL,
  `object_type` varchar(20) DEFAULT '',
  `description` text,
  `created` datetime NOT NULL
);

CREATE INDEX `project_id` ON `event`(`project_id`);

CREATE INDEX `object_id` ON `event`(`object_id`);

CREATE INDEX `created` ON `event`(`created`);

alter table `task` add `created` datetime null;
alter table `task` add `start` datetime null;
alter table `task` add `end` datetime null;
