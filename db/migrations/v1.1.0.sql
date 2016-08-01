/* update precision of output logs - also prevents confusion when sorting the table to guarantee it is in the same order */
alter table task__output change `time` `time` datetime not null;
