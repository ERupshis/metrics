package client

type BaseClient interface {
	PostJson(url string, body []byte) error
}
