# Pull Requests

When creating a pull-request you should:

- __Open an issue first:__ Confirm that the change or feature will be accepted
- __gofmt and vet the code:__ Use  `gofmt`, `golint`, `govet` and `goimports` to clean up your code.
- __Update api documentation:__ If your pull-request adding/modifying an API request, make sure you update the swagger documentation (`swagger.yml`)

# Installing dependencies

```
go get -u ./...
go get -u github.com/jteeuwen/go-bindata/...
go get github.com/mitchellh/gox
npm install -g nodemon less jade
```

# Running in development

```
# edit config.json file
./make.sh watch
```