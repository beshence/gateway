package memory

import (
	"sync"

	"golang.org/x/time/rate"
)

var (
	Banks = map[string]Bank{}
	//Challenges    = map[string]Challenge{}
	BanksApiUrls  = map[string]BankApiUrls{}
	Mutex         sync.Mutex
	Limiters      = map[string]*rate.Limiter{}
	LimitersMutex sync.Mutex
)
