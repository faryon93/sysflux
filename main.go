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
	"flag"
	"os"
	"syscall"
	"time"

	"github.com/faryon93/util"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/sirupsen/logrus"
)

// ---------------------------------------------------------------------------------------
//  application entry
// ---------------------------------------------------------------------------------------

func main() {
	var colors bool
	var timeout time.Duration
	flag.BoolVar(&colors, "colors", false, "force color logging")
	flag.DurationVar(&timeout, "timeout", 800*time.Millisecond, "influxdb write timeout")
	flag.Parse()

	// setup logger
	formater := logrus.TextFormatter{ForceColors: colors}
	logrus.SetFormatter(&formater)
	logrus.SetOutput(os.Stdout)

	logrus.Infoln("starting", GetAppVersion())

	conf, err := LoadConf()
	if err != nil {
		panic(err)
	}

	// construct the influxdb configuration
	influxConfig := client.HTTPConfig{
		Addr:     conf.Influx.Addr,
		Username: conf.Influx.User,
		Password: conf.Influx.Password,
		Timeout:  timeout,
	}

	recorders := make([]*Recorder, 0)
	for i, syslog := range conf.Syslog {
		influx, err := client.NewHTTPClient(influxConfig)
		if err != nil {
			logrus.Errorf("syslog(%d): failed to create influx client: %s", i, err.Error())
			continue
		}

		logrus.Infof("starting syslog(%d) listener (sz: %d, timeout: %s, listen: %s)",
			i, syslog.BatchSize, syslog.BatchTimeout, syslog.Listen)

		// setup the recorder
		rec := Recorder{Influx: influx, Conf: *syslog}
		err = rec.Setup()
		if err != nil {
			logrus.Errorf("syslog(%d): failed to setup recorders: %s", i, err.Error())
			continue
		}

		// start the recorder
		recorders = append(recorders, &rec)
		go rec.Run()
	}

	// wait for stop signals
	util.WaitSignal(os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	logrus.Infoln("received SIGINT / SIGTERM going to shutdown")

	for i, rec := range recorders {
		rec.Stop()
		logrus.Infof("stopped syslog(%d) recorder", i)
	}
}
