package goconf

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"log"
)

type (
	Source       string
	RemoteOption struct {
		Type string
		DSN  string
		Key  string
	}
	Option struct {
		Type         string
		Filename     string
		Prefix       string
		Dir          []string
		RemoteOption *RemoteOption
	}
	With func(option *Option)
)

const (
	EnvRemoteType = "GOCONF_RTYPE"
	EnvRemoteDSN  = "GOCONF_RDSN"
	EnvRemoteKey  = "GOCONF_RKEY"

	EnvConfigType = "GOCONF_TYPE"
	EnvConfigName = "GOCONF_FILENAME"
	EnvEnvPrefix  = "GOCONF_ENV_PREFIX"

	SourceEnv    Source = "env"
	SourceFile   Source = "file"
	SourceRemote Source = "consul"
)

var (
	defaultOption = Option{
		Type:     "json",
		Filename: "config",
		Prefix:   "",
		Dir: []string{
			".",
			"$HOME",
			"/usr/local/etc",
			"/etc",
		},
		RemoteOption: nil,
	}

	errEnv, errFile, errRemote error

	c *viper.Viper
)

func WithType(t string) With {
	//@todo limit supported types
	return func(option *Option) {
		option.Type = t
	}
}

func WithFilename(f string) With {
	return func(option *Option) {
		option.Filename = f
	}
}

func WithPrefix(p string) With {
	return func(option *Option) {
		option.Prefix = p
	}
}

func WithDirs(s ...string) With {
	return func(option *Option) {
		option.Dir = s
	}
}

func WithRemote(r RemoteOption) With {
	return func(option *Option) {
		option.RemoteOption = &r
	}
}

func ConfigureOptions(options ...With) {
	var opt = &defaultOption
	for _, o := range options {
		o(opt)
	}
	configure(*opt)
}

func Configure() {
	configure(defaultOption)
}

//Configure bootstrap configuration for this service identified by name
func configure(opt Option) {
	// first lets load .env file
	if err := godotenv.Load(); err != nil {
		errEnv = errors.Cause(err)
	}

	// we override defined value based on given os env
	{
		var (
			t = os.Getenv(EnvRemoteType)
			d = os.Getenv(EnvRemoteDSN)
			k = os.Getenv(EnvRemoteKey)
		)
		if t != "" && d != "" && k != "" {
			opt.RemoteOption = &RemoteOption{t, d, k}
		}
	}

	if v := os.Getenv(EnvConfigType); v != "" {
		opt.Type = v
	}
	if v := os.Getenv(EnvConfigName); v != "" {
		opt.Filename = v
	}
	if v := os.Getenv(EnvEnvPrefix); v != "" {
		opt.Prefix = v
	}

	// setup and configure viper instance
	c = viper.New()
	c.SetConfigType(opt.Type)
	c.SetConfigName(opt.Filename)
	if opt.Prefix != "" {
		c.SetEnvPrefix(opt.Prefix)
	}
	c.AutomaticEnv()

	// next we load from remote; only if configuration given
	if r := opt.RemoteOption; r != nil {
		if err := c.AddRemoteProvider(r.Type, r.DSN, r.Key); err != nil {
			errRemote = errors.Cause(err)
		} else {
			if err := c.ReadRemoteConfig(); err != nil {
				errRemote = errors.Cause(err)
			}
		}
	} else {
		errRemote = errors.New("no remote configuration given")
	}

	// last, we attempt to load from file in configured dir
	for _, d := range opt.Dir {
		c.AddConfigPath(d)
	}
	if err := c.ReadInConfig(); err != nil {
		errFile = errors.Cause(err)
	}
}

func MustSource(s ...Source) {
	if len(s) == 0 {
		if errEnv != nil && errFile != nil && errRemote != nil {
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
	case SourceRemote:
		return errRemote
	default:
		return errFile
	}
}

func Config() *viper.Viper {
	return c
}

func GetString(key string) string  {
	return c.GetString(key)
}
