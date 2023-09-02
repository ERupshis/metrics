package confighelper

import (
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
