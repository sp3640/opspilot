package discovery

import (
	"time"

	"github.com/sp3640/opspilot/backend/internal/models"
)

type DiscoveryResult struct {
	Resources []models.Resource
	Warnings  []string
	Errors    []string
	Duration  time.Duration
}
