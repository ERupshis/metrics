package client

type BaseClient interface {
	PostJSON(url string, body []byte) error
}
