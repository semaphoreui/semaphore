// +build ignore

package main

import (
	"log"
	"os"
	"text/template"
)

var versionTmpl = `package util

//The Semaphore build version
var Version = "{{ .VERSION }}"
`

func main(){

	if len(os.Args) <= 1 {
		log.Fatalln("Must pass in version number")
	}

	data := make(map[string]string)
	data["VERSION"] = os.Args[1]

	tmpl := template.New("version")
	var err error
	if tmpl, err = tmpl.Parse(versionTmpl); err != nil {
		log.Fatalln(err)
	}

	f, err := os.Create("util/version.go")
	if err != nil {
		log.Fatalln(err)
	}
	defer func(r *os.File) {
		err = r.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}(f)

	tmpl.Execute(f, data)
}