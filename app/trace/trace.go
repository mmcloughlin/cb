// Package trace manipulates benchmark result timeseries.
package trace

import (
	"bufio"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/mmcloughlin/cb/internal/errutil"
)

// Point represents a point in a collection of benchmark timeseries.
type Point struct {
	BenchmarkUUID   uuid.UUID
	EnvironmentUUID uuid.UUID
	CommitIndex     int
	CommitTime      time.Time
	Value           float64
}

func (p *Point) key() string {
	return fmt.Sprintf(",benchmark_uuid=%x,environment_uuid=%x,", p.BenchmarkUUID[:], p.EnvironmentUUID[:])
}

// WritePoints writes supplied points to w.
func WritePoints(w io.Writer, ps []Point) error {
	z := gzip.NewWriter(w)
	for _, p := range ps {
		_, err := fmt.Fprintf(z, "%s,%s,%d,%d,%v\n",
			p.BenchmarkUUID,
			p.EnvironmentUUID,
			p.CommitIndex,
			p.CommitTime.UnixNano(),
			p.Value,
		)
		if err != nil {
			return err
		}
	}
	return z.Close()
}

// ReadPoints from r.
func ReadPoints(r io.Reader) ([]Point, error) {
	z, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	s := bufio.NewScanner(z)
	ps := []Point{}
	for s.Scan() {
		line := s.Text()
		parts := strings.Split(line, ",")
		if len(parts) != 5 {
			return nil, errors.New("expected five fields")
		}

		var p Point
		p.BenchmarkUUID, err = uuid.Parse(parts[0])
		if err != nil {
			return nil, err
		}

		p.EnvironmentUUID, err = uuid.Parse(parts[1])
		if err != nil {
			return nil, err
		}

		p.CommitIndex, err = strconv.Atoi(parts[2])
		if err != nil {
			return nil, err
		}

		unixNano, err := strconv.ParseInt(parts[3], 10, 64)
		if err != nil {
			return nil, err
		}
		p.CommitTime = time.Unix(0, unixNano)

		p.Value, err = strconv.ParseFloat(parts[4], 64)
		if err != nil {
			return nil, err
		}

		ps = append(ps, p)
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	if err := z.Close(); err != nil {
		return nil, err
	}
	return ps, nil
}

// WritePointsFile writes points ps to filename.
func WritePointsFile(filename string, ps []Point) error {
	return writePointsFile(filename, ps)
}

func writePointsFile(filename string, ps []Point) (err error) {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer errutil.CheckClose(&err, f)
	return WritePoints(f, ps)
}

// ReadPointsFile reads points from filename.
func ReadPointsFile(filename string) ([]Point, error) {
	return readPointsFile(filename)
}

func readPointsFile(filename string) (_ []Point, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer errutil.CheckClose(&err, f)
	return ReadPoints(f)
}