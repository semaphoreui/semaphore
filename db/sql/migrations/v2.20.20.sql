-- A key whose deletion is in progress: the Unix time in milliseconds until which the claim holds. The secret
-- of a key leaves its storage before the row does; a variable group must not link the key in between.
alter table `access_key` add `delete_claim` bigint null;
