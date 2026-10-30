package apiclient

import (
	"fmt"
	"net/url"
	"strconv"
)

// projectPath builds "/project/{id}" + suffix.
func projectPath(projectID int, format string, args ...any) string {
	return "/project/" + strconv.Itoa(projectID) + fmt.Sprintf(format, args...)
}

// withQuery appends url.Values to an API path; an empty set is a no-op.
func withQuery(apiPath string, q url.Values) string {
	if len(q) == 0 {
		return apiPath
	}
	return apiPath + "?" + q.Encode()
}
