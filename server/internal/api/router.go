package api

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/lnardon/arete/internal/api/handlers"
	"github.com/lnardon/arete/internal/api/middleware"
	"github.com/lnardon/arete/internal/auth"
)

type RouterConfig struct {
	StaticDir         string
	HabitHandler      *handlers.HabitHandler
	CompletionHandler *handlers.CompletionHandler
	AuthHandler       *handlers.AuthHandler
	AuthService       *auth.Service
	AllowedOrigin     string
}

// Serves static files and falls back to index.html for SPA client-side routes.
func spaHandler(staticDir string) http.Handler {
	fs := http.FileServer(http.Dir(staticDir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(staticDir, r.URL.Path)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}

		indexPath := filepath.Join(staticDir, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			http.ServeFile(w, r, indexPath)
			return
		}
		http.NotFound(w, r)
	})
}

func NewRouter(config RouterConfig) http.Handler {
	r := mux.NewRouter()
	r.Use(middleware.CORS(config.AllowedOrigin))
	r.Use(middleware.Logging)

	jwtMiddleware := middleware.JWT(config.AuthService)

	// Public routes
	public := r.PathPrefix("/api/v1").Subrouter()
	public.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	}).Methods("GET")
	public.HandleFunc("/auth/register", config.AuthHandler.Register).Methods("POST")
	public.HandleFunc("/auth/login", config.AuthHandler.Login).Methods("POST")
	public.HandleFunc("/auth/logout", config.AuthHandler.Logout).Methods("POST")

	// Protected routes
	protected := r.PathPrefix("/api/v1").Subrouter()
	protected.Handle("/auth/me", jwtMiddleware(http.HandlerFunc(config.AuthHandler.Me))).Methods("GET")
	protected.Handle("/habits", jwtMiddleware(http.HandlerFunc(config.HabitHandler.ListHabits))).Methods("GET")
	protected.Handle("/habits", jwtMiddleware(http.HandlerFunc(config.HabitHandler.CreateHabit))).Methods("POST")
	protected.Handle("/habits/{id}", jwtMiddleware(http.HandlerFunc(config.HabitHandler.UpdateHabit))).Methods("PUT")
	protected.Handle("/habits/{id}", jwtMiddleware(http.HandlerFunc(config.HabitHandler.DeleteHabit))).Methods("DELETE")
	protected.Handle("/completions", jwtMiddleware(http.HandlerFunc(config.CompletionHandler.GetCompletions))).Methods("GET")
	protected.Handle("/completions/range", jwtMiddleware(http.HandlerFunc(config.CompletionHandler.GetCompletionsRange))).Methods("GET")
	protected.Handle("/completions/toggle", jwtMiddleware(http.HandlerFunc(config.CompletionHandler.ToggleCompletion))).Methods("POST")

	// Serve frontend SPA
	if config.StaticDir != "" {
		r.PathPrefix("/").Handler(spaHandler(config.StaticDir)).Methods("GET", "OPTIONS")
	}

	return r
}
