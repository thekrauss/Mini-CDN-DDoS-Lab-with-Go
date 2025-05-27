package config

import (
	"slices"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		GRPCPort int    `mapstructure:"grpc_port"`
		HTTPPort int    `mapstructure:"http_port"`
		TLSCert  string `mapstructure:"tls_cert"`
		TLSKey   string `mapstructure:"tls_key"`
	}
	JWT struct {
		SecretKey          string        `mapstructure:"secret_key"`
		RefreshSecret      string        `mapstructure:"refresh_secret"`
		AccessTokenExpiry  time.Duration `mapstructure:"access_token_expiry"`
		RefreshTokenExpiry time.Duration `mapstructure:"refresh_token_expiry"`
		Issuer             string        `mapstructure:"issuer"`
	}
	Database struct {
		Host     string
		Port     int
		User     string
		Password string
		Name     string
		SSLMode  string
	}
	Redis struct {
		Host     string
		Port     int
		Password string
		DB       int
	}
	Metrics struct {
		PrometheusEnabled bool `mapstructure:"prometheus_enabled"`
		PrometheusPort    int  `mapstructure:"prometheus_port"`
	}
	Logging struct {
		Level  string
		Format string
		File   string
	}
	Tenancy struct {
		MultiTenant   bool
		DefaultTenant string
	}
	Security struct {
		MaxFailedAttempts int           `mapstructure:"max_failed_attempts"`
		LockoutDuration   time.Duration `mapstructure:"lockout_duration"`
		AllowedOrigins    []string      `mapstructure:"allowed_origins"`
	}
	Tracking struct {
		EnableIPLogging        bool `mapstructure:"enable_ip_logging"`
		EnableUserAgentLogging bool `mapstructure:"enable_user_agent_logging"`
	}
	RolesA []string `mapstructure:"roles_a"`
	RolesB []string `mapstructure:"roles_b"`
}

var AppConfig Config

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	AppConfig = config

	return &config, nil
}

func IsRoleA(role string) bool {
	return slices.Contains(AppConfig.RolesA, role)
}

func IsRoleB(role string) bool {
	return slices.Contains(AppConfig.RolesB, role)
}
