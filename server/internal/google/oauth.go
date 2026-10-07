package google

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/lnardon/arete/internal/config"
	"golang.org/x/oauth2"
	googleoauth "golang.org/x/oauth2/google"
)

var scopes = []string{
	"https://www.googleapis.com/auth/calendar",
	"https://www.googleapis.com/auth/userinfo.email",
}

type OAuthService struct {
	cfg *oauth2.Config
}

func NewOAuthService(cfg config.GoogleConfig) *OAuthService {
	return &OAuthService{
		cfg: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes:       scopes,
			Endpoint:     googleoauth.Endpoint,
		},
	}
}

// AuthURL builds the Google consent screen URL. access_type=offline plus
// prompt=consent guarantee a refresh token is issued even if the user has
// authorized this app before — Google only returns one on the very first
// consent otherwise.
func (s *OAuthService) AuthURL(state string) string {
	return s.cfg.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
	)
}

func (s *OAuthService) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return s.cfg.Exchange(ctx, code)
}

func (s *OAuthService) TokenSource(ctx context.Context, tok *oauth2.Token) oauth2.TokenSource {
	return s.cfg.TokenSource(ctx, tok)
}

// Scope returns the space-separated scopes this app requests, stored
// alongside a connected account for reference.
func (s *OAuthService) Scope() string {
	return strings.Join(scopes, " ")
}

// FetchEmail looks up the connected Google account's email address, shown in
// Settings as "Connected as x@gmail.com". A plain HTTP call rather than
// pulling in the full generated oauth2/v2 client for one field.
func (s *OAuthService) FetchEmail(ctx context.Context, ts oauth2.TokenSource) (string, error) {
	client := oauth2.NewClient(ctx, ts)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch userinfo: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("userinfo endpoint returned %d", resp.StatusCode)
	}

	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("decode userinfo: %w", err)
	}
	return body.Email, nil
}

// RevokeToken best-effort revokes a token with Google, called on disconnect.
func RevokeToken(ctx context.Context, token string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://oauth2.googleapis.com/revoke?token="+url.QueryEscape(token), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("revoke returned %d", resp.StatusCode)
	}
	return nil
}
