update `task` set project_id = (select project_id from project__template where project__template.id = `task`.template_id) where project_id is null;
