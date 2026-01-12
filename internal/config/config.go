package config

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Storage  StorageConfig  `mapstructure:"storage"`
	Security SecurityConfig `mapstructure:"security"`
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type StorageConfig struct {
	BasePath string `mapstructure:"basePath"`
}

type SecurityConfig struct {
	Enabled bool          `mapstructure:"enabled"`
	Admin   AdminConfig   `mapstructure:"admin"`
	Captcha CaptchaConfig `mapstructure:"captcha"`
}

type AdminConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type CaptchaConfig struct {
	Enabled            bool   `mapstructure:"enabled"`
	TurnstileSiteKey   string `mapstructure:"turnstileSiteKey"`
	TurnstileSecretKey string `mapstructure:"turnstileSecretKey"`
}

var GlobalConfig Config

func Init() error {
	// 1. Set Defaults
	viper.SetDefault("server.port", 8080)
	cwd, _ := os.Getwd()
	viper.SetDefault("storage.basePath", cwd)
	viper.SetDefault("security.enabled", false)
	viper.SetDefault("security.captcha.enabled", false)
	viper.SetDefault("security.captcha.turnstileSiteKey", "")
	viper.SetDefault("security.captcha.turnstileSecretKey", "")

	// 2. Define CLI Flags
	// We use zero values for defaults here so we can distinguish if user provided them
	pflag.Int("port", 0, "Port to run the server on (default 8080)")
	pflag.String("data-dir", "", "Base directory for data storage (default current directory)")
	pflag.Bool("auth", false, "Enable authentication")
	pflag.String("user", "", "Admin username")
	pflag.String("password", "", "Admin password")
	pflag.String("config", "", "Path to config file")
	pflag.Bool("verify", false, "Enable Cloudflare Turnstile verification")
	pflag.String("turnstile-sitekey", "", "Cloudflare Turnstile site key")
	pflag.String("turnstile-secret", "", "Cloudflare Turnstile secret key")

	pflag.Parse()

	// 3. Bind Flags to Config only if they are changed (provided by user)
	// Or explicitly bind specific flags that we know map 1:1 and rely on 0-value logic if acceptable,
	// but for "port", 0 is not valid, so if it's 0 we know it wasn't set (assuming valid port > 0).
	// For bools, default is false. If user sets --auth, it becomes true.

	if pflag.Lookup("port").Changed {
		viper.BindPFlag("server.port", pflag.Lookup("port"))
	}
	if pflag.Lookup("data-dir").Changed {
		viper.BindPFlag("storage.basePath", pflag.Lookup("data-dir"))
	}
	if pflag.Lookup("auth").Changed {
		viper.BindPFlag("security.enabled", pflag.Lookup("auth"))
	}
	if pflag.Lookup("user").Changed {
		viper.BindPFlag("security.admin.username", pflag.Lookup("user"))
	}
	if pflag.Lookup("password").Changed {
		viper.BindPFlag("security.admin.password", pflag.Lookup("password"))
	}
	if pflag.Lookup("verify").Changed {
		viper.BindPFlag("security.captcha.enabled", pflag.Lookup("verify"))
	}
	if pflag.Lookup("turnstile-sitekey").Changed {
		viper.BindPFlag("security.captcha.turnstileSiteKey", pflag.Lookup("turnstile-sitekey"))
	}
	if pflag.Lookup("turnstile-secret").Changed {
		viper.BindPFlag("security.captcha.turnstileSecretKey", pflag.Lookup("turnstile-secret"))
	}

	// 4. Load Config File
	configFile, _ := pflag.CommandLine.GetString("config")
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName("minimax-voice-workbench")
		viper.SetConfigType("yaml")

		// Add search paths
		if home, err := os.UserHomeDir(); err == nil {
			viper.AddConfigPath(home)
		}
		viper.AddConfigPath(".")

		// Linux specific paths
		if runtime.GOOS == "linux" {
			viper.AddConfigPath("/etc")
			viper.AddConfigPath("/usr/local/etc")
		}
	}

	// Environment variables
	viper.SetEnvPrefix("MVW")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}

	if err := viper.Unmarshal(&GlobalConfig); err != nil {
		return fmt.Errorf("unable to decode into struct: %w", err)
	}

	return nil
}
