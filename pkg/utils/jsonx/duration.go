package jsonx

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

const QuotesByte = 34

type Duration struct {
	value int
	unit  byte
}

func NewDuration(value int, unit byte) Duration {
	return Duration{value, unit}
}

var allowedDs = map[byte]time.Duration{
	's': time.Second,
	'm': time.Minute,
	'h': time.Hour,
	'd': time.Hour * 24,
	'w': time.Hour * 24 * 7,
	'n': time.Hour * 24 * 30,
	'y': time.Hour * 24 * 365,
}

// UnmarshalJSON разрешенный формат <duration int >= 0><unit rune>. Allowed units:
// - s second
// - m minute
// - h hour
// - d day
// - w week
// - n 30 days
// - y 365 days.
func (d *Duration) UnmarshalJSON(data []byte) error {
	var err error
	if data[0] == QuotesByte {
		*d, err = ParseDuration(string(data[1 : len(data)-1]))
	} else {
		*d, err = ParseDuration(string(data))
	}
	return err
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return []byte(d.String()), nil
}

func (d Duration) Valid() bool {
	_, ok := allowedDs[d.unit]
	return ok
}

func (d Duration) String() string {
	return fmt.Sprintf("%d%s", d.value, string(d.unit))
}

func (d Duration) AsDuration() (time.Duration, error) {
	if !d.Valid() {
		return time.Second * 0, errors.New("not valid duration")
	}
	return time.Duration(d.value) * allowedDs[d.unit], nil
}

func ParseDuration(data string) (Duration, error) {
	var (
		s = make([]byte, 0, len(data))
		u = make([]byte, 0, len(data))
	)
	var unit bool
	for _, ch := range data {
		if unit {
			u = append(u, byte(ch))
		} else {
			if ch >= '0' && ch <= '9' {
				s = append(s, byte(ch))
			} else {
				u = append(u, byte(ch))
				unit = true
			}
		}
	}
	if len(s) == 0 {
		return Duration{}, fmt.Errorf("no value: %s", data)
	}
	n, err := strconv.Atoi(string(s))
	if err != nil || n < 0 {
		return Duration{}, fmt.Errorf("wrong value: %s", s)
	}
	if len(u) == 0 {
		u = append(u, 's')
	}
	if len(u) != 1 {
		return Duration{}, fmt.Errorf("wrong unit: %s", u)
	}
	if _, ok := allowedDs[u[0]]; !ok {
		return Duration{}, fmt.Errorf("unknown unit: %s", u)
	}
	return Duration{n, u[0]}, nil
}
