package db

type Alias struct {
	ID        int
	Alias     string
	ProjectID int
}

type Aliasable interface {
	ToAlias() Alias
}
