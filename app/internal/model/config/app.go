package config

type App struct {
	PrimaryDomain string `mapstructure:"primaryDomain" yaml:"primaryDomain"`
	Domain        string `mapstructure:"domain" yaml:"domain"`
	PrefixUrl     string `mapstructure:"prefixUrl" yaml:"prefixUrl"`
}
