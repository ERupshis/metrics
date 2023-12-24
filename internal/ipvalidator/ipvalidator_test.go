package ipvalidator

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatorIP_ValidateIPHandler(t *testing.T) {
	type fields struct {
		subnet string
	}
	type args struct {
		IP string
	}
	type want struct {
		statusCode      int
		errorCIDRParser bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "valid IP from client",
			fields: fields{
				subnet: "192.168.0.1/32",
			},
			args: args{
				IP: "192.168.0.1",
			},
			want: want{
				statusCode:      http.StatusOK,
				errorCIDRParser: false,
			},
		},
		{
			name: "invalid IP from client",
			fields: fields{
				subnet: "192.168.0.1/32",
			},
			args: args{
				IP: "192.168.0.0",
			},
			want: want{
				statusCode:      http.StatusForbidden,
				errorCIDRParser: false,
			},
		},
		{
			name: "incorrect mask size for server",
			fields: fields{
				subnet: "192.168.0.1/33",
			},
			args: args{
				IP: "192.168.0.0",
			},
			want: want{
				statusCode:      http.StatusOK,
				errorCIDRParser: true,
			},
		},
		{
			name: "missing ip in client's request",
			fields: fields{
				subnet: "192.168.0.1/24",
			},
			args: args{
				IP: "",
			},
			want: want{
				statusCode:      http.StatusForbidden,
				errorCIDRParser: false,
			},
		},
		{
			name: "subnet is not set for server",
			fields: fields{
				subnet: "",
			},
			args: args{
				IP: "192.168.0.0",
			},
			want: want{
				statusCode:      http.StatusOK,
				errorCIDRParser: true,
			},
		},
		{
			name: "incorrect IP in request",
			fields: fields{
				subnet: "192.168.0.1/32",
			},
			args: args{
				IP: "192.1.1",
			},
			want: want{
				statusCode:      http.StatusForbidden,
				errorCIDRParser: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, subnet, err := net.ParseCIDR(tt.fields.subnet)
			if (err != nil) != tt.want.errorCIDRParser {
				t.Errorf("ParseCIDR() error = %v, wantErr %v", err, tt.want.errorCIDRParser)
				return
			}
			err = nil

			v := &ValidatorIP{
				trustedSubnet: subnet,
			}

			req := httptest.NewRequest("POST", "/", nil)
			defer func() {
				_ = req.Body.Close()
			}()
			req.Header.Add("X-Real-IP", tt.args.IP)

			rr := httptest.NewRecorder()

			handler := v.ValidateIPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
		})
	}
}
