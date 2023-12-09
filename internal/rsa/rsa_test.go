package rsa

import (
	"bytes"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/erupshis/metrics/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	certRSA = "../../../rsa/cert.pem"
	keyRSA  = "../../../rsa/key.pem"
)

func TestCreateEncoder(t *testing.T) {
	type args struct {
		certFilePath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid path for cert",
			args: args{
				certFilePath: certRSA,
			},
			wantErr: false,
		},
		{
			name: "incorrect path for cert",
			args: args{
				certFilePath: "../../../cert.pem",
			},
			wantErr: true,
		},
		{
			name: "incorrect private key instead of cert",
			args: args{
				certFilePath: keyRSA,
			},
			wantErr: true,
		},
	}
	for _, ttCommon := range tests {
		tt := ttCommon
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := CreateEncoder(tt.args.certFilePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateEncoder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestCreateDecoder(t *testing.T) {
	type args struct {
		keyFilePath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid path for cert",
			args: args{
				keyFilePath: keyRSA,
			},
			wantErr: false,
		},
		{
			name: "incorrect path for cert",
			args: args{
				keyFilePath: "../../../key.pem",
			},
			wantErr: true,
		},
		{
			name: "incorrect cert instead of private key",
			args: args{
				keyFilePath: certRSA,
			},
			wantErr: true,
		},
	}
	for _, ttCommon := range tests {
		tt := ttCommon
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := CreateDecoder(tt.args.keyFilePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateDecoder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestEncoder_Encode(t *testing.T) {
	encoder, err := CreateEncoder(certRSA)
	require.NoError(t, err)
	decoder, err := CreateDecoder(keyRSA)
	require.NoError(t, err)

	type fields struct {
		keyPublic  *rsa.PublicKey
		keyPrivate *rsa.PrivateKey
	}
	type args struct {
		msg []byte
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wantErrEncoder bool
		wantErrDecoder bool
	}{
		{
			name: "valid",
			fields: fields{
				keyPublic:  encoder.key,
				keyPrivate: decoder.key,
			},
			args: args{
				msg: []byte("some message"),
			},
		},
		{
			name: "empty message",
			fields: fields{
				keyPublic:  encoder.key,
				keyPrivate: decoder.key,
			},
			args: args{
				msg: []byte{},
			},
		},
		{
			name: "missing public key",
			fields: fields{
				keyPublic:  nil,
				keyPrivate: decoder.key,
			},
			args: args{
				msg: nil,
			},
			wantErrEncoder: true,
			wantErrDecoder: true,
		},
		{
			name: "missing private key",
			fields: fields{
				keyPublic:  encoder.key,
				keyPrivate: nil,
			},
			args: args{
				msg: nil,
			},
			wantErrDecoder: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Encoder{
				key: tt.fields.keyPublic,
			}
			msgEncoded, err := e.Encode(tt.args.msg)
			if (err != nil) != tt.wantErrEncoder {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErrEncoder)
				return
			}

			de := &Decoder{
				key: tt.fields.keyPrivate,
			}
			msgDecoded, err := de.Decode(msgEncoded)
			if (err != nil) != tt.wantErrDecoder {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErrDecoder)
				return
			}

			assert.Equal(t, tt.args.msg, msgDecoded)
		})
	}
}

func TestDecoder_DecodeRSAHandler(t *testing.T) {
	encoder, err := CreateEncoder(certRSA)
	require.NoError(t, err)
	decoder, err := CreateDecoder(keyRSA)
	require.NoError(t, err)

	type fields struct {
		log        logger.BaseLogger
		keyPublic  *rsa.PublicKey
		keyPrivate *rsa.PrivateKey
	}
	type args struct {
		body []byte
	}
	type want struct {
		statusCode int
		errEncoder bool
		errDecoder bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "common case",
			fields: fields{
				log:        logger.CreateMock(),
				keyPublic:  encoder.key,
				keyPrivate: decoder.key,
			},
			args: args{
				body: []byte("{\"some message text\"}"),
			},
			want: want{
				statusCode: http.StatusOK,
				errEncoder: false,
				errDecoder: false,
			},
		},
		{
			name: "empty message",
			fields: fields{
				log:        logger.CreateMock(),
				keyPublic:  encoder.key,
				keyPrivate: decoder.key,
			},
			args: args{
				body: []byte{},
			},
			want: want{
				statusCode: http.StatusOK,
				errEncoder: false,
				errDecoder: false,
			},
		},
		{
			name: "empty public key",
			fields: fields{
				log:        logger.CreateMock(),
				keyPublic:  nil,
				keyPrivate: decoder.key,
			},
			args: args{
				body: []byte{},
			},
			want: want{
				statusCode: http.StatusBadRequest,
				errEncoder: true,
				errDecoder: false,
			},
		},
		{
			name: "empty private key",
			fields: fields{
				log:        logger.CreateMock(),
				keyPublic:  encoder.key,
				keyPrivate: nil,
			},
			args: args{
				body: []byte{},
			},
			want: want{
				statusCode: http.StatusInternalServerError,
				errEncoder: false,
				errDecoder: false,
			},
		},
	}
	for _, ttCommon := range tests {
		tt := ttCommon
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := &Encoder{
				key: tt.fields.keyPublic,
			}

			msgEncoded, err := e.Encode(tt.args.body)
			if (err != nil) != tt.want.errEncoder {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.want.errEncoder)
				return
			}
			err = nil

			req := httptest.NewRequest("POST", "/", bytes.NewBuffer(msgEncoded))
			defer func() {
				_ = req.Body.Close()
			}()

			rr := httptest.NewRecorder()

			de := &Decoder{
				key: tt.fields.keyPrivate,
			}
			handler := de.DecodeRSAHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err = w.Write([]byte("correct"))
			}))

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			require.NoError(t, err)
		})
	}
}
