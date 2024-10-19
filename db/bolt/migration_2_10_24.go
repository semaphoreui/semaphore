package bolt

import "fmt"

type migration_2_10_24 struct {
	migration
}

func (d migration_2_10_24) Apply() (err error) {
	projectIDs, err := d.getProjectIDs()
	if err != nil {
		return err
	}

	for _, projectID := range projectIDs {
		templates, err := d.getObjects(projectID, "template")
		if err != nil {
			return err
		}

		var templateVaultID int = 1
		for templateID, template := range templates {
			if template["vault_key_id"] != nil {
				templateVault := map[string]interface{}{
					"id":           templateVaultID,
					"project_id":   template["project_id"],
					"template_id":  template["id"],
					"vault_key_id": template["vault_key_id"],
					"name":         nil,
				}
				err = d.setObject(projectID, "template_vault", fmt.Sprintf("%010d", templateVaultID), templateVault)
				if err != nil {
					return err
				}
				templateVaultID++
			}
			delete(template, "vault_key_id")
			err = d.setObject(projectID, "template", templateID, template)
			if err != nil {
				return err
			}
		}
	}

	return
}
