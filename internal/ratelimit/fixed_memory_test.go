package ratelimit

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFixedMemory(t *testing.T) {
	t.Run("new_limiter", func(t *testing.T) {
		limiter := NewFixedMemory()
		require.Equal(t, 0, limiter.Capacity())
		require.Equal(t, []string{}, limiter.BucketNames())
	})

	t.Run("add_destroy", func(t *testing.T) {
		var (
			ok  bool
			err error
		)
		limiter := NewFixedMemory()
		opts := Limits{Period: time.Second, Limit: 10}
		keys := []string{"key1", "key2", "key3", "key4"}
		now := time.Now()
		for _, key := range keys {
			ok, err = limiter.ExceedLimit(key, opts)
			require.False(t, ok)
			require.NoError(t, err)
		}

		err = limiter.Destroy()
		require.NoError(t, err)
		require.Equal(t, []string{}, limiter.BucketNames())
		// данная проверка показывает, что лимитер закрывает все бакеты именно внешним сигналом.
		// значение 20мс взято от природы, но должно подходить
		require.Less(t, time.Since(now), time.Millisecond*20)
	})

	t.Run("add_auto_remove", func(t *testing.T) {
		var (
			ok  bool
			err error
		)
		limiter := NewFixedMemory()
		for i := 1; i <= 4; i++ {
			opts := Limits{Period: time.Millisecond * 200 * time.Duration(i), Limit: 10}
			ok, err = limiter.ExceedLimit(fmt.Sprintf("key%d", i), opts)
			require.False(t, ok)
			require.NoError(t, err)
		}
		require.ElementsMatch(t, []string{"key1", "key2", "key3", "key4"}, limiter.BucketNames())

		<-time.After(time.Millisecond * 1000)
		require.ElementsMatch(t, []string{"key3", "key4"}, limiter.BucketNames())
		<-time.After(time.Millisecond * 1000)
		require.ElementsMatch(t, []string{}, limiter.BucketNames())

		err = limiter.Destroy()
		require.NoError(t, err)
		require.Equal(t, 0, limiter.Capacity())
	})

	t.Run("add_manual_remove", func(t *testing.T) {
		var (
			ok  bool
			err error
		)
		limiter := NewFixedMemory()
		for i := 1; i <= 4; i++ {
			opts := Limits{Period: time.Second, Limit: 10}
			ok, err = limiter.ExceedLimit(fmt.Sprintf("key%d", i), opts)
			require.False(t, ok)
			require.NoError(t, err)
		}
		require.ElementsMatch(t, []string{"key1", "key2", "key3", "key4"}, limiter.BucketNames())

		_, err = limiter.ResetBucket("key2")
		require.NoError(t, err)
		_, err = limiter.ResetBucket("key4")
		require.NoError(t, err)
		require.ElementsMatch(t, []string{"key1", "key3"}, limiter.BucketNames())

		err = limiter.Destroy()
		require.NoError(t, err)
		require.Equal(t, 0, limiter.Capacity())
	})

	t.Run("check_limits", func(t *testing.T) {
		var (
			ok          bool
			err         error
			limit       = 10
			bucketCodeF = "key%d"
		)

		limiter := NewFixedMemory()
		opts := Limits{Period: time.Second, Limit: int64(limit)}

		for i := 1; i <= 2; i++ {
			bucketCode := fmt.Sprintf(bucketCodeF, i)
			for j := 1; j <= limit+1; j++ {
				ok, err = limiter.ExceedLimit(bucketCode, opts)
				require.NoError(t, err)
				if j > limit {
					require.True(t, ok)
				} else {
					require.False(t, ok)
				}
				<-time.After(time.Millisecond * 10)
			}
		}

		err = limiter.Destroy()
		require.NoError(t, err)
		require.Equal(t, 0, limiter.Capacity())
	})
}
