package events

import (
	"strconv"
	"time"

	"github.com/sp3640/opspilot/backend/internal/dto"
	corev1 "k8s.io/api/core/v1"
)

type EventMapper struct {
	helpers *ReasonHelper
	now     func() time.Time
}

func NewEventMapper(helper *ReasonHelper) *EventMapper {
	if helper == nil {
		helper = NewReasonHelper()
	}

	return &EventMapper{
		helpers: helper,
		now:     time.Now,
	}
}

func (m *EventMapper) CalculateEventAge(event corev1.Event) string {
	base := event.CreationTimestamp.Time
	if !event.LastTimestamp.IsZero() {
		base = event.LastTimestamp.Time
	} else if !event.EventTime.IsZero() {
		base = event.EventTime.Time
	} else if !event.FirstTimestamp.IsZero() {
		base = event.FirstTimestamp.Time
	}
	if base.IsZero() {
		return ""
	}

	delta := m.now().Sub(base)
	if delta < time.Minute {
		return timeDurationString(int(delta.Seconds()), "s")
	}
	if delta < time.Hour {
		return timeDurationString(int(delta.Minutes()), "m")
	}
	if delta < 24*time.Hour {
		return timeDurationString(int(delta.Hours()), "h")
	}

	return timeDurationString(int(delta.Hours()/24), "d")
}

func (m *EventMapper) MapEvent(event corev1.Event, includeDetails bool) dto.EventDetailResponse {
	component, source := m.helpers.ResolveSource(event)

	response := dto.EventDetailResponse{
		EventResponse: dto.EventResponse{
			Name:           event.Name,
			Namespace:      event.Namespace,
			Reason:         event.Reason,
			Message:        event.Message,
			Type:           m.helpers.NormalizeType(event.Type),
			Count:          event.Count,
			FirstTimestamp: ptrTime(event.FirstTimestamp.Time),
			LastTimestamp:  ptrTime(event.LastTimestamp.Time),
			Age:            m.CalculateEventAge(event),
			InvolvedObject: m.helpers.ResolveInvolvedObject(event),
			Component:      component,
			Source:         source,
		},
	}

	if !includeDetails {
		return response
	}

	response.UID = string(event.UID)
	response.ResourceVersion = event.ResourceVersion
	response.ReportingController = event.ReportingController
	response.ReportingInstance = event.ReportingInstance
	response.Action = event.Action
	response.Labels = event.Labels
	response.Annotations = event.Annotations
	if event.Series != nil {
		response.Series = &dto.EventSeriesResponse{
			Count:            event.Series.Count,
			LastObservedTime: ptrTime(event.Series.LastObservedTime.Time),
		}
	}

	return response
}

func ptrTime(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}

	copied := value
	return &copied
}

func timeDurationString(value int, suffix string) string {
	if value < 0 {
		value = 0
	}

	return strconv.Itoa(value) + suffix
}
