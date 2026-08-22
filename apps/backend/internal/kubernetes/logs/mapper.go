package logs

import (
	"time"

	"github.com/sp3640/opspilot/backend/internal/dto"
)

type LogMapper struct {
	now func() time.Time
}

func NewLogMapper() *LogMapper {
	return &LogMapper{now: time.Now}
}

func (m *LogMapper) MapPodLogs(pod string, container string, namespace string, logOutput string, previous bool) *dto.PodLogResponse {
	if m == nil {
		m = NewLogMapper()
	}

	return &dto.PodLogResponse{
		Pod:         pod,
		Container:   container,
		Namespace:   namespace,
		Log:         logOutput,
		Previous:    previous,
		RetrievedAt: m.now().UTC(),
	}
}
