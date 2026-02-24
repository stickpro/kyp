package cfg

type ConfigField struct {
	Path              string
	EnvName           string
	DefaultValue      string
	Usage             string
	Example           string
	ValidateParams    string
	IsRequired        bool
	IsSecret          bool
	DisableValidation bool
}
