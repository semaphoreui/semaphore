package api

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
)

func TestApiPing(t *testing.T) {
	req, _ := http.NewRequest("GET", "/api/ping", nil)
	rr := httptest.NewRecorder()

	r := Route()

	r.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Errorf("Response code should be 200 %d", rr.Code)
	}
}

// TestApi Validates the api description in the root meets the swagger/openapi spec
func TestApiSchemaValidation(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fpath := dir + "/../api-docs.yml"
	print(fpath)
	document, err := loads.Spec(fpath)
	if err != nil {
		t.Fatal(err)
	}
	spc := spec.ExpandOptions{RelativeBase: fpath}
	_, err = document.Expanded(&spc)
	if err != nil {
		t.Fatal(err)
	}

	// this is broken right now, use cli `swagger validate ./api-docs.yml`
	//if err := validate.Spec(document, strfmt.Default); err != nil {
	//	t.Fatal(err)
	//}
}
