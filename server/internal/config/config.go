package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	JWT       JWTConfig
	App       AppConfig
	Evolution EvolutionConfig
	OpenAI    OpenAIConfig
}

type ServerConfig struct {
	Port         string
	ReadTimeout  int
	WriteTimeout int
	IdleTimeout  int
	StaticDir    string // path to frontend build ( /app/frontend)
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	SecretKey string
	ExpiresIn int // hours
}

type AppConfig struct {
	Environment  string
	Debug        bool
	AppDomain    string
	CookieSecure bool
	Timezone     string // IANA name the WhatsApp assistant uses to resolve "today"
}

type EvolutionConfig struct {
	BaseURL       string // internal base URL, e.g. http://evolution-api:8080
	APIKey        string
	InstanceName  string
	WebhookSecret string // shared secret checked on inbound webhook requests
}

type OpenAIConfig struct {
	APIKey             string
	Model              string
	TranscriptionModel string
}

func Load() (*Config, error) {
	// prod
	_ = godotenv.Load()

	// local
	_ = godotenv.Load(".env.local")

	config := &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			ReadTimeout:  getEnvAsInt("SERVER_READ_TIMEOUT", 15),
			WriteTimeout: getEnvAsInt("SERVER_WRITE_TIMEOUT", 15),
			IdleTimeout:  getEnvAsInt("SERVER_IDLE_TIMEOUT", 60),
			StaticDir:    getEnv("STATIC_DIR", "frontend/dist"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "hayek_dev"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			SecretKey: getEnv("JWT_SECRET", "your-secret-key"),
			ExpiresIn: getEnvAsInt("JWT_EXPIRES_IN", 24),
		},
		App: AppConfig{
			Environment:  getEnv("ENV", "development"),
			Debug:        getEnvAsBool("DEBUG", false),
			AppDomain:    getEnv("APP_DOMAIN", "http://localhost:5173"),
			CookieSecure: getEnvAsBool("COOKIE_SECURE", true),
			Timezone:     getEnv("APP_TIMEZONE", "America/Sao_Paulo"),
		},
		Evolution: EvolutionConfig{
			BaseURL:       getEnv("EVOLUTION_API_BASE_URL", "http://localhost:8081"),
			APIKey:        getEnv("EVOLUTION_API_KEY", ""),
			InstanceName:  getEnv("EVOLUTION_INSTANCE_NAME", "arete"),
			WebhookSecret: getEnv("EVOLUTION_WEBHOOK_SECRET", ""),
		},
		OpenAI: OpenAIConfig{
			APIKey:             getEnv("OPENAI_API_KEY", ""),
			Model:              getEnv("OPENAI_MODEL", "gpt-5.6-luna"),
			TranscriptionModel: getEnv("OPENAI_TRANSCRIPTION_MODEL", "gpt-4o-mini-transcribe"),
		},
	}

	if len(config.JWT.SecretKey) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters (got %d)", len(config.JWT.SecretKey))
	}

	return config, nil
}

func LoadWithFile(filename string) (*Config, error) {
	if err := godotenv.Load(filename); err != nil {
		return nil, err
	}
	return Load()
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

func getEnvAsBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return fallback
}
