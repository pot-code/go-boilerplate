package infra

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// ContextLogger .
type ContextLogger string

// EnvPrefix env prefix for viper
const EnvPrefix = "GOAPP"

// ContextLoggerKey logger key in request context
const ContextLoggerKey ContextLogger = "logger"

// AppConfig App option object
type AppConfig struct {
	AppID          string        `mapstructure:"app_id" json:"app_id" yaml:"app_id" validate:"required"`            // Application ID
	Host           string        `mapstructure:"host" json:"host" yaml:"host"`                                      // bind host address
	Port           int           `mapstructure:"port" json:"port" yaml:"port"`                                      // bind listen port
	Env            string        `mapstructure:"env" json:"env" yaml:"env" validate:"oneof=development production"` // runtime environment
	SessionTimeout time.Duration `mapstructure:"session_timeout" json:"session_timeout" yaml:"session_timeout"`
	SessionRefresh time.Duration `mapstructure:"session_refresh" json:"session_refresh" yaml:"session_refresh"` // session refresh threshold
	Database       struct {
		Driver   string `mapstructure:"driver" json:"driver" yaml:"driver" validate:"required"`                      // driver name
		Host     string `mapstructure:"host" json:"host" yaml:"host" validate:"required"`                            // server host
		MaxConn  int32  `mapstructure:"maxconn" json:"maxconn" yaml:"maxconn" validate:"min=100"`                    // maximum opening connections number
		Password string `mapstructure:"password" json:"password" yaml:"password" validate:"required"`                // db password
		Port     int    `mapstructure:"port" json:"port" yaml:"port"`                                                // server port
		Protocol string `mapstructure:"protocol" json:"protocol" yaml:"protocol" validate:"omitempty,oneof=tcp udp"` // connection protocol, eg.tcp
		Query    string `mapstructure:"query" json:"query" yaml:"query"`                                             // DSN query parameter
		Schema   string `mapstructure:"schema" json:"schema" yaml:"schema" validate:"required"`                      // use schema
		User     string `mapstructure:"username" json:"username" yaml:"username" validate:"required"`                // db username
	} `mapstructure:"database" json:"database" yaml:"database"`
	Logging struct {
		FilePath string `mapstructure:"file_path" json:"file_path" yaml:"file_path"`                            // log file path
		Level    string `mapstructure:"level" json:"level" yaml:"level" validate:"oneof=debug info warn error"` // global logging level
	} `mapstructure:"logging" json:"logging" yaml:"logging"`
	Security struct {
		IDLength         int           `mapstructure:"id_length" json:"id_length" yaml:"id_length"` // length of generated ID for entities
		JWTMethod        string        `mapstructure:"jwt_method" json:"jwt_method" yaml:"jwt_method" validate:"oneof=HS256 HS512 ES256"`
		JWTSecret        string        `mapstructure:"jwt_secret" json:"jwt_secret" yaml:"jwt_secret" validate:"required"`
		TokenName        string        `mapstructure:"token_name" json:"token_name" yaml:"token_name" validate:"required"`     // jwt token name set in cookie
		MaxLoginAttempts int           `mapstructure:"max_login_attempts" json:"max_login_attempts" yaml:"max_login_attempts"` // maximum login attempts
		RetryTimeout     time.Duration `mapstructure:"retry_timeout" json:"retry_timeout" yaml:"retry_timeout"`                // retry wait
	} `mapstructure:"security" json:"security" yaml:"security"`
	KVStore struct {
		Host     string `mapstructure:"host" json:"host" yaml:"host"`                                 // bind host address
		Port     int    `mapstructure:"port" json:"port" yaml:"port"`                                 // bind listen port
		Password string `mapstructure:"password" json:"password" yaml:"password" validate:"required"` // password for security reasons
	} `mapstructure:"kv" json:"kv" yaml:"kv"`
	DevOP struct {
		APM bool `mapstructure:"apm" json:"apm" yaml:"apm"`
	} `mapstructure:"devop" json:"devop" yaml:"devop"`
}

