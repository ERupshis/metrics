package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_createJSONGaugeMessage(t *testing.T) {
	type args struct {
		name  string
		value float64
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "valid case",
			args: args{"asd", 123},
			want: []byte("{\"id\":\"asd\",\"type\":\"gauge\",\"value\":123}"),
		},
		{
			name: "valid case without value",
			args: args{name: "asd"},
			want: []byte("{\"id\":\"asd\",\"type\":\"gauge\",\"value\":0}"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.want, string(createJSONGaugeMessage(tt.args.name, tt.args.value)))
		})
	}
}

func Test_createJSONCounterMessage(t *testing.T) {
	type args struct {
		name  string
		value int64
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "valid case",
			args: args{"asd", 123},
			want: []byte("{\"id\":\"asd\",\"type\":\"counter\",\"delta\":123}"),
		},
		{
			name: "valid case without value",
			args: args{name: "asd"},
			want: []byte("{\"id\":\"asd\",\"type\":\"counter\",\"delta\":0}"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.want, string(createJSONCounterMessage(tt.args.name, tt.args.value)))
		})
	}
}
