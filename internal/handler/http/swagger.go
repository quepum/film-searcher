package http

import (
	_ "film-searcher/docs"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func SwaggerHandler() http.Handler {
	return httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	)
}
