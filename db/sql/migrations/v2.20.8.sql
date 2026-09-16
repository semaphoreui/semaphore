alter table `project__proxy` add `requires_proxy_id` int references `project__proxy`(`id`) on delete set null;
