package routes

import (
	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/apidocs"
	"github.com/AlexGalhardo/vukem/router"
)

// APIReference serves the public v1 OpenAPI document and the Scalar reference (SDD 05 §6.5). Both are public
// by design, so they opt in to production; the app document (/app/v1) is not served. Scalar stays English
// (open issue 05-7).
// https://vukem.dev/docs/rest-apis/openapi "API reference UI"
func APIReference(r *router.Router, app *vukem.App) error {
	vukem.ServeOpenAPIWith(app, r, "/api/v1/openapi.json", vukem.OpenAPIOptions{
		PathPrefix: "/api/v1", InProduction: true,
	})
	apidocs.Serve(app, r, "/docs/reference", apidocs.Config{
		SpecURL: "/api/v1/openapi.json", Title: "HexaScore API", InProduction: true, Name: "apidocs",
	})

	return nil
}
