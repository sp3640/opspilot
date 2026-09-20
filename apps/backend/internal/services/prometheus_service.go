package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/sp3640/opspilot/backend/internal/connector"
	"github.com/sp3640/opspilot/backend/internal/dto"
)

const (
	prometheusDefaultTimeout = 10 * time.Second
)

var errPrometheusNoResponse = errors.New("prometheus returned no data")

type PrometheusService struct {
	client *http.Client
}

func NewPrometheusService() *PrometheusService {
	return &PrometheusService{client: &http.Client{Timeout: prometheusDefaultTimeout}}
}

func (s *PrometheusService) WithClient(client *http.Client) *PrometheusService {
	if client != nil {
		s.client = client
	}
	return s
}

func (s *PrometheusService) QueryInstant(ctx context.Context, cfg connector.Config, query string) (*dto.PrometheusInstantQueryResult, error) {
	baseURL, err := extractAPIURL(cfg)
	if err != nil {
		return nil, err
	}
	body, err := s.executeQuery(ctx, baseURL, extractBearerToken(cfg), query, "query")
	if err != nil {
		return nil, err
	}
	var payload struct {
		Status string `json:"status"`
		Error  string `json:"error"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Metric map[string]string `json:"metric"`
				Value  []any             `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode Prometheus instant query response: %w", err)
	}
	if payload.Status == "error" {
		return nil, fmt.Errorf("prometheus query error: %s", payload.Error)
	}
	if payload.Data.ResultType == "" && len(payload.Data.Result) == 0 {
		return &dto.PrometheusInstantQueryResult{Query: query, Samples: nil}, nil
	}
	result := &dto.PrometheusInstantQueryResult{Query: query, Samples: make([]dto.PrometheusSample, 0, len(payload.Data.Result))}
	for _, item := range payload.Data.Result {
		if len(item.Value) != 2 {
			continue
		}
		timestamp, err := strconv.ParseFloat(fmt.Sprintf("%v", item.Value[0]), 64)
		if err != nil {
			continue
		}
		value, err := strconv.ParseFloat(fmt.Sprintf("%v", item.Value[1]), 64)
		if err != nil {
			continue
		}
		result.Samples = append(result.Samples, dto.PrometheusSample{
			Timestamp: time.Unix(int64(timestamp), 0).UTC(),
			Value:     value,
			Labels:    item.Metric,
		})
	}
	return result, nil
}

func (s *PrometheusService) QueryRange(ctx context.Context, cfg connector.Config, query string, start, end time.Time, step string) (*dto.PrometheusRangeQueryResult, error) {
	baseURL, err := extractAPIURL(cfg)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Set("query", query)
	params.Set("start", strconv.FormatFloat(float64(start.Unix()), 'f', -1, 64))
	params.Set("end", strconv.FormatFloat(float64(end.Unix()), 'f', -1, 64))
	params.Set("step", step)
	endpoint, err := buildPrometheusURL(baseURL, "query_range", params)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build Prometheus range query request: %w", err)
	}
	if token := extractBearerToken(cfg); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call Prometheus API: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("prometheus api returned status %d", resp.StatusCode)
	}
	body, err := readAllBody(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read Prometheus range query response: %w", err)
	}
	var payload struct {
		Status string `json:"status"`
		Error  string `json:"error"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Metric map[string]string `json:"metric"`
				Values [][]any           `json:"values"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode Prometheus range query response: %w", err)
	}
	if payload.Status == "error" {
		return nil, fmt.Errorf("prometheus query error: %s", payload.Error)
	}
	result := &dto.PrometheusRangeQueryResult{Query: query, Series: make([]dto.PrometheusSeries, 0, len(payload.Data.Result))}
	for _, item := range payload.Data.Result {
		series := dto.PrometheusSeries{Metric: item.Metric, Values: make([]dto.PrometheusSample, 0, len(item.Values))}
		for _, point := range item.Values {
			if len(point) != 2 {
				continue
			}
			timestamp, ok := parsePrometheusNumber(point[0])
			if !ok {
				continue
			}
			value, ok := parsePrometheusNumber(point[1])
			if !ok {
				continue
			}
			series.Values = append(series.Values, dto.PrometheusSample{
				Timestamp: time.Unix(int64(timestamp), 0).UTC(),
				Value:     value,
				Labels:    item.Metric,
			})
		}
		result.Series = append(result.Series, series)
	}
	return result, nil
}

func (s *PrometheusService) executeQuery(ctx context.Context, baseURL, bearerToken, query, mode string) ([]byte, error) {
	params := url.Values{}
	params.Set("query", query)
	endpoint, err := buildPrometheusURL(baseURL, mode, params)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build Prometheus query request: %w", err)
	}
	if bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call Prometheus API: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("prometheus api returned status %d", resp.StatusCode)
	}
	body, err := readAllBody(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read Prometheus query response: %w", err)
	}
	if len(body) == 0 {
		return nil, errPrometheusNoResponse
	}
	return body, nil
}

func buildPrometheusURL(baseURL, mode string, params url.Values) (string, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid Prometheus API URL: %w", err)
	}
	queryURL := *parsed
	queryURL.RawQuery = params.Encode()
	path := strings.TrimRight(queryURL.Path, "/") + "/api/v1/" + strings.TrimLeft(mode, "/")
	queryURL.Path = path
	return queryURL.String(), nil
}

func readAllBody(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}

func extractAPIURL(cfg connector.Config) (string, error) {
	for _, key := range []string{"api_url", "apiUrl", "url", "prometheus_url"} {
		if value := strings.TrimSpace(cfg.Metadata[key]); value != "" {
			return validateAPIURL(value)
		}
	}
	for _, key := range []string{"api_url", "apiUrl", "url", "prometheus_url"} {
		if value := strings.TrimSpace(cfg.Credentials[key]); value != "" {
			return validateAPIURL(value)
		}
	}
	return "", errors.New("prometheus api url is required")
}

func extractBearerToken(cfg connector.Config) string {
	for _, source := range []map[string]string{cfg.Credentials, cfg.Metadata} {
		for _, key := range []string{"bearer_token", "bearerToken", "token", "access_token", "authorization_token"} {
			if value := strings.TrimSpace(source[key]); value != "" {
				return value
			}
		}
	}
	return ""
}

func validateAPIURL(value string) (string, error) {
	parsed, err := url.Parse(value)
	if err != nil || parsed == nil {
		return "", errors.New("prometheus api url is invalid")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errors.New("prometheus api url is invalid")
	}
	if parsed.Host == "" {
		return "", errors.New("prometheus api url is invalid")
	}
	return strings.TrimRight(parsed.String(), "/"), nil
}

func parsePrometheusNumber(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case json.Number:
		parsed, err := v.Float64()
		return parsed, err == nil
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}
