package retryer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/erupshis/metrics/internal/logger"
	"github.com/jackc/pgerrcode"
	"github.com/stretchr/testify/assert"
)

var databaseErrorsToRetry = []error{
	errors.New(pgerrcode.UniqueViolation),
	errors.New(pgerrcode.ConnectionException),
	errors.New(pgerrcode.ConnectionDoesNotExist),
	errors.New(pgerrcode.ConnectionFailure),
	errors.New(pgerrcode.SQLClientUnableToEstablishSQLConnection),
	errors.New(pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection),
	errors.New(pgerrcode.TransactionResolutionUnknown),
	errors.New(pgerrcode.ProtocolViolation),
}

func Test_canRetryCall(t *testing.T) {
	type args struct {
		err              error
		repeatableErrors []error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "valid",
			args: args{
				err:              errors.New(`08000`),
				repeatableErrors: databaseErrorsToRetry,
			},
			want: true,
		},
		{
			name: "valid with missing slice",
			args: args{
				err:              errors.New(`any error`),
				repeatableErrors: nil,
			},
			want: true,
		},
		{
			name: "invalid error is not in slice",
			args: args{
				err:              errors.New(`any error`),
				repeatableErrors: databaseErrorsToRetry,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, canRetryCall(tt.args.err, tt.args.repeatableErrors))
		})
	}
}

func TestRetryCallWithTimeout(t *testing.T) {
	type args struct {
		ctx              context.Context
		log              logger.BaseLogger
		intervals        []int
		repeatableErrors []error
		callback         func(context.Context) error
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "valid",
			args: args{
				ctx:              context.Background(),
				log:              logger.CreateLogger("Info"),
				intervals:        []int{1, 1, 1},
				repeatableErrors: nil,
				callback: func(ctx context.Context) error {
					currentTime := time.Now()
					for {
						select {
						case <-ctx.Done():
							return errors.New(pgerrcode.ConnectionException)
						default:
							if time.Since(currentTime) > 10*time.Second {
								return nil
							}
							time.Sleep(50 * time.Millisecond)
						}
					}
				},
			},
			wantErr: errors.New(pgerrcode.ConnectionException),
		},
		{
			name: "valid with success",
			args: args{
				ctx:              context.Background(),
				log:              logger.CreateLogger("Info"),
				intervals:        nil,
				repeatableErrors: nil,
				callback: func(ctx context.Context) error {
					return nil
				},
			},
			wantErr: nil,
		},
		{
			name: "valid should retry",
			args: args{
				ctx:              context.Background(),
				log:              logger.CreateLogger("Info"),
				intervals:        []int{1, 1, 1},
				repeatableErrors: databaseErrorsToRetry,
				callback: func(ctx context.Context) error {
					currentTime := time.Now()
					for {
						select {
						case <-ctx.Done():
							return errors.New(pgerrcode.ConnectionException)
						default:
							if time.Since(currentTime) > 10*time.Second {
								return nil
							}
							time.Sleep(50 * time.Millisecond)
						}
					}
				},
			},
			wantErr: errors.New(pgerrcode.ConnectionException),
		},
		{
			name: "valid shouldn't retry",
			args: args{
				ctx:              context.Background(),
				log:              logger.CreateLogger("Info"),
				intervals:        []int{1, 1, 1},
				repeatableErrors: databaseErrorsToRetry,
				callback: func(ctx context.Context) error {
					currentTime := time.Now()
					for {
						select {
						case <-ctx.Done():
							return errors.New("some error")
						default:
							if time.Since(currentTime) > 10*time.Second {
								return nil
							}
							time.Sleep(50 * time.Millisecond)
						}
					}
				},
			},
			wantErr: errors.New("some error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantErr, RetryCallWithTimeout(tt.args.ctx, tt.args.log, tt.args.intervals, tt.args.repeatableErrors, tt.args.callback))
		})
	}
}
