package common

import (
	"sync"

	log "github.com/Sirupsen/logrus"
)

var (
	GlobalClockLock sync.Mutex
	GlobalClock     = 0
)

func UpdateGlobalClock(currTimestamp int, local bool) {
	GlobalClockLock.Lock()
	defer GlobalClockLock.Unlock()

	if local {
		GlobalClock += 1
		return
	}
	if currTimestamp > GlobalClock {
		GlobalClock = currTimestamp + 1
	} else {
		GlobalClock += 1
	}
	log.WithFields(log.Fields{
		"clock": GlobalClock,
	}).Debug("updated the global clock")
}
