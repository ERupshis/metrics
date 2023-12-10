package configutils

import (
	"bufio"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

const testFolder = "/test"

type testConfig struct {
	Test string `json:"test"`
}

func TestParseConfigFromFile(t *testing.T) {
	type args struct {
		filePath     string
		pathToSeek   string
		fileData     string
		structToFill any
	}
	type want struct {
		fieldData string
		wantErr   assert.ErrorAssertionFunc
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "valid",
			args: args{
				filePath:     testFolder + "/test",
				pathToSeek:   testFolder + "/test",
				fileData:     `{"test": "test"}`,
				structToFill: &testConfig{},
			},
			want: want{
				wantErr:   assert.NoError,
				fieldData: "test",
			},
		},
		{
			name: "incorrect path",
			args: args{
				filePath:     testFolder + "/test2",
				pathToSeek:   testFolder + "/wrong",
				fileData:     `{"test": "test"}`,
				structToFill: &testConfig{},
			},
			want: want{
				wantErr:   assert.Error,
				fieldData: "",
			},
		},
		{
			name: "broken json",
			args: args{
				filePath:     testFolder + "/test3",
				pathToSeek:   testFolder + "/test3",
				fileData:     `{"test": "test"`,
				structToFill: &testConfig{},
			},
			want: want{
				wantErr:   assert.Error,
				fieldData: "",
			},
		},
	}
	for _, ttCommon := range tests {
		tt := ttCommon
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f, err := os.Create(tt.args.filePath)
			require.NoError(t, err)
			defer func() {
				_ = f.Close()
			}()

			defer func() {
				_ = os.RemoveAll(testFolder)
			}()

			w := bufio.NewWriter(f)
			_, err = w.Write([]byte(tt.args.fileData))
			require.NoError(t, err)
			err = w.Flush()
			require.NoError(t, err)

			tt.want.wantErr(t, ParseConfigFromFile(tt.args.pathToSeek, tt.args.structToFill),
				fmt.Sprintf("ParseConfigFromFile(%v, %v)", tt.args.pathToSeek, tt.args.structToFill))
			assert.Equal(t, tt.want.fieldData, tt.args.structToFill.(*testConfig).Test)
		})
	}
}

const testFolder1 = "/test1"

func TestCheckConfigFile(t *testing.T) {
	_ = os.RemoveAll(testFolder1)
	err := os.Mkdir(testFolder1, 0750)
	require.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(testFolder1)
	}()

	test1 := "/test1"
	test2 := "/test2"
	test3 := "/test3"

	type args struct {
		filePath    string
		fileData    string
		config      any
		initializer func()
	}
	type want struct {
		fieldData string
		wantErr   assert.ErrorAssertionFunc
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "without env",
			args: args{
				config: &testConfig{},
				initializer: func() {

				},
				filePath: testFolder1 + test3,
				fileData: `{"test": "test"}`,
			},
			want: want{
				fieldData: "",
				wantErr:   assert.NoError,
			},
		},
		{
			name: "valid env",
			args: args{
				config: &testConfig{},
				initializer: func() {
					_ = os.Setenv("CONFIG", testFolder1+test1)
				},
				filePath: testFolder1 + test1,
				fileData: `{"test": "test"}`,
			},
			want: want{
				fieldData: "test",
				wantErr:   assert.NoError,
			},
		},
		{
			name: "incorrect file path",
			args: args{
				config: &testConfig{},
				initializer: func() {
					_ = os.Setenv("CONFIG", testFolder1+"/wrong")
				},
				filePath: testFolder1 + test2,
				fileData: `{"test": "test"}`,
			},
			want: want{
				fieldData: "",
				wantErr:   assert.Error,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.initializer()

			f, err := os.Create(tt.args.filePath)
			require.NoError(t, err)
			defer func() {
				_ = f.Close()
			}()

			w := bufio.NewWriter(f)
			_, err = w.Write([]byte(tt.args.fileData))
			require.NoError(t, err)
			err = w.Flush()
			require.NoError(t, err)

			tt.want.wantErr(t, CheckConfigFile(tt.args.config), fmt.Sprintf("CheckConfigFile(%v)", tt.args.config))
			assert.Equal(t, tt.want.fieldData, tt.args.config.(*testConfig).Test)
		})
	}
}
