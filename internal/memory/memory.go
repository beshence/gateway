package memory

import "sync"

var (
	Banks      = map[string]Bank{}
	Challenges = map[string]Challenge{}
	APIURLs    = map[string]ApiUrls{}
	Mutex      sync.Mutex
)
