package confighelper

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetEnvToParamIfNeed(t *testing.T) {
	var intVal int64
	var intWant int64
	tests := []struct {
		name string
		arg  string
		want string
	}{
		{
			name: "valid int64",
			arg:  "123",
			want: "123",
		},
	}
	for _, tt := range tests {
		intVal = 0
		intRes, _ := strconv.Atoi(tt.want)
		intWant = int64(intRes)

		t.Run(tt.name, func(t *testing.T) {
			SetEnvToParamIfNeed(&intVal, tt.arg)
			assert.Equal(t, intVal, intWant)
		})
	}
}

func TestAtoi64(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "valid",
			args: args{
				"123",
			},
			want:    123,
			wantErr: assert.NoError,
		},
		{
			name: "incorrect type float",
			args: args{
				"123.3",
			},
			want:    0,
			wantErr: assert.Error,
		},
		{
			name: "incorrect type string",
			args: args{
				"asd",
			},
			want:    0,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Atoi64(tt.args.value)
			if !tt.wantErr(t, err, fmt.Sprintf("Atoi64(%v)", tt.args.value)) {
				return
			}
			assert.Equalf(t, tt.want, got, "Atoi64(%v)", tt.args.value)
		})
	}
}

func TestAddHTTPPrefixIfNeed(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "without prefix",
			args: args{
				"asd.asd",
			},
			want: "http://asd.asd",
		},
		{
			name: "with prefix",
			args: args{
				"http://asd.asd",
			},
			want: "http://asd.asd",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, AddHTTPPrefixIfNeed(tt.args.value), "AddHTTPPrefixIfNeed(%v)", tt.args.value)
		})
	}
}
