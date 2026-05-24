package features

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
)

func GetFeatures(user *db.User, plan string) pro_interfaces.Features {
	return pro_interfaces.Features{}
}
