package main

// sysflux
// Copyright (C) 2018 Maximilian Pachl

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// ---------------------------------------------------------------------------------------
//  imports
// ---------------------------------------------------------------------------------------

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

// ---------------------------------------------------------------------------------------
//  types
// ---------------------------------------------------------------------------------------

type Conf struct {
	Influx *ConfInflux   `yaml:"influx"`
	Syslog []*ConfSyslog `yaml:"syslog"`
}

type ConfInflux struct {
	Addr     string
	User     string
	Password string
	Database string
}

type ConfSyslog struct {
	Database     string
	Measurement  string
	Listen       string
	Regex        string
	BatchSize    int           `mapstructure:"batch_size"`
	BatchTimeout time.Duration `mapstructure:"batch_timeout"`
}

// ---------------------------------------------------------------------------------------
//  public functions
// ---------------------------------------------------------------------------------------

// LoadConf loads the configuration file.
func LoadConf() (*Conf, error) {
	// setup config file paths
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/sysflux/")
	viper.SetConfigFile("sysflux.yml")

	// setup env variable parsing
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("SYSFLUX")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	var conf Conf
	err = viper.Unmarshal(&conf)
	if err != nil {
		return nil, err
	}

	// set the default database if no other is specified
	for i := range conf.Syslog {
		if conf.Syslog[i].Database == "" {
			conf.Syslog[i].Database = conf.Influx.Database
		}
	}

	return &conf, nil
}
