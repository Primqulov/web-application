package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	MongoURI string
	MongoDB  string

	HTTPAddr string

	JWTAccessSecret  string
	JWTRefreshSecret string
	JWTAccessTTL     time.Duration
	JWTRefreshTTL    time.Duration

	CORSOrigins []string
	AvatarDir   string

	AdminSeedUser string
	AdminSeedPass string

	BotSharedSecret string
	OTPLength       int
	OTPTTL          time.Duration
	OTPDevReturn    bool

	TelegramBotToken    string
	TelegramBotUsername string

	// AWS S3
	AWSRegion          string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSS3Bucket        string
	AWSS3PublicBaseURL string
}

func envStr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
func envInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}
func envBool(k string, def bool) bool {
	v := strings.ToLower(os.Getenv(k))
	switch v {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	}
	return def
}

func Load() Config {
	return Config{
		MongoURI:            envStr("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:             envStr("MONGO_DB", "ishchibormi"),
		HTTPAddr:            envStr("HTTP_ADDR", ":8080"),
		JWTAccessSecret:     envStr("JWT_ACCESS_SECRET", "dev-access-secret"),
		JWTRefreshSecret:    envStr("JWT_REFRESH_SECRET", "dev-refresh-secret"),
		JWTAccessTTL:        time.Duration(envInt("JWT_ACCESS_TTL_MIN", 60)) * time.Minute,
		JWTRefreshTTL:       time.Duration(envInt("JWT_REFRESH_TTL_HRS", 720)) * time.Hour,
		CORSOrigins:         strings.Split(envStr("CORS_ORIGINS", "http://localhost:3000"), ","),
		AvatarDir:           envStr("AVATAR_DIR", "./data/avatars"),
		AdminSeedUser:       envStr("ADMIN_SEED_USER", "admin"),
		AdminSeedPass:       envStr("ADMIN_SEED_PASS", "Admin123!"),
		BotSharedSecret:     envStr("BOT_SHARED_SECRET", "dev-shared"),
		OTPLength:           envInt("OTP_LENGTH", 6),
		OTPTTL:              time.Duration(envInt("OTP_TTL_SECONDS", 300)) * time.Second,
		OTPDevReturn:        envBool("OTP_DEV_RETURN", true),
		TelegramBotToken:    envStr("TELEGRAM_BOT_TOKEN", ""),
		TelegramBotUsername: envStr("TELEGRAM_BOT_USERNAME", ""),

		AWSRegion:          envStr("AWS_REGION", "eu-central-1"),
		AWSAccessKeyID:     envStr("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey: envStr("AWS_SECRET_ACCESS_KEY", ""),
		AWSS3Bucket:        envStr("AWS_S3_BUCKET", ""),
		AWSS3PublicBaseURL: envStr("AWS_S3_PUBLIC_BASE_URL", ""),
	}
}
