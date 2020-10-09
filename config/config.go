package config

var (
	Config WeepConfig
)

type MetaDataPath struct {
	Path string `yaml:"path"`
	Data string `yaml:"data"`
}

type MetaDataConfig struct {
	Routes []MetaDataPath `yaml:"routes"`
}

type MtlsSettings struct {
	Cert string `yaml:"cert"`
	Key string `yaml:"key"`
}

type ChallengeSettings struct {
	User string `yaml:"user"`
}

type WeepConfig struct {
	MetaData MetaDataConfig `yaml:"metadata"`
	ConsoleMeUrl string `yaml:"consoleme_url"`
	MtlsSettings MtlsSettings `yaml:"mtls_settings"`
	ChallengeSettings ChallengeSettings `yaml:"challenge_settings"`
	AuthenticationMethod string `yaml:"authentication_method"`
}
