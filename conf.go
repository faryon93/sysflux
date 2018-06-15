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
	"github.com/spf13/viper"
	"strings"
)

// ---------------------------------------------------------------------------------------
//  types
// ---------------------------------------------------------------------------------------

type Conf struct {
	Influx *ConfInflux   `yaml:"influx"`
	Syslog []*ConfSyslog `yaml:"syslog"`
}

type ConfInflux struct {
	Addr     string `yaml:"string"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type ConfSyslog struct {
	Measurement string `yaml:"measurement"`
	Listen      string `yaml:"listen"`
	Regex       string `yaml:"regex"`
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
	return &conf, viper.Unmarshal(&conf)
}
