package config

import (
	"flag"
	"github.com/BurntSushi/toml"
	golog "github.com/JUYAFEI/go-framework/log"
	"os"
)

type Config struct {
	Log      map[string]any
	Template map[string]any
	Db       map[string]any
	Pool     map[string]any
}

var Conf = &Config{
	Log:      make(map[string]any),
	Template: make(map[string]any),
	Db:       make(map[string]any),
	Pool:     make(map[string]any),
}

func init() {
	loadToml()
}

func loadToml() {
	confFile := flag.String("conf", "conf/app.toml", "app config file")
	flag.Parse()
	if _, err := os.Stat(*confFile); err != nil {
		golog.DefaultLogger().Info("conf/app.toml file not load，because not exist")
		return
	}

	_, err := toml.DecodeFile(*confFile, Conf)
	if err != nil {
		golog.DefaultLogger().Info("conf/app.toml decode fail check format")
		return
	}
}
