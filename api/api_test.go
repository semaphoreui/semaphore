package api

import (
	"testing"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/validate"
	"github.com/go-openapi/strfmt"
	"os"
	"log"
	)

// TestApi Validates the api description in the root meets the swagger/openapi spec
func TestApiSchemaValidation(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fpath := dir+"/../api-docs.yml"
	print(fpath)
	document, err := loads.Spec(fpath)
	if err != nil {
		t.Fatal(err)
	}
	spc := spec.ExpandOptions{RelativeBase: fpath}
	document, err = document.Expanded(&spc)
	if err != nil {
		t.Fatal(err)
	}
	if err := validate.Spec(document, strfmt.Default); err != nil {
		t.Fatal(err)
	}
}
