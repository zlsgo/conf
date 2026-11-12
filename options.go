package conf

type Options struct {
	// FileName is the config file name without extension, e.g. "config".
	FileName string
	// AutomaticEnv, when enabled, makes every Get attempt to override the config
	// value with an environment variable of the form "EnvPrefix_KEY".
	AutomaticEnv bool
	// EnvPrefix is the environment variable prefix, e.g. "APP".
	EnvPrefix string
	// AutoCreate, when enabled, automatically creates the config file if it does not exist.
	AutoCreate bool
	// PrimaryAliss is the alias of the primary config, e.g. "app", used to
	// distinguish between different config files.
	PrimaryAliss string
	// EnvKeyReplacer specifies the string that replaces "." in keys when looking up
	// environment variables. Environment variable names cannot contain ".", so it
	// defaults to "_" when left empty, e.g. db.host maps to DB_HOST.
	EnvKeyReplacer string
	// AllowEmptyEnv, when enabled, also treats environment variables with empty
	// string values as valid overrides.
	AllowEmptyEnv bool
}
