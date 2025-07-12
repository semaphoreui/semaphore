//go:build !pro

package features

func GetFeatures() map[string]bool {
	return map[string]bool{
		"project_runners":   false,
		"terraform_backend": false,
		"task_summary":      false,
		"secret_storages":   false,
	}
}
