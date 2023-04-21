package chunkmanager

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"simple-storage/internal/utils"
	"strings"
)

type httpClient interface {
	Do(req *http.Request) (resp *http.Response, err error)
}

type Client struct {
	log     *log.Logger
	address string
	client  httpClient
}

func New(log *log.Logger, address string, httpClient httpClient) *Client {
	log = utils.LoggerExtendWithPrefix(log, "chunk-manager-client ->")

	return &Client{
		log:     log,
		client:  httpClient,
		address: address,
	}
}

func (c *Client) RegisterStorageServer(address string) error {
	url := fmt.Sprintf("http://%s/register", c.address)
	body := strings.NewReader(address)

	req, err := http.NewRequestWithContext(context.Background(), "POST", url, body)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(
			fmt.Sprintf("status code: %d %s", resp.StatusCode, resp.Status))
	}

	return nil
}
