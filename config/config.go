package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Env    string   `json:"env" yaml:"env"`
	Server Server   `json:"server" yaml:"server"`
	Logger Logger   `json:"log" yaml:"log"`
	Room   Room     `json:"room" yaml:"room"`
	Groups []string `json:"groups" yaml:"groups"`
}

type Server struct {
	Host string `json:"host" yaml:"host"`
	Port string `json:"port" yaml:"port"`
}

func (s *Server) setDefault() {
	if s.Host == "" {
		s.Host = "0.0.0.0"
	}
}

type Logger struct {
	Level    slog.Level `json:"level" yaml:"level"`
	Filepath string     `json:"filepath" yaml:"filepath"`
}

func (l *Logger) setDefault() {
	if l.Level == 0 {
		l.Level = slog.LevelDebug
	}
}

type Room struct {
	NumRows     int `yaml:"num_rows"`
	NumCols     int `yaml:"num_cols"`
	MinDistance int `yaml:"min_distance"`
}

type setDefaulter interface {
	setDefault()
}

func Read(getEnv func(string) string) (empty AppConfig, err error) {
	filename := getEnv("CONFIG_PATH")

	f, err := os.Open(filepath.Clean(filename))
	if err != nil {
		return empty, fmt.Errorf("open config: %w", err)
	}
	defer func() {
		err = errors.Join(err, f.Close())
	}()

	var appCfg AppConfig
	if err = yaml.NewDecoder(f).Decode(&appCfg); err != nil {
		return empty, fmt.Errorf("parse config: %w", err)
	}

	rv := reflect.ValueOf(appCfg)
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		if !field.CanInterface() {
			continue
		}

		if d, ok := field.Interface().(setDefaulter); ok {
			d.setDefault()
		}
	}

	if appCfg.Env == "" {
		appCfg.Env = "local"
	}

	return appCfg, nil
}
