package db

type Runner struct {
	ID    int    `db:"id" json:"-"`
	token string `db:"token" json:"token"`
}
