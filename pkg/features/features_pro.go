//go:build pro

package features

import (
	pro "github.com/semaphoreui/semaphore/pro/pkg/features"
)

func GetFeatures() map[string]bool {
	return pro.GetFeatures()
}
