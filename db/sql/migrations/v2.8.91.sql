ALTER TABLE project__user ADD `role` varchar(50) NOT NULL DEFAULT 'task_runner';

UPDATE project__user SET `role` = 'owner' WHERE `admin`;

ALTER TABLE project__user DROP COLUMN `admin`;