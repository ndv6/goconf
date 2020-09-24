package goconf

import (
	"log"
	"os"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

type Source string

const (
	// default values by convention
	DefaultType     = "json"
	DefaultFilename = "config"

	// environment variable key names
	EnvConsulHostKey = "GOCONF_CONSUL"
	EnvTypeKey       = "GOCONF_TYPE"
	EnvFileNameKey   = "GOCONF_FILENAME"
	EnvPrefixKey     = "GOCONF_ENV_PREFIX"

	//configuration sources
	SourceEnv    Source = "env"
	SourceFile   Source = "file"
	SourceConsul Source = "consul"
)

var (
	typ    = DefaultType
	fname  = DefaultFilename
	prefix string

	c    *viper.Viper
	dirs = []string{
		".",
		"$HOME",
		"/usr/local/etc",
		"/etc",
	}

	errEnv, errFile, errConsul error
)

func init() {
	Configure()
}

//Configure bootstrap configuration for this service identified by name
func Configure() {
	// first lets load .env file
	if err := godotenv.Load(); err != nil {
		errEnv = errors.Cause(err)
	}

	if v := os.Getenv(EnvTypeKey); len(v) > 0 {
		typ = v
	}
	if v := os.Getenv(EnvFileNameKey); len(v) > 0 {
		fname = v
	}
	if v := os.Getenv(EnvPrefixKey); len(v) > 0 {
		prefix = v
	}

	// setup and configure viper instance
	c = viper.New()
	c.SetConfigType(typ)
	c.SetConfigName(fname)
	if len(prefix) > 0 {
		c.SetEnvPrefix(prefix)
	}
	c.AutomaticEnv()

	// next we load from consul; only if consul host defined
	if ch := os.Getenv(EnvConsulHostKey); ch != "" {
		if err := c.AddRemoteProvider("consul", ch, fname); err != nil {
			errConsul = errors.Cause(err)
		} else {
			connect := func() error { return c.ReadRemoteConfig() }
			notify := func(err error, t time.Duration) { log.Println("[goconf]", err.Error(), t) }
			b := backoff.NewExponentialBackOff()
			b.MaxElapsedTime = 2 * time.Minute

			err := backoff.RetryNotify(connect, b, notify)
			if err != nil {
				log.Printf("[goconf] giving up connecting to remote config ")
				errConsul = errors.Cause(err)
			}
		}
	} else {
		errConsul = errors.New("failed loading remote source; ENV not defined")
	}

	// last, we attempt to load from file in configured dir
	for _, d := range dirs {
		c.AddConfigPath(d)
	}
	if err := c.ReadInConfig(); err != nil {
		errFile = errors.Cause(err)
	}
}

func MustSource(s ...Source) {
	if len(s) == 0 {
		if errEnv != nil && errFile != nil && errConsul != nil {
			log.Fatalln("no configuration loaded from any possible source")
		}
		return
	}
	for _, v := range s {
		if err := Err(v); err != nil {
			log.Fatalf("%+v\n", err)
		}
	}
}

func MustLoad(s ...string) {
	for _, k := range s {
		if nil == c.Get(k) {
			log.Fatalf("config [%s] is not defined\n", k)
		}
	}
}

func Err(s Source) error {
	switch s {
	case SourceEnv:
		return errEnv
	case SourceConsul:
		return errConsul
	default:
		return errFile
	}
}

//Config retrieve config instance
func Config() *viper.Viper {
	return c
}

func Get(k string) interface{} {
	return c.Get(k)
}

func GetString(k string) string {
	return c.GetString(k)
}

func GetInt(k string) int {
	return c.GetInt(k)
}

func GetFloat64(k string) float64 {
	return c.GetFloat64(k)
}

func GetStringSlice(k string) []string {
	return c.GetStringSlice(k)
}
