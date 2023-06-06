package config

type EMail struct {
	Account  string `mapstructure:"account" yaml:"account"`
	Password string `mapstructure:"password" yaml:"password"`
}
