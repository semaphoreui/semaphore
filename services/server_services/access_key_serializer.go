package server_services

import (
	"github.com/semaphoreui/semaphore/db"
)

type AccessKeyKeyDeserializer interface {
	DeserializeSecret(key *db.AccessKey) (string, error)
	SerializeSecret(key *db.AccessKey) error
}
