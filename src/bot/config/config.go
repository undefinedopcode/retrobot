package config

import "io/ioutil"
import "gopkg.in/yaml.v2"

type TriggerConfig struct {
	Regex string
	Command int
}

type Config struct {
	Username string
	BaseURL string
	IconURL string
	Port int
	Token string
	SlashStrictTokens bool
	SlashTokens []string
	DataDir string
}

func Load(filename string) (*Config, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	err = yaml.Unmarshal(b, cfg)
	return cfg, err
}

func (c *Config) IsTokenValid( isSlash bool, token string ) bool {
	if !isSlash {
		return token == c.Token
	}
	if !c.SlashStrictTokens {
		return true
	}
	for _, t := range c.SlashTokens {
		if t == token {
			return true
		}
	}
	return false
}

func (c *Config) GetDataPath(filename string) string {
	return c.DataDir + "/" + filename
}