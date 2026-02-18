package db

type App struct {
	ID        string `db:"id" json:"id"`
	Title     string `db:"title" json:"title"`
	Icon      string `db:"icon" json:"icon"`
	Color     string `db:"color" json:"color"`
	DarkColor string `db:"dark_color" json:"dark_color"`
	Active    bool   `db:"active" json:"active"`
	Priority  int    `db:"priority" json:"priority"`
}

type AppVersion struct {
	ID       int     `db:"id" json:"id"`
	AppID    string  `db:"app_id" json:"app_id"`
	Name     string  `db:"name" json:"name"`
	Path     string  `db:"path" json:"path"`
	Args     *string `db:"args" json:"args,omitempty"`
	Active   bool    `db:"active" json:"active"`
	Priority int     `db:"priority" json:"priority"`
}
