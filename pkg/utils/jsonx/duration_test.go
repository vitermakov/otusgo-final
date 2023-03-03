package jsonx

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseDuration(t *testing.T) {
	testCases := []struct {
		input    string
		expected Duration
		isErr    bool
		duration time.Duration
	}{
		{
			input:    "3er",
			expected: Duration{},
			isErr:    true,
		}, {
			input:    "b3",
			expected: Duration{},
			isErr:    true,
		}, {
			input:    "3y6",
			expected: Duration{},
			isErr:    true,
		}, {
			input:    "-1y",
			expected: Duration{},
			isErr:    true,
		}, {
			input:    "m",
			expected: Duration{},
			isErr:    true,
		}, {
			input:    "5s",
			expected: Duration{},
			isErr:    false,
			duration: time.Second * 5,
		}, {
			input:    "50",
			expected: Duration{},
			isErr:    false,
			duration: time.Second * 50,
		}, {
			input:    "5m",
			expected: Duration{},
			isErr:    false,
			duration: time.Minute * 5,
		}, {
			input:    "5h",
			expected: Duration{},
			isErr:    false,
			duration: time.Hour * 5,
		}, {
			input:    "5d",
			expected: NewDuration(5, 'd'),
			isErr:    false,
			duration: time.Hour * 24 * 5,
		}, {
			input:    "5w",
			expected: NewDuration(5, 'w'),
			isErr:    false,
			duration: time.Hour * 24 * 7 * 5,
		}, {
			input:    "5n",
			expected: NewDuration(5, 'n'),
			isErr:    false,
			duration: time.Hour * 24 * 30 * 5,
		}, {
			input:    "5y",
			expected: NewDuration(5, 'y'),
			isErr:    false,
			duration: time.Hour * 24 * 365 * 5,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("test %d", i+1), func(t *testing.T) {
			d, err := ParseDuration(tc.input)
			if tc.isErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				ds, _ := d.AsDuration()
				require.Equal(t, tc.duration, ds, "(%d, %s)", d.value, string(d.unit))
			}
		})
	}
}
