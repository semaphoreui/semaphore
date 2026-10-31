package output

const (
	ansiReset  = "\033[0m"
	ansiBold   = "\033[1m"
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiCyan   = "\033[36m"
	ansiGray   = "\033[90m"
)

func (p *Printer) colorOn() bool { return p.opts.TTY && p.opts.Color }

func (p *Printer) bold(s string) string {
	if !p.colorOn() {
		return s
	}
	return ansiBold + s + ansiReset
}

// colorCell colours the values people scan for: task and runner statuses and
// booleans. Any other field is returned unchanged.
func (p *Printer) colorCell(field, value string) string {
	if !p.colorOn() {
		return value
	}
	switch field {
	case "status", "tpl_status":
		return StatusColor(value) + value + ansiReset
	case "active", "admin", "alert":
		if value == "true" {
			return ansiGreen + value + ansiReset
		}
		if value == "false" {
			return ansiGray + value + ansiReset
		}
	}
	return value
}

// StatusColor maps a task status to an ANSI colour code (empty when unknown).
func StatusColor(status string) string {
	switch status {
	case "success", "online":
		return ansiGreen
	case "error", "offline":
		return ansiRed
	case "stopped", "stopping", "rejected", "waiting_confirmation":
		return ansiYellow
	case "running", "starting", "confirmed":
		return ansiCyan
	case "waiting":
		return ansiGray
	}
	return ""
}
