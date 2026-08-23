package dto

type ApplicationHealthFactorResponse struct {
	Key    string `json:"key"`
	Label  string `json:"label"`
	State  string `json:"state"`
	Score  *int   `json:"score,omitempty"`
	Weight int    `json:"weight"`
	Reason string `json:"reason"`
}

type ApplicationHealthResponse struct {
	ApplicationID string                            `json:"applicationId"`
	Score         *int                              `json:"score,omitempty"`
	State         string                            `json:"state"`
	Factors       []ApplicationHealthFactorResponse `json:"factors"`
}
