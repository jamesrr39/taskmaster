package webservices

import (
	"github.com/go-chi/chi/v5"
	"github.com/jamesrr39/go-openapix"
	"github.com/jamesrr39/taskmaster/dal"
	"github.com/swaggest/openapi-go/openapi3"
	"github.com/swaggest/rest"
	"github.com/swaggest/rest/chirouter"
	"github.com/swaggest/rest/jsonschema"
	"github.com/swaggest/rest/nethttp"
	"github.com/swaggest/rest/openapi"
	"github.com/swaggest/rest/request"
	"github.com/swaggest/rest/response"
)

func CreateApiRouter(taskDAL *dal.TaskDAL, baseDir string) (*openapi.Collector, *chirouter.Wrapper) {
	apiSchema := &openapi.Collector{}
	apiSchema.Reflector().SpecEns().Info.Title = "Taskmaster"
	apiSchema.Reflector().SpecEns().Info.WithDescription("REST API definitions for Taskmaster")

	serverDesc := "API server"

	apiSchema.Reflector().SpecEns().Info.Version = "0"

	apiSchema.Reflector().Spec.Servers = append(apiSchema.Reflector().Spec.Servers, openapi3.Server{
		URL:         "/api",
		Description: &serverDesc,
	})

	// Setup request decoder and validator.
	validatorFactory := jsonschema.NewFactory(apiSchema, apiSchema)
	decoderFactory := request.NewDecoderFactory()
	decoderFactory.ApplyDefaults = true
	decoderFactory.SetDecoderFunc(rest.ParamInPath, chirouter.PathToURLValues)

	apiRouter := chirouter.NewWrapper(chi.NewRouter())
	apiRouter.Use(
		nethttp.OpenAPIMiddleware(apiSchema),          // Documentation collector.
		request.DecoderMiddleware(decoderFactory),     // Request decoder setup.
		request.ValidatorMiddleware(validatorFactory), // Request validator setup.
		response.EncoderMiddleware,                    // Response encoder setup.
	)

	apiRouter.Route("/v1", func(r chi.Router) {
		openapix.Get(r, "/tasks", GetAllTasks(taskDAL, baseDir))
	})

	// check array types are marked as non-null; i.e. no items will return "[]" instead of "null"
	openapix.MustCheckNonNullArrays(apiSchema.Reflector().Spec.Components.Schemas.MapOfSchemaOrRefValues)
	openapix.MustNotHaveDuplicateOperationIDOrUnknownSecurity(apiSchema.Reflector().Spec)

	return apiSchema, apiRouter
}
