package logger

import (
	"bytes"
	"errors"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoggerLevel(t *testing.T) {
	testCases := []struct {
		value    string
		expected Level
		err      error
	}{
		{
			value:    "info",
			expected: LevelInfo,
		}, {
			value:    "warn",
			expected: LevelWarn,
		}, {
			value:    "error",
			expected: LevelError,
		}, {
			value:    "",
			expected: LevelNone,
			err:      ErrorUnknownLevel,
		}, {
			value:    "trace",
			expected: LevelNone,
			err:      ErrorUnknownLevel,
		}, {
			value:    "x",
			expected: LevelNone,
			err:      ErrorUnknownLevel,
		},
	}
	for _, tc := range testCases {
		t.Run("log level", func(t *testing.T) {
			level, err := ParseLevel(tc.value)
			require.Equal(t, level, tc.expected)
			if tc.err == nil {
				require.NoError(t, err)
			} else {
				require.True(t, errors.Is(err, tc.err))
			}
		})
	}
}

func TestLogrusAdapter(t *testing.T) {
	file, err := os.CreateTemp("", "log.")
	require.NoError(t, err)
	defer file.Close()
	defer os.Remove(file.Name())

	t.Run("wrong level", func(t *testing.T) {
		_, err := NewLogrus(Config{
			Level:     555,
			FileName:  file.Name(),
			IsTesting: true,
		})
		require.ErrorIs(t, err, ErrorUnknownLevel)
	})
	t.Run("wrong file name", func(t *testing.T) {
		_, err := NewLogrus(Config{
			Level:     LevelInfo,
			FileName:  "abra/cadabra",
			IsTesting: true,
		})
		require.ErrorIs(t, err, ErrorOpenLogFile)
	})
	t.Run("logger ok", func(t *testing.T) {
		logger, err := NewLogrus(Config{
			Level:     LevelError,
			FileName:  file.Name(),
			IsTesting: true,
		})
		require.NoError(t, err)

		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer func() {
			log.SetOutput(os.Stderr)
		}()

		logger.Debug("debug")
		logger.Info("info")
		logger.Error("error")
		logger.Warn("warn")
		logger.Fatal("fatal")

		s, err := os.ReadFile(file.Name())
		logged := string(s)
		require.NoError(t, err)
		require.NotContains(t, logged, "debug")
		require.NotContains(t, logged, "info")
		require.NotContains(t, logged, "warn")
		require.Contains(t, logged, "error")
		require.Contains(t, logged, "fatal")
		require.Contains(t, buf.String(), "exit")
	})
}
