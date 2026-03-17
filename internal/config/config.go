package config

import (
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

const (
	namespaceId string = "85f59dc8-caee-41a7-a143-c6ef80b07fbc"
)

// Returns the constant namespace uuid for this microservice. This namespace id
// was generated randomly according to uuid v4, and can be used as a valid
// namespace id when generating uuid v5. This uuid should be concidered an
// application specific namespace id as per RFC9562 §6.6
func GetNamespaceId() uuid.UUID {
	uuid, err := uuid.Parse(namespaceId)
	if err != nil {
		panic(fmt.Sprintf("namespace id is no longer a valid uuid, make sure the namespace id constant is a parsable uuid and redeploy: %s", err.Error()))
	}
	return uuid
}

type Config struct {
	Develop      bool `mapstructure:"viti_agent_develop"`
	LogLevel     int  `mapstructure:"viti_agent_log_level"`
	PollInterval int  `mapstructure:"viti_agent_poll_interval"`

	VitiHost  string `mapstructure:"viti_host"`
	VitiPort  string `mapstructure:"viti_port"`
	VitiToken string `mapstructure:"viti_token"`

	RorUrl     string `mapstructure:"ror_url"`
	RorRole    string `mapstructure:"ror_role"`
	RorCommit  string `mapstructure:"ror_commit"`
	RorVersion string `mapstructure:"ror_version"`
	RorApikey  string `mapstructure:"ror_apikey"`

	VaultUrl string `mapstructure:"vault_url"`

	RabbitMQRole          string `mapstructure:"rabbitmq_role"`
	RabbitMQHost          string `mapstructure:"rabbitmq_host"`
	RabbitMQPort          string `mapstructure:"rabbitmq_port"`
	RabbitMQBroadcastName string `mapstructure:"rabbitmq_broadcast_name"`
}

const (
	Develop      = "VITI_AGENT_DEVELOP"
	LogLevel     = "VITI_AGENT_LOG_LEVEL"
	PollInterval = "VITI_AGENT_POLL_INTERVAL"

	VitiHost  = "VITI_HOST"
	VitiPort  = "VITI_PORT"
	VitiToken = "VITI_TOKEN"

	RorUrl     = "ROR_URL"
	RorRole    = "ROR_ROLE"
	RorCommit  = "ROR_COMMIT"
	RorVersion = "ROR_VERSION"
	RorApikey  = "ROR_APIKEY"

	VaultUrl = "VAULT_URL"

	RabbitMQRole          = "RABBITMQ_ROLE"
	RabbitMQHost          = "RABBITMQ_HOST"
	RabbitMQPort          = "RABBITMQ_Port"
	RabbitMQBroadcastName = "RABBITMQ_BROADCAST_NAME"
)

func NewConfig() (*Config, error) {
	viper.AutomaticEnv()

	viper.SetDefault("viti_agent_log_level", 0)
	viper.SetDefault("viti_agent_develop", false)

	loadDefaultConfigParameters()
	setConfigSources()

	err := viper.ReadInConfig()
	if err != nil {
		slog.Warn("could not find config files, using defaults", "viper error", err)
	}
	var config Config

	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	return &config, nil
}

func loadDefaultConfigParameters() {
	_ = viper.BindEnv("viti_agent_develop", Develop)
	_ = viper.BindEnv("viti_agent_log_level", LogLevel)
	_ = viper.BindEnv("viti_agent_poll_interval", PollInterval)

	_ = viper.BindEnv("viti_host", VitiHost)
	_ = viper.BindEnv("viti_port", VitiPort)
	_ = viper.BindEnv("viti_token", VitiToken)

	_ = viper.BindEnv("ror_url", RorUrl)
	_ = viper.BindEnv("ror_role", RorRole)
	_ = viper.BindEnv("ror_commit", RorCommit)
	_ = viper.BindEnv("ror_version", RorVersion)
	_ = viper.BindEnv("ror_apikey", RorApikey)
}

func setConfigSources() {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	// config from working dir
	// really just a hack since i always run the program from repo root
	// not very robust
	wd, err := os.Getwd()
	if err != nil {
		slog.Error("could not get working dir", "error", err)
		return
	}

	viper.AddConfigPath(wd)

	// config from ~/devel
	home, err := os.UserHomeDir()
	if err != nil {
		slog.Error("could not find user home dir", "error", err)
		return
	}
	devel := path.Join(home, "devel")

	viper.AddConfigPath(devel)
}
