//go:build !pro

package db

func (t TemplateApp) NeedTaskAlias() bool {
	return false
}
