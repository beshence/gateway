package versioning

import (
	"gateway/internal/api"
	"gateway/internal/api/endpoints/misc"
	"gateway/internal/api/endpoints/pk"
	"gateway/internal/api/endpoints/urls"
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
				"/ping":                 misc.PingV1(),
				"/bank/:bankId/pk":      pk.GetPublicKeyV1(),
				"/bank/:bankId/urls":    urls.GetApiUrlsV1(),
				"/bank/:bankId/urls/ps": urls.GetApiUrlsPayloadSignatureV1(),
				//"/bank/:bankId/challenge": challenge.GetChallengeV1(),
				//"/bank/:bankId/ws":        ws.WSV1(deps),
			},
			http.MethodPost: {
				"/bank/:bankId/pk":   pk.PostPublicKeyV1(),
				"/bank/:bankId/urls": urls.PostApiUrlsV1(),
				//"/bank/:bankId/challenge": challenge.PassChallengeV1(deps),
			},
		},
	}
}
