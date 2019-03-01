# GW2 Verify for Go

## Generate API doc

`raml2html api/api.raml > api.html`

## Generate API code

`go-raml server --ramlfile api/api.raml --dir internal/api --package api --import-path github.com/vennekilde/gw2verify/internal/api --no-apidocs`