package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	DbHost string `default:"progtest.opentech.local" split_words:"true"`
	DbPort string `default:"6432" split_words:"true"`
	DbUser string `default:"auth_service" split_words:"true"`
	DbPass string `default:"PzB6YmaX" split_words:"true"`
	DbName string `default:"opentech" split_words:"true" `

	LdapHost   string `default:"ldap.opentech.local" split_words:"true"`
	LdapPort   string `default:"389" split_words:"true"`
	LdapLogin  string `default:"cisco" split_words:"true"`
	LdapPass   string `default:"cisco" split_words:"true"`
	LdapBaseDn string `default:"ou=people,dc=opentech,dc=local" split_words:"true"`

	AccountsAPIURL string `default:"http://api.shop2.test/user" split_words:"true"`

	RedisHost string `default:"redis.opentech.local:6379" split_words:"true"`
	RedisPass string `default:"" split_words:"true"`
	RedisDb   int    `default:"0" split_words:"true"`
}

// NewConfig loads ENV variables to Config structure
func NewConfig() (*Config, error) {
	var c Config
	if err := envconfig.Process(ServiceName, &c); err != nil {
		return nil, err
	}

	return &c, nil
}