// InitConfig init app config using viper
func InitConfig() (*AppConfig, error) {
	// app
	pflag.String("host", "", "binding address")
	pflag.String("app_id", "", "application identifier (required)")
	pflag.String("env", "development", "runtime environment, can be 'development' or 'production'")
	pflag.Int("port", 8081, "listening port")
	pflag.Duration("session_timeout", 30*time.Minute, "JWT lifetime(m, s and h units are supported), eg.30m")
	pflag.Duration("session_refresh", 5*time.Minute, "session refresh threshold(m, s and h units are supported), eg.5m")

	// database
	pflag.String("database.driver", "mysql", "database driver to use")
	pflag.String("database.host", "127.0.0.1", "database host")
	pflag.Int("database.port", 3306, "database server port")
	pflag.String("database.protocol", "", "connection protocol(if mysql is used, this flag must be set), eg.tcp")
	pflag.String("database.username", "", "database username (required)")
	pflag.String("database.password", "", "database password (required)")
	pflag.String("database.schema", "", "database schema (required)")
	pflag.String("database.query", "", `additional DSN query parameters('?' is auto prefixed), if you work with mysql and wish to
work with time.Time, you may specify "parseTime=true"`)
	pflag.Int32("database.maxconn", 200, `max connection count, if you encounter a "too many connections" error, please consider
increasing the max_connection value of your db server, or lower this value`)

	// logging
	pflag.String("logging.level", "info", "logging level")
	pflag.String("logging.file_path", "", "log to file")

	// security
	pflag.Int("security.id_length", 24, "set length of generated ID for entities")
	pflag.String("security.jwt_method", "HS256", "hash algorithm used for JWT auth")
	pflag.String("security.jwt_secret", "", "JWT secret (required)")
	pflag.String("security.token_name", "", "cookie name to store the token (required)")
	pflag.Int("security.max_login_attempts", 3, "maximum login attempts")
	pflag.Duration("security.retry_timeout", 1*time.Hour, "retry wait")

	// kv storage
	pflag.String("kv.host", "127.0.0.1", "kv host")
	pflag.Int("kv.port", 6379, "kv server port")
	pflag.String("kv.password", "", "kv server password (required)")

	// DevOp
	pflag.Bool("devop.apm", false, "enable apm metrics")

	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
	viper.AutomaticEnv()
	viper.SetEnvPrefix(EnvPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	var config = new(AppConfig)
	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}
	if err := validateConfig(config); err != nil {
		return nil, err
	}
	if config.Logging.Level == "debug" {
		if configJSON, err := json.MarshalIndent(config, "", "  "); err == nil {
			log.Printf("App config: %s\n", string(configJSON))
		}
	}
	return config, nil
}

func validateConfig(config *AppConfig) error {
	validate := validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := fld.Tag.Get("json")
		if name == "-" || name == "" {
			name = fld.Tag.Get("yaml")
			if name == "-" || name == "" {
				return ""
			}
		}
		return name
	})
	err := validate.Struct(config)
	if _, ok := err.(*validator.InvalidValidationError); ok {
		log.Fatalf("Failed to validate config: %s", err)
	}
	if err == nil {
		return nil
	}

	var msg []string
	for _, field := range err.(validator.ValidationErrors) {
		namespace := field.Namespace()
		fieldName := namespace[strings.IndexByte(namespace, '.')+1:] // trim top level namespace
		switch field.Tag() {
		case "required":
			msg = append(msg, fmt.Sprintf("%s is required", fieldName))
		case "oneof":
			msg = append(msg, fmt.Sprintf("%s must be one of (%s)", fieldName, field.Param()))
		}
	}
	if len(msg) > 0 {
		return fmt.Errorf("failed to validate config: \n%s", strings.Join(msg, "\n"))
	}
	return nil
}
