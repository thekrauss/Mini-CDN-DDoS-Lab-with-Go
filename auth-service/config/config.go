package config

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/spf13/viper"
	"google.golang.org/api/option"
)

type Config struct {
	Server struct {
		Host     string `mapstructure:"host"`
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
	AuthService struct {
		Host string
		Port int
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
	RolesA      []string `mapstructure:"roles_a"`
	RolesB      []string `mapstructure:"roles_b"`
	GoogleCloud struct {
		SecretManager struct {
			GCloudKey string `mapstructure:"gcloud_key"`
		} `mapstructure:"secret_manager"`
	} `mapstructure:"google_cloud"`
	AppPort              int  `mapstructure:"app_port"`
	UseCloudSecrets      bool `mapstructure:"use_cloud_secrets"`
	UseCloudSecretsForDB bool `mapstructure:"use_cloud_secrets_for_db"`
	GCloudKeyPath        string
	Email                struct {
		SMTPHost     string `mapstructure:"smtp_host"`
		SMTPPort     int    `mapstructure:"smtp_port"`
		SMTPUser     string `mapstructure:"smtp_user"`
		SMTPPassword string `mapstructure:"smtp_password"`
	}

	Firebase struct {
		FirebaseAPIKey      string `mapstructure:"firebase_api_key"`
		FirebaseCredentials string `mapstructure:"firebase_credentials"`
	} `mapstructure:"firebase"`
}

var AppConfig Config

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("erreur de chargement du fichier de configuration: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("erreur d'analyse du fichier de configuration: %w", err)
	}

	AppConfig = config

	if config.UseCloudSecrets {

		if err := setupGoogleCloudCredentials(&config); err != nil {
			return nil, err
		}
		loadSecrets(&config)
		loadRoles(&config)
	}

	return &config, nil
}

func setupGoogleCloudCredentials(config *Config) error {
	secret, err := GetSecret(config.GoogleCloud.SecretManager.GCloudKey)
	if err != nil {
		return fmt.Errorf("erreur récupération du gcloud_key: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "gcloud-key-*.json")
	if err != nil {
		return fmt.Errorf("erreur création fichier temporaire: %w", err)
	}
	defer tmpFile.Close()

	if _, err := tmpFile.Write([]byte(secret)); err != nil {
		return fmt.Errorf("erreur écriture fichier temporaire: %w", err)
	}

	config.GCloudKeyPath = tmpFile.Name()
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", config.GCloudKeyPath)
	log.Printf("GOOGLE_APPLICATION_CREDENTIALS défini sur %s", config.GCloudKeyPath)
	return nil
}

func loadSecrets(config *Config) {

	secrets := map[string]*string{
		"AUTH_DB_HOST":         &config.Database.Host,
		"AUTH_DB_NAME":         &config.Database.Name,
		"AUTH_DB_USER":         &config.Database.User,
		"AUTH_DB_PASSWORD":     &config.Database.Password,
		"FIREBASE_API_KEY":     &config.Firebase.FirebaseAPIKey,
		"FIREBASE_CREDENTIALS": &config.Firebase.FirebaseCredentials,
		"JWT_SECRET_KEY":       &config.JWT.SecretKey,
		"JWT_REFRESH_SECRET":   &config.JWT.RefreshSecret,
	}

	ports := map[string]*int{
		"AUTH_DB_PORT": &config.Database.Port,
		"SMTP_PORT":    &config.Email.SMTPPort,
	}

	for key, field := range secrets {
		if val, err := GetSecret(key); err == nil {
			*field = val
			log.Printf(" Secret %s chargé.", key)
		} else {
			log.Printf(" Erreur chargement secret %s: %v", key, err)
		}
	}

	for key, field := range ports {
		if val, err := GetSecret(key); err == nil {
			if port, err := strconv.Atoi(val); err == nil {
				*field = port
				log.Printf(" Port %s chargé.", key)
			}
		} else {
			log.Printf(" Erreur chargement port %s: %v", key, err)
		}
	}
}

func loadRoles(config *Config) {
	if roles, err := GetRoles("ROLES_A"); err == nil {
		config.RolesA = roles
	}
	if roles, err := GetRoles("ROLES_B"); err == nil {
		config.RolesB = roles
	}
}

func GetSecret(secretName string) (string, error) {
	if !AppConfig.UseCloudSecrets {
		return "", fmt.Errorf("mode local: pas de récupération pour %s", secretName)
	}
	if secretName == "" {
		return "", fmt.Errorf("le nom du secret est vide")
	}

	ctx := context.Background()
	credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credentialsPath == "" {
		return "", fmt.Errorf("GOOGLE_APPLICATION_CREDENTIALS non défini")
	}

	client, err := secretmanager.NewClient(ctx, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		return "", fmt.Errorf("erreur création client SecretManager: %w", err)
	}
	defer client.Close()

	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/syk-microservices/secrets/%s/versions/latest", secretName),
	}
	result, err := client.AccessSecretVersion(ctx, accessRequest)
	if err != nil {
		return "", fmt.Errorf("erreur accès secret %s: %w", secretName, err)
	}

	return string(result.Payload.Data), nil
}

func GetRoles(secretName string) ([]string, error) {
	secretData, err := GetSecret(secretName)
	if err != nil {
		return nil, err
	}
	var roles []string
	if err := json.Unmarshal([]byte(secretData), &roles); err != nil {
		return nil, fmt.Errorf("erreur parsing JSON rôles %s: %w", secretName, err)
	}
	return roles, nil
}
func IsRoleA(role string) bool {
	return slices.Contains(AppConfig.RolesA, role)
}

func IsRoleB(role string) bool {
	return slices.Contains(AppConfig.RolesB, role)
}
