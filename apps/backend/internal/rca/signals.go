package rca

import (
	"math"
	"sort"
	"strings"
	"time"
)

// logKeywords is checked in order against each log line (case-insensitive
// substring match); the first match wins. This is a fixed, deterministic
// keyword list - not an LLM classification - so the same log lines always
// produce the same LogSignal set.
var logKeywords = []string{
	"oomkilled",
	"out of memory",
	"killed process",
	"panic",
	"segmentation fault",
	"core dumped",
	"fatal",
	"traceback",
	"unhandled exception",
	"exception",
	"connection refused",
	"connection reset",
	"timed out",
	"timeout",
	"crashloopbackoff",
	"unable to connect",
}

// extractLogSignals scans already-fetched log lines for the fixed keyword
// list above and returns only the lines that matched, each paired with the
// keyword that flagged it. Order is preserved from the input.
func extractLogSignals(logs []LogLine) []LogSignal {
	signals := make([]LogSignal, 0)
	for _, log := range logs {
		lower := strings.ToLower(log.Line)
		for _, keyword := range logKeywords {
			if strings.Contains(lower, keyword) {
				signals = append(signals, LogSignal{
					PodName:   log.PodName,
					Container: log.Container,
					Line:      log.Line,
					Keyword:   keyword,
				})
				break
			}
		}
	}
	return signals
}

// metricKey groups metric points into one series to compute a before/after
// signal for.
type metricKey struct {
	metricType string
	metricName string
}

// computeMetricSignals groups metric points by (type, name) and compares
// the average value before splitAt against the average value at/after it.
// A series needs at least one point on each side of splitAt to produce a
// signal - a series with data on only one side cannot honestly support a
// before/after comparison and is omitted rather than guessed at.
func computeMetricSignals(points []MetricPoint, splitAt time.Time) []MetricSignal {
	grouped := make(map[metricKey][]MetricPoint)
	for _, point := range points {
		key := metricKey{metricType: point.MetricType, metricName: point.MetricName}
		grouped[key] = append(grouped[key], point)
	}

	signals := make([]MetricSignal, 0, len(grouped))
	for key, series := range grouped {
		var beforeSum, afterSum float64
		var beforeCount, afterCount int
		unit := ""
		for _, point := range series {
			if unit == "" {
				unit = point.Unit
			}
			if point.Timestamp.Before(splitAt) {
				beforeSum += point.Value
				beforeCount++
			} else {
				afterSum += point.Value
				afterCount++
			}
		}

		if beforeCount == 0 || afterCount == 0 {
			continue
		}

		before := beforeSum / float64(beforeCount)
		after := afterSum / float64(afterCount)

		signals = append(signals, MetricSignal{
			MetricType:    key.metricType,
			MetricName:    key.metricName,
			Unit:          unit,
			Before:        before,
			After:         after,
			ChangePercent: percentChange(before, after),
			Direction:     direction(before, after),
			SampleCount:   beforeCount + afterCount,
		})
	}

	sort.Slice(signals, func(i, j int) bool {
		if signals[i].MetricType != signals[j].MetricType {
			return signals[i].MetricType < signals[j].MetricType
		}
		return signals[i].MetricName < signals[j].MetricName
	})

	return signals
}

func percentChange(before, after float64) float64 {
	if before == 0 {
		if after == 0 {
			return 0
		}
		return 100
	}
	return math.Round(((after-before)/math.Abs(before))*1000) / 10
}

func direction(before, after float64) string {
	const stableThresholdPercent = 10.0
	change := percentChange(before, after)
	switch {
	case change > stableThresholdPercent:
		return "increased"
	case change < -stableThresholdPercent:
		return "decreased"
	default:
		return "stable"
	}
}
