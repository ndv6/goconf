# Configuration

### Definitions

configuration could be provided in different way (ordered by priority)

1. ENV variables
    - utilizing [github.com/joho/godotenv](https://github.com/joho/godotenv)
    - it will try to load `.env` file in current executable directory
    - env prefix key defined via env var `GOCONF_ENV_PREFIX`, if given, all loaded configuration via `.env` should be prefixed
    - all configuration key could have a dot (.) which then translated into nested object
2. File config
    - configuration schema type defined via env var `GOCONF_TYPE`, supported types: `json`, `yaml`, `toml`, if not defined, default value given `json`
    - configuration filename defined via env var `GOCONF_FILENAME` if not defined, default value `config`
    - must exists in one of directory defined, `.`, `$HOME`, `/usr/local/etc`, `/etc`
3. Remote config through [consul](https://www.consul.io/)
    - consul host provided through env variable with key `GOCONF_CONSUL`
    - configuration schema type defined via env var `GOCONF_TYPE`, supported types: `json`, `yaml`, `toml`, if not defined, default value given `json`
    - configuration filename defined via env var `GOCONF_FILENAME` if not defined, default value `config`
    - configuration defined in a single key named `/${GOCONF_FILENAME}`

### Usages
    
Access from any part of codes:

1. initialization done via `init()`, ensure package loaded at the very top of `import` in `main()`, to ensure its executed first before other `init()` which expected loaded config
2. utilizing library [github.com/spf13/viper](https://github.com/spf13/viper) 
3. provided behaviour could be accessed through package methods or resolve viper instance through `Config()` method