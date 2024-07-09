package db

import (
	"fmt"
	"regexp"
)

type Option struct {
	Key   string `db:"key" json:"key"`
	Value string `db:"value" json:"value"`
}

func ValidateOptionKey(key string) error {
	m, err := regexp.Match(`^[\w.]+$`, []byte(key))
	if err != nil {
		return err
	}

	if !m {
		return fmt.Errorf("invalid key format")
	}

	return nil
}
