package workers

import (
	"testing"
	"time"

	"github.com/erupshis/metrics/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateWorkersPool(t *testing.T) {
	type args struct {
		count int64
		log   logger.BaseLogger
		jobs  []func() error
	}
	type want struct {
		createErr    bool
		jobsLen      int
		jobsChClosed bool
		resLen       int
		resChClosed  bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "valid",
			args: args{
				count: 3,
				log:   logger.CreateLogger("info"),
				jobs: []func() error{
					func() error {
						return nil
					},
					func() error {
						return nil
					},
					func() error {
						return nil
					},
				},
			},
			want: want{
				createErr:    false,
				jobsLen:      0,
				jobsChClosed: true,
				resLen:       3,
				resChClosed:  false,
			},
		},
		{
			name: "valid",
			args: args{
				count: 0,
				log:   logger.CreateLogger("info"),
			},
			want: want{
				createErr: true,
				jobsLen:   0,
				resLen:    3,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool, err := CreateWorkersPool(tt.args.count, tt.args.log)
			if pool == nil && !tt.want.createErr {
				t.Errorf("CreateWorkersPool() = %v", pool)
			}

			require.Equal(t, tt.want.createErr, err != nil)
			if tt.want.createErr {
				return
			}

			for _, job := range tt.args.jobs {
				pool.AddJob(job)
			}

			assert.Equal(t, tt.want.jobsLen, len(pool.jobs))
			time.Sleep(time.Millisecond)
			assert.Equal(t, tt.want.resLen, len(pool.results))

			pool.CloseJobsChan()
			_, ok := <-pool.jobs
			assert.Equal(t, tt.want.jobsChClosed, !ok)

			pool.CloseResultsChan()
			_, ok = <-pool.results
			assert.Equal(t, tt.want.jobsChClosed, ok)
		})
	}
}
