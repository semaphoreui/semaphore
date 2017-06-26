ALTER TABLE access_key ADD COLUMN `owner` int(11) DEFAULT 0;
UPDATE access_key SET owner=0;

ALTER TABLE user ADD COLUMN `extra_vars` TEXT DEFAULT "";
UPDATE user SET extra_vars = "", vault="";

ALTER TABLE user ADD COLUMN `vault` TEXT DEFAULT "";
UPDATE user SET `vault` = "";

ALTER TABLE project__template ADD COLUMN `user_vault` TINYINT(1) NOT NULL DEFAULT '0';

ALTER TABLE project__template ADD COLUMN `user_vars` TINYINT(1) NOT NULL DEFAULT '0';

ALTER TABLE project__template ADD COLUMN `user_key` TINYINT(1) NOT NULL DEFAULT '0';

UPDATE project__template SET user_vault=0, user_vars=0, user_key=0;