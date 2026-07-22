package versioning

import (
	"gateway/internal/api"
	"gateway/internal/api/endpoints/challenge"
	"gateway/internal/api/endpoints/ek"
	"gateway/internal/api/endpoints/misc"
	"net/http"
)

const (
	HeaderAPIVersion  = "X-Beshence-Gateway-API-Version"
	VersionV1dot0dot0 = "v1.0.0"
	DefaultAPIVersion = VersionV1dot0dot0
)

var supportedVersions = []string{VersionV1dot0dot0 /*, VersionV1dot1*/}

var versionIndex = map[string]int{
	VersionV1dot0dot0: 0,
	// VersionV1dot1dot0: 1,
}

func GetVersionedEndpoints(deps *api.Dependencies) VersionedEndpoints {
	return VersionedEndpoints{
		VersionV1dot0dot0: {
			http.MethodGet: {
				"/ping":                   misc.PingV1(),
				"/bank/:bankId/ek":        ek.GetEKV1(),
				"/bank/:bankId/challenge": challenge.GetChallengeV1(),
			},
			http.MethodPost: {
				"/bank/:bankId/ek":        ek.PostEKV1(),
				"/bank/:bankId/challenge": challenge.PassChallengeV1(deps),
			},
		},
	}
}
