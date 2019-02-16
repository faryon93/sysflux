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
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/sirupsen/logrus"
	"gopkg.in/mcuadros/go-syslog.v2"
	"gopkg.in/mcuadros/go-syslog.v2/format"
)

// --------------------------------------------------------------------------------------
//  constants
// --------------------------------------------------------------------------------------

const (
	PrefixTag   = "tag_"
	PrefixValue = "val_"
)

// ---------------------------------------------------------------------------------------
//  types
// ---------------------------------------------------------------------------------------

type Recorder struct {
	Influx client.Client
	Conf   ConfSyslog

	// internal variables
	matcher *regexp.Regexp
	syslog  *syslog.Server
	batch   Batch
}

type Tags map[string]string
type Values map[string]interface{}

// ---------------------------------------------------------------------------------------
//  public functions
// ---------------------------------------------------------------------------------------

func (r *Recorder) Setup() error {
	// compile the regex and make sure it is valid
	matcher, err := regexp.Compile(r.Conf.Regex)
	if err != nil {
		return err
	}
	r.matcher = matcher

	// construct the initial point batch
	r.batch = Batch{
		Timeout:     r.Conf.BatchTimeout,
		Size:        r.Conf.BatchSize,
		Influx:      r.Influx,
		Database:    r.Conf.Database,
		Measurement: r.Conf.Measurement,
	}

	// configure the syslog server
	r.syslog = syslog.NewServer()
	r.syslog.SetFormat(syslog.RFC3164)
	r.syslog.SetHandler(r)

	// boot the udp server to start reception of log messages
	err = r.syslog.ListenUDP(r.Conf.Listen)
	if err != nil {
		return err
	}

	err = r.syslog.Boot()
	if err != nil {
		return err
	}

	// start the timeout write of the batch
	if r.Conf.BatchTimeout == 0 {
		logrus.Warnln("no batch timeout configured: batch writes may be late")
	} else {
		go r.batch.Run()
	}

	return nil
}

// Stop destroys this syslog recorder.
func (r *Recorder) Stop() {
	err := r.syslog.Kill()
	if err != nil {
		logrus.Errorln("failed to stop syslog:", err.Error())
	}
}

// Processes all incomming syslog messages and transforms them
// into influxdb points.
func (r *Recorder) Handle(message format.LogParts, t int64, syslogErr error) {
	// parse the syslog message and make sure everything exists
	timestamp := time.Now()
	content, ok := message["content"].(string)
	if !ok {
		logrus.Warnln("missing message field \"content\": ignoring message")
		return
	}

	// check if the received log messages matches the
	// configured regex
	matches := r.matcher.FindStringSubmatch(content)
	if len(matches) < len(r.matcher.SubexpNames()) {
		logrus.Warnln("ignoring message:", content)
		return
	}

	// process the message
	tags, values, err := r.process(matches)
	if err != nil {
		logrus.Warnln("failed to process message:", err.Error())
		logrus.Infoln(content)
		return
	}

	if host, ok := tags["host"]; ok && strings.Contains(host, "proxy") {
		logrus.Infoln("malformed message:", content)
	}

	err = r.batch.Add(timestamp, tags, values)
	if err != nil {
		logrus.Errorln("failed to write datapoint:", err.Error())
		return
	}
}

// ----------------------------------------------------------------------------------
//  private members
// ----------------------------------------------------------------------------------

// process processes a log messages.
func (r *Recorder) process(matches []string) (Tags, Values, error) {
	// maps which are used to construct the new datapoint
	tags := make(map[string]string)
	values := make(map[string]interface{})

	// process all regex caputure groups and add to the coresponding
	// map in oder to insert the data into the datapoint
	for i, name := range r.matcher.SubexpNames() {
		if i > 0 && len(name) > 0 {
			val := matches[i]

			// we are processing a tag
			if strings.HasPrefix(name, PrefixTag) {
				tags[strings.TrimPrefix(name, PrefixTag)] = val

				// we are processing a value
			} else if strings.HasPrefix(name, PrefixValue) {
				// convert to floating point value
				value, err := strconv.ParseFloat(val, 32)
				if err != nil {
					continue
				}

				values[strings.TrimPrefix(name, PrefixValue)] = float32(value)
			} else {
				return nil, nil, errors.New("unknown capture group naming prefix")
			}
		}
	}

	return tags, values, nil
}
