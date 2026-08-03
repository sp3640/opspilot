package resourcesync

import "time"

type SyncResult struct {
	Created  int
	Updated  int
	Deleted  int
	Restored int
	Duration time.Duration
	Errors   []string
}
