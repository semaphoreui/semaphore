package apiclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// readFile is a seam for tests; it is os.ReadFile in production.
var readFile = os.ReadFile

const masked = "***"

// maskedHeaders never reach the debug trace in clear text.
var maskedHeaders = map[string]bool{
	"Authorization": true,
	"Cookie":        true,
	"Set-Cookie":    true,
}

// secretJSONKeys are masked when a request or response body is traced.
var secretJSONKeys = regexp.MustCompile(`(?i)"(password|passcode|token|secret|registration_token|login_password|ssh|plain)"\s*:\s*("(?:[^"\\]|\\.)*"|\{[^{}]*\})`)

func (c *Client) trace(format string, args ...any) {
	if c.debug == nil {
		return
	}
	_, _ = fmt.Fprintf(c.debug, format+"\n", args...)
}

func (c *Client) traceHeaders(prefix string, h http.Header) {
	if c.debug == nil {
		return
	}
	for name, values := range h {
		for _, v := range values {
			if maskedHeaders[http.CanonicalHeaderKey(name)] {
				v = masked
			}
			_, _ = fmt.Fprintf(c.debug, "%s%s: %s\n", prefix, name, v)
		}
	}
}

// redactJSON masks well-known secret fields of a JSON document for the debug
// trace. It works on the text so that a body that is not an object (an array,
// plain text) passes through unchanged.
func redactJSON(body []byte) string {
	if !json.Valid(body) {
		return strings.TrimSpace(string(body))
	}
	return secretJSONKeys.ReplaceAllString(string(body), `"$1":"`+masked+`"`)
}
