package util

type App struct {
	Active    bool     `json:"active"`
	Order     int      `json:"order"`
	Title     string   `json:"title"`
	Icon      string   `json:"icon"`
	Color     string   `json:"color"`
	DarkColor string   `json:"dark_color"`
	AppPath   string   `json:"path"`
	AppArgs   []string `json:"args"`
}
