package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/lnardon/arete/internal/auth"
	"github.com/lnardon/arete/internal/google"
	"github.com/lnardon/arete/internal/repository"
)

type GoogleHandler struct {
	accounts  *repository.GoogleAccountRepository
	oauth     *google.OAuthService
	sync      *google.SyncEngine
	authSvc   *auth.Service
	appDomain string
}

func NewGoogleHandler(accounts *repository.GoogleAccountRepository, oauth *google.OAuthService, syncEngine *google.SyncEngine, authSvc *auth.Service, appDomain string) *GoogleHandler {
	return &GoogleHandler{accounts: accounts, oauth: oauth, sync: syncEngine, authSvc: authSvc, appDomain: appDomain}
}

func (h *GoogleHandler) AuthURL(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	state, err := h.authSvc.GenerateStateToken(authUser.ID)
	if err != nil {
		http.Error(w, "failed to start google connection", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"authUrl": h.oauth.AuthURL(state)})
}

// Callback is public: Google's redirect back here is a cross-site top-level
// navigation, so our SameSite=Strict session cookie is not attached. Identity
// instead comes from the signed state token minted by AuthURL.
func (h *GoogleHandler) Callback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	userID, err := h.authSvc.ValidateStateToken(state)
	if err != nil || code == "" {
		http.Redirect(w, r, h.appDomain+"/settings?google=error", http.StatusFound)
		return
	}

	tok, err := h.oauth.Exchange(r.Context(), code)
	if err != nil {
		slog.Error("google callback: exchange failed", "error", err)
		http.Redirect(w, r, h.appDomain+"/settings?google=error", http.StatusFound)
		return
	}

	// Best-effort — a missing email just means Settings won't have a nice
	// label to show; it doesn't block the connection itself.
	email, err := h.oauth.FetchEmail(r.Context(), h.oauth.TokenSource(r.Context(), tok))
	if err != nil {
		slog.Warn("google callback: fetch email failed", "userId", userID, "error", err)
	}

	if _, err := h.accounts.Upsert(r.Context(), userID, email, tok.AccessToken, tok.RefreshToken, tok.Expiry, h.oauth.Scope()); err != nil {
		slog.Error("google callback: upsert account failed", "userId", userID, "error", err)
		http.Redirect(w, r, h.appDomain+"/settings?google=error", http.StatusFound)
		return
	}

	// Sync immediately in the background so the user doesn't have to wait
	// for the next tick to see anything land.
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		if err := h.sync.SyncUser(ctx, userID); err != nil {
			slog.Error("google callback: initial sync failed", "userId", userID, "error", err)
		}
	}()

	http.Redirect(w, r, h.appDomain+"/settings?google=connected", http.StatusFound)
}

func (h *GoogleHandler) Status(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	account, err := h.accounts.GetByUserID(r.Context(), authUser.ID)
	if errors.Is(err, repository.ErrNotFound) {
		writeJSON(w, http.StatusOK, map[string]any{"connected": false})
		return
	}
	if err != nil {
		http.Error(w, "failed to fetch status", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"connected":    true,
		"email":        account.GoogleEmail,
		"lastSyncedAt": account.LastSyncedAt,
	})
}

func (h *GoogleHandler) Disconnect(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	account, err := h.accounts.GetByUserID(r.Context(), authUser.ID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		http.Error(w, "failed to disconnect", http.StatusInternalServerError)
		return
	}
	if err == nil {
		if revokeErr := google.RevokeToken(r.Context(), account.RefreshToken); revokeErr != nil {
			slog.Warn("google disconnect: revoke failed", "userId", authUser.ID, "error", revokeErr)
		}
	}

	if err := h.accounts.Delete(r.Context(), authUser.ID); err != nil && !errors.Is(err, repository.ErrNotFound) {
		http.Error(w, "failed to disconnect", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Sync triggers an immediate sync pass for the caller, so the UI can offer a
// "Sync now" button instead of waiting for the next scheduled tick.
func (h *GoogleHandler) Sync(w http.ResponseWriter, r *http.Request) {
	authUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	if err := h.sync.SyncUser(r.Context(), authUser.ID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "google calendar not connected", http.StatusNotFound)
			return
		}
		http.Error(w, "sync failed", http.StatusInternalServerError)
		return
	}

	account, err := h.accounts.GetByUserID(r.Context(), authUser.ID)
	if err != nil {
		http.Error(w, "sync completed but failed to fetch status", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"connected":    true,
		"email":        account.GoogleEmail,
		"lastSyncedAt": account.LastSyncedAt,
	})
}
