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
	"sync"
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

// ---------------------------------------------------------------------------------------
//  types
// ---------------------------------------------------------------------------------------

type Batch struct {
	Size        int
	Influx      client.Client
	Timeout     time.Duration
	Database    string
	Measurement string

	batch client.BatchPoints
	timer *time.Timer
	sync.Mutex
}

// ---------------------------------------------------------------------------------------
//  public members
// ---------------------------------------------------------------------------------------

// Run is the task which writes the batch points after a certain amount of time.
func (b *Batch) Run() {
	b.timer = time.NewTimer(b.Timeout)
	for range b.timer.C {
		b.write()
		b.timer.Reset(b.Timeout)
	}
}

// Add inserts a new point into this batch.
func (b *Batch) Add(timestamp time.Time, tags Tags, values Values) error {
	if len(values) < 1 {
		return nil
	}

	// construct a new batch if necessary
	if b.batch == nil {
		b.batch, _ = client.NewBatchPoints(client.BatchPointsConfig{
			Precision: "us",
			Database:  b.Database,
		})

		if b.timer != nil {
			b.timer.Reset(b.Timeout)
		}
	}

	// construct the new databpoint for influxdb
	pt, err := client.NewPoint(b.Measurement, tags, values, timestamp)
	if err != nil {
		return err
	}
	b.batch.AddPoint(pt)

	// don't write the point batch to influxdb until
	// the threshold size has been reached
	if b.Size > 0 && len(b.batch.Points()) < b.Size {
		return nil
	}

	return b.write()
}

// write writes this batch to influxdb.
func (b *Batch) write() error {
	b.Lock()
	defer b.Unlock()

	if b.batch == nil {
		return nil
	}

	err := b.Influx.Write(b.batch)
	if err != nil {
		return err
	}
	b.batch = nil

	return nil
}
