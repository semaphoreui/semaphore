package db

func (t TemplateApp) NeedTaskAlias() bool {
	return t.IsTerraform()
}
