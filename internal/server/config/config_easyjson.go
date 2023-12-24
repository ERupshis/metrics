// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package config

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
	time "time"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjson6615c02eDecodeGithubComErupshisMetricsInternalServerConfig1(in *jlexer.Lexer, out *envConfig) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "Host":
			out.Host = string(in.String())
		case "LogLevel":
			out.LogLevel = string(in.String())
		case "Restore":
			out.Restore = bool(in.Bool())
		case "StoragePath":
			out.StoragePath = string(in.String())
		case "StoreInterval":
			out.StoreInterval = string(in.String())
		case "DataBaseDSN":
			out.DataBaseDSN = string(in.String())
		case "Key":
			out.Key = string(in.String())
		case "KeyRSA":
			out.KeyRSA = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson6615c02eEncodeGithubComErupshisMetricsInternalServerConfig1(out *jwriter.Writer, in envConfig) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"Host\":"
		out.RawString(prefix[1:])
		out.String(string(in.Host))
	}
	{
		const prefix string = ",\"LogLevel\":"
		out.RawString(prefix)
		out.String(string(in.LogLevel))
	}
	{
		const prefix string = ",\"Restore\":"
		out.RawString(prefix)
		out.Bool(bool(in.Restore))
	}
	{
		const prefix string = ",\"StoragePath\":"
		out.RawString(prefix)
		out.String(string(in.StoragePath))
	}
	{
		const prefix string = ",\"StoreInterval\":"
		out.RawString(prefix)
		out.String(string(in.StoreInterval))
	}
	{
		const prefix string = ",\"DataBaseDSN\":"
		out.RawString(prefix)
		out.String(string(in.DataBaseDSN))
	}
	{
		const prefix string = ",\"Key\":"
		out.RawString(prefix)
		out.String(string(in.Key))
	}
	{
		const prefix string = ",\"KeyRSA\":"
		out.RawString(prefix)
		out.String(string(in.KeyRSA))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v envConfig) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson6615c02eEncodeGithubComErupshisMetricsInternalServerConfig1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v envConfig) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson6615c02eEncodeGithubComErupshisMetricsInternalServerConfig1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *envConfig) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson6615c02eDecodeGithubComErupshisMetricsInternalServerConfig1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *envConfig) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson6615c02eDecodeGithubComErupshisMetricsInternalServerConfig1(l, v)
}
func easyjson6615c02eDecodeGithubComErupshisMetricsInternalServerConfig2(in *jlexer.Lexer, out *Config) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "address":
			out.Host = string(in.String())
		case "log_level":
			out.LogLevel = string(in.String())
		case "restore":
			out.Restore = bool(in.Bool())
		case "store_interval":
			out.StoreInterval, _ = time.ParseDuration(in.String())
		case "storage_file":
			out.StoragePath = string(in.String())
		case "database_dsn":
			out.DataBaseDSN = string(in.String())
		case "hash_key":
			out.Key = string(in.String())
		case "crypto_key":
			out.KeyRSA = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson6615c02eEncodeGithubComErupshisMetricsInternalServerConfig2(out *jwriter.Writer, in Config) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"address\":"
		out.RawString(prefix[1:])
		out.String(string(in.Host))
	}
	{
		const prefix string = ",\"log_level\":"
		out.RawString(prefix)
		out.String(string(in.LogLevel))
	}
	{
		const prefix string = ",\"restore\":"
		out.RawString(prefix)
		out.Bool(bool(in.Restore))
	}
	{
		const prefix string = ",\"store_interval\":"
		out.RawString(prefix)
		out.String(string(in.StoreInterval.String()))
	}
	{
		const prefix string = ",\"storage_file\":"
		out.RawString(prefix)
		out.String(string(in.StoragePath))
	}
	{
		const prefix string = ",\"database_dsn\":"
		out.RawString(prefix)
		out.String(string(in.DataBaseDSN))
	}
	{
		const prefix string = ",\"hash_key\":"
		out.RawString(prefix)
		out.String(string(in.Key))
	}
	{
		const prefix string = ",\"crypto_key\":"
		out.RawString(prefix)
		out.String(string(in.KeyRSA))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Config) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson6615c02eEncodeGithubComErupshisMetricsInternalServerConfig2(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Config) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson6615c02eEncodeGithubComErupshisMetricsInternalServerConfig2(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Config) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson6615c02eDecodeGithubComErupshisMetricsInternalServerConfig2(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Config) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson6615c02eDecodeGithubComErupshisMetricsInternalServerConfig2(l, v)
}
