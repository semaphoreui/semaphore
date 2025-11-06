package helpers

import (
	"github.com/semaphoreui/semaphore/db"
	"net/url"
	"slices"
	"strconv"
)

func QueryParamsForProps(url *url.URL, props db.ObjectProps) (params db.RetrieveQueryParams) {
	sortBy := ""

	if url.Query().Get("sort") != "" {
		i := slices.Index(props.SortableColumns, url.Query().Get("sort"))
		if i != -1 {
			sortBy = props.SortableColumns[i]
		}
	}

	params = db.RetrieveQueryParams{
		SortBy:       sortBy,
		SortInverted: url.Query().Get("order") == "desc",
	}

	return
}

func QueryParams(url *url.URL) db.RetrieveQueryParams {
	return db.RetrieveQueryParams{
		SortBy:       url.Query().Get("sort"),
		SortInverted: url.Query().Get("order") == "desc",
	}
}

func QueryParamsWithOwner(url *url.URL, props db.ObjectProps) db.RetrieveQueryParams {
	res := QueryParamsForProps(url, props)

	hasOwnerFilter := false

	for _, ownership := range props.Ownerships {
		s := url.Query().Get(ownership.ReferringColumnSuffix)
		if s == "" {
			continue
		}

		id, err2 := strconv.Atoi(s)
		if err2 != nil {
			continue
		}

		res.Ownership.SetOwnerID(*ownership, id)
		hasOwnerFilter = true
	}

	if !hasOwnerFilter {
		res.Ownership.WithoutOwnerOnly = true
	}

	return res
}
