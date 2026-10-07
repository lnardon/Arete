package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lnardon/arete/internal/config"
)

const CookieName = "arete_token"

type Claims struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// StateClaims signs a short-lived, single-purpose token used to carry a user
// ID through a redirect that won't have the session cookie attached (e.g.
// Google's OAuth callback, a cross-site top-level navigation that our
// SameSite=Strict cookie is not sent on). Purpose guards against a state
// token ever being confused with — or accepted in place of — a real session
// token, even though ValidateClaims/ValidateStateToken already parse into
// distinct types.
type StateClaims struct {
	UserID  string `json:"userId"`
	Purpose string `json:"purpose"`
	jwt.RegisteredClaims
}

const googleOAuthStatePurpose = "google_oauth_state"
const stateTokenTTL = 10 * time.Minute

type Service struct {
	cfg config.JWTConfig
	secure bool
}

func NewService(cfg config.JWTConfig, cookieSecure bool) *Service {
	return &Service{cfg: cfg, secure: cookieSecure}
}

func (s *Service) GenerateToken(userID, username string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.cfg.ExpiresIn) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(s.cfg.SecretKey))
}

func (s *Service) ValidateClaims(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.SecretKey), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// GenerateStateToken signs a short-lived token carrying userID, for use as
// the OAuth "state" parameter — verified at the callback in place of the
// session cookie (see StateClaims).
func (s *Service) GenerateStateToken(userID string) (string, error) {
	claims := StateClaims{
		UserID:  userID,
		Purpose: googleOAuthStatePurpose,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(stateTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.SecretKey))
}

// ValidateStateToken verifies a token minted by GenerateStateToken and
// returns the user ID it carries.
func (s *Service) ValidateStateToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &StateClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.SecretKey), nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*StateClaims)
	if !ok || !token.Valid || claims.Purpose != googleOAuthStatePurpose {
		return "", fmt.Errorf("invalid state token")
	}
	return claims.UserID, nil
}

func (s *Service) SetCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   s.cfg.ExpiresIn * 3600,
		HttpOnly: true,
		Secure:   s.secure,
		SameSite: http.SameSiteStrictMode,
	})
}

func (s *Service) ClearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.secure,
		SameSite: http.SameSiteStrictMode,
	})
}
