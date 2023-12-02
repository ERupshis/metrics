package workers

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/erupshis/metrics/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Example() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	log := logger.CreateMock()

	workersPool, err := CreateWorkersPool(3, log)
	if err != nil {
		log.Info("failed to create workers.")
		return
	}
	defer workersPool.CloseJobsChan()
	defer workersPool.CloseResultsChan()

	jobsCount := 10
	wg := sync.WaitGroup{} // for test purpose
	wg.Add(10)

	jobClean := func() error {
		time.Sleep(time.Duration(rand.Intn(4)) * time.Second)
		return nil
	}

	jobWithError := func() error {
		time.Sleep(time.Duration(rand.Intn(4)) * time.Second)
		return fmt.Errorf("unexpected error")
	}

	jobs := []func() error{jobClean, jobWithError}

	go func() {
		for i := 0; i < jobsCount; i++ {
			workersPool.AddJob(jobs[rand.Intn(2)])
		}
	}()

	go func() {
		for res := range workersPool.GetResultChan() {
			if res != nil {
				log.Info("[WorkersPool] failed work: %v", res)
			}
			wg.Done()
		}
	}()

	wg.Wait()

	fmt.Println("1") // trigger for launch via go test
	// Output:
	// 1
}

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
				count: 2,
				log:   logger.CreateMock(),
			},
			want: want{
				createErr: false,
				jobsLen:   0,
				resLen:    0,
			},
		},
		{
			name: "incorrect workers count",
			args: args{
				count: 0,
				log:   logger.CreateMock(),
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
			assert.Equal(t, tt.want.jobsChClosed, ok)

			pool.CloseResultsChan()
			_, ok = <-pool.results
			assert.Equal(t, tt.want.jobsChClosed, ok)
		})
	}
}
