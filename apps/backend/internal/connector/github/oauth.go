package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	defaultAuthorizeURL   = "https://github.com/login/oauth/authorize"
	defaultAccessTokenURL = "https://github.com/login/oauth/access_token"
	// oauthScopes is deliberately minimal and read-only in spirit: "repo"
	// is required by GitHub's classic OAuth Apps to read private repository
	// contents/commits/PRs (GitHub has no narrower "read-only private repo"
	// scope for OAuth Apps); "read:user" identifies the connected account.
	// Neither grants OpsPilot any GitHub write capability by itself - this
	// package simply never calls a GitHub write endpoint.
	oauthScopes = "repo read:user"
)

// OAuthConfig holds a GitHub OAuth App's registration details. Empty
// ClientID/ClientSecret means GitHub OAuth is not configured for this
// deployment - callers must check IsConfigured before starting a flow
// rather than attempting one that can only fail.
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string

	// AuthorizeURL/AccessTokenURL are overridable so tests can point the
	// token exchange at an httptest.Server instead of real GitHub.
	AuthorizeURL   string
	AccessTokenURL string

	httpClient *http.Client
}

func NewOAuthConfig(clientID, clientSecret, redirectURL string) *OAuthConfig {
	return &OAuthConfig{
		ClientID:       clientID,
		ClientSecret:   clientSecret,
		RedirectURL:    redirectURL,
		AuthorizeURL:   defaultAuthorizeURL,
		AccessTokenURL: defaultAccessTokenURL,
		httpClient:     &http.Client{Timeout: requestTimeout},
	}
}

func (o *OAuthConfig) IsConfigured() bool {
	return o != nil && strings.TrimSpace(o.ClientID) != "" && strings.TrimSpace(o.ClientSecret) != ""
}

// BuildAuthorizeURL returns the URL the browser should be redirected to,
// embedding the caller-supplied, already-signed state token verbatim -
// this package never signs or interprets the state itself (that's
// GitHubService's concern, reusing the existing JWT signing key rather than
// a new one).
func (o *OAuthConfig) BuildAuthorizeURL(state string) string {
	query := url.Values{
		"client_id":    {o.ClientID},
		"redirect_uri": {o.RedirectURL},
		"scope":        {oauthScopes},
		"state":        {state},
	}
	return o.AuthorizeURL + "?" + query.Encode()
}

type oauthTokenResponse struct {
	AccessToken      string `json:"access_token"`
	Scope            string `json:"scope"`
	TokenType        string `json:"token_type"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// ExchangeCode trades a one-time authorization code for an access token.
// Never logs or wraps the code/token into an error message - a failure
// here reports only GitHub's own error *category* (via ErrUnauthorized/
// ErrUnavailable), never the raw request/response body.
func (o *OAuthConfig) ExchangeCode(ctx context.Context, code string) (string, error) {
	if !o.IsConfigured() {
		return "", fmt.Errorf("github oauth is not configured")
	}

	requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	form := url.Values{
		"client_id":     {o.ClientID},
		"client_secret": {o.ClientSecret},
		"code":          {code},
		"redirect_uri":  {o.RedirectURL},
	}

	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, o.AccessTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("build github oauth exchange request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := o.httpClient
	if client == nil {
		client = &http.Client{Timeout: requestTimeout}
	}

	resp, err := client.Do(req)
	if err != nil {
		if requestCtx.Err() != nil {
			return "", ErrTimeout
		}
		return "", ErrUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return "", ErrUnavailable
	}

	var parsed oauthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", fmt.Errorf("decode github oauth response: %w", err)
	}

	if parsed.Error != "" {
		return "", ErrUnauthorized
	}
	if parsed.AccessToken == "" {
		return "", ErrUnauthorized
	}

	return parsed.AccessToken, nil
}
