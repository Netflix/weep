package main

import (
	"embed"

	"github.com/netflix/weep/cmd"
	"github.com/netflix/weep/config"
)

//go:embed configs/*.yaml
var Configs embed.FS

//go:embed extras/*
var Extras embed.FS

func init() {
	cmd.SetupExtras = Extras
	config.EmbeddedConfigs = Configs
}
