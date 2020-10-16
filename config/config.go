package config

var (
	Config WeepConfig
)

type MetaDataPath struct {
	Path string `mapstructure:"path"`
	Data string `mapstructure:"data"`
}

type MetaDataConfig struct {
	Routes []MetaDataPath `mapstructure:"routes"`
}

type MtlsSettings struct {
	Cert     string   `mapstructure:"cert"`
	Key      string   `mapstructure:"key"`
	CATrust  string   `mapstructure:"catrust"`
	Insecure bool     `mapstructure:"insecure"`
	Darwin   []string `mapstructure:"darwin"`
	Linux    []string `mapstructure:"linux"`
	Windows  []string `mapstructure:"windows"`
}

type ChallengeSettings struct {
	User string `mapstructure:"user"`
}

type WeepConfig struct {
	MetaData             MetaDataConfig    `mapstructure:"metadata"`
	ConsoleMeUrl         string            `mapstructure:"consoleme_url"`
	MtlsSettings         MtlsSettings      `mapstructure:"mtls_settings"`
	ChallengeSettings    ChallengeSettings `mapstructure:"challenge_settings"`
	AuthenticationMethod string            `mapstructure:"authentication_method"`
}
