package webservices

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/swaggest/rest/chirouter"
	"github.com/swaggest/rest/openapi"
	"github.com/swaggest/rest/response/gzip"

	"github.com/jamesrr39/taskmaster/dal"
	"github.com/swaggest/swgui/v3cdn"
)

func CreateRouter(taskDAL *dal.TaskDAL, baseDir string) (*chirouter.Wrapper, *openapi.Collector) {

	apiSchema, apiRouter := CreateApiRouter(taskDAL, baseDir)

	rootRouter := chirouter.NewWrapper(chi.NewRouter())

	rootRouter.Use(
		middleware.Recoverer, // Panic recovery.
		gzip.Middleware,      // Response compression with support for direct gzip pass through.
		middleware.DefaultLogger,
	)

	rootRouter.Route("/docs", func(r chi.Router) {
		r.Use(cspHeaderMiddleware)
		r.Method(http.MethodGet, "/openapi.json", apiSchema)
		r.Mount("/", v3cdn.NewHandler(
			apiSchema.Reflector().Spec.Info.Title,
			"/docs/openapi.json",
			"/docs"),
		)
	})

	rootRouter.Mount("/api", apiRouter)

	rootRouter.Mount("/", NewClientHandler())

	return rootRouter, apiSchema
}

func cspHeaderMiddleware(next http.Handler) http.Handler {
	csp := strings.Join([]string{
		"default-src: 'self'",
		// "font-src: 'fonts.googleapis.com'",
		"frame-src: 'none'",
	}, "; ")

	fn := func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Security-Policy", csp)
		next.ServeHTTP(w, r)

	}

	return http.HandlerFunc(fn)

}
