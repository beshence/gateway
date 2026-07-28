package main

import (
	"gateway/internal/api"
	"gateway/internal/api/versioning"
	"gateway/internal/auth"
	"gateway/internal/environment"
	"gateway/internal/signal"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	env, err := environment.Load()
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"*"},
		MaxAge:          24 * time.Hour,
	}))

	jwt := auth.NewJWTManager(
		env.JWTSecret,
		env.JWTTTLSeconds,
	)

	sig := signal.NewSignalManager()

	dependencies := api.NewDependencies(jwt, sig)

	versionedEndpoints := versioning.GetVersionedEndpoints(dependencies)

	apiRoute := router.Group("/api")
	versioning.RegisterVersionedRoutes(apiRoute, versionedEndpoints)

	err = router.Run(":27462")
	if err != nil {
		log.Fatal(err)
	}
}
