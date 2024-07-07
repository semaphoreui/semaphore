package util

type AppConfig struct {
	Title   string            `json:"title"`
	Icon    string            `json:"icon"`
	Active  bool              `json:"active"`
	AppPath string            `json:"path"`
	AppArgs map[string]string `json:"args"`
}
