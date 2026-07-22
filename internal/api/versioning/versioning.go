package versioning

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type EndpointsHandlers map[string]gin.HandlerFunc
type MethodsEndpoints map[string]EndpointsHandlers
type VersionedEndpoints map[string]MethodsEndpoints

func RegisterVersionedRoutes(g *gin.RouterGroup, versionedEndpoints VersionedEndpoints) {
	seen := make(map[string]map[string]struct{})

	for _, methods := range versionedEndpoints {
		for method, endpoints := range methods {
			if seen[method] == nil {
				seen[method] = make(map[string]struct{})
			}

			for endpoint := range endpoints {
				if _, ok := seen[method][endpoint]; ok {
					continue
				}

				seen[method][endpoint] = struct{}{}
				VersionEndpoint(g, versionedEndpoints, method, endpoint)
			}
		}
	}
}

func VersionEndpoint(g *gin.RouterGroup, versionedEndpoints VersionedEndpoints, method string, endpoint string, handlers ...gin.HandlerFunc) {
	chain := make([]gin.HandlerFunc, 0, len(handlers)+1)
	chain = append(chain, handlers...)
	chain = append(chain, func(c *gin.Context) {
		version := c.GetHeader(HeaderAPIVersion)
		if version == "" {
			version = DefaultAPIVersion
		}

		requestedIndex, ok := versionIndex[version]
		if !ok {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"err":     "UNKNOWN",
				"message": fmt.Sprintf("unsupported API version: %s", version),
			})
			return
		}

		for i := requestedIndex; i >= 0; i-- {
			candidateVersion := supportedVersions[i]
			if methods, ok := versionedEndpoints[candidateVersion]; ok {
				if endpoints, ok := methods[method]; ok {
					if handler, ok := endpoints[endpoint]; ok {
						handler(c)
						return
					}
				}
			}
		}

		for i := requestedIndex; i >= 0; i-- {
			candidateVersion := supportedVersions[i]
			if methods, ok := versionedEndpoints[candidateVersion]; ok {
				for _, endpoints := range methods {
					if _, exists := endpoints[endpoint]; exists {
						c.AbortWithStatusJSON(http.StatusMethodNotAllowed, gin.H{
							"err":     "UNKNOWN",
							"message": fmt.Sprintf("method %s is not available for endpoint %s in version %s or earlier", method, endpoint, version),
						})
						return
					}
				}
			}
		}

		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"err":     "UNKNOWN",
			"message": fmt.Sprintf("endpoint %s is not available for version %s or earlier", endpoint, version),
		})
	})
	g.Handle(method, endpoint, chain...)
}
