package scheduler

import (
	"log"
	"strings"

	"github.com/citadel/citadel"
)

type LabelScheduler struct {
}

func (l *LabelScheduler) Schedule(c *citadel.Image, e *citadel.Engine) (bool, error) {
	if len(c.Labels) == 0 || l.contains(e, c.Labels) {
		return true, nil
	}

	return false, nil
}

func (l *LabelScheduler) contains(r *citadel.Engine, constraints map[string]string) bool {
	for k, v := range constraints {
		// Skip "host" constraint - it will be fullfilled by the host scheduler.
		if k == "host" {
			continue
		}
		k, v = strings.ToLower(k), strings.ToLower(v)
		if strings.ToLower(r.Labels[k]) != v {
			log.Printf("Discarding %s: %s != %s", r.ID, r.Labels[k], v)
			return false
		}
	}

	return true
}
