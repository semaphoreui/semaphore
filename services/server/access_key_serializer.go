package server

import (
	"github.com/semaphoreui/semaphore/db"
)

type AccessKeyDeserializer interface {
	DeserializeSecret(key *db.AccessKey) (string, error)
	SerializeSecret(key *db.AccessKey) error
	DeleteSecret(key *db.AccessKey) error
}
