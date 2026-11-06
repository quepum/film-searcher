package main

import (
	_ "film-searcher/docs"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func swaggerHandler() http.Handler {
	return httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	)
}
