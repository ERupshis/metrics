package configutils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetEnvToParamIfNeed(t *testing.T) {
	// Test case 1: Set int64 parameter
	var intValue int64
	SetEnvToParamIfNeed(&intValue, "123")
	if intValue != 123 {
		t.Errorf("Expected intValue to be 123, got %d", intValue)
	}

	// Test case 2: Set string parameter
	var stringValue string
	SetEnvToParamIfNeed(&stringValue, "testString")
	if stringValue != "testString" {
		t.Errorf("Expected stringValue to be 'testString', got '%s'", stringValue)
	}

	// Test case 3: Empty value, should not modify parameters
	SetEnvToParamIfNeed(&intValue, "")
	if intValue != 123 {
		t.Errorf("Expected intValue to remain 123, got %d", intValue)
	}

	SetEnvToParamIfNeed(&stringValue, "")
	if stringValue != "testString" {
		t.Errorf("Expected stringValue to remain 'testString', got '%s'", stringValue)
	}

	// Test case 4: Wrong input type, should panic with an error message
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for wrong input type, but no panic occurred")
		} else {
			errMsg, ok := r.(error)
			if !ok || errMsg.Error() != "wrong input param type" {
				t.Errorf("Unexpected panic message: %v", r)
			}
		}
	}()

	SetEnvToParamIfNeed(42, "test")
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
	for _, ttCommon := range tests {
		tt := ttCommon
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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
	for _, ttCommon := range tests {
		tt := ttCommon
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equalf(t, tt.want, AddHTTPPrefixIfNeed(tt.args.value), "AddHTTPPrefixIfNeed(%v)", tt.args.value)
		})
	}
}
