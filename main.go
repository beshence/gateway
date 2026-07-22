package main

import (
	"gateway/internal/api"
	"gateway/internal/api/versioning"
	"gateway/internal/auth"
	"gateway/internal/env"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	env, err := env.Load()
	if err != nil {
		log.Fatal(err)
	}

	jwt := auth.NewJWTManager(
		env.JWTSecret,
		env.JWTTTLSeconds,
	)

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"*"},
		MaxAge:          24 * time.Hour,
	}))

	dependencies := api.NewDependencies(
		jwt,
	)

	versionedEndpoints := versioning.GetVersionedEndpoints(dependencies)

	apiRoute := router.Group("/api")
	versioning.RegisterVersionedRoutes(apiRoute, versionedEndpoints)

	err = router.Run(":27462")
	if err != nil {
		log.Fatal(err)
	}
}
