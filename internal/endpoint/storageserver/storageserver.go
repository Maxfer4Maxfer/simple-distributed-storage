package storageserver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"simple-storage/internal/utils"
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
	log = utils.LoggerExtendWithPrefix(log, "storage-server-client ->")

	return &Client{
		log:     log,
		client:  httpClient,
		address: address,
	}
}

func (c *Client) UploadChunk(chunkID string, buf []byte) error {
	url := fmt.Sprintf("http://%s", c.address)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fw, err := writer.CreateFormFile("chunk", chunkID)
	if err != nil {
		return err
	}

	_, err = io.Copy(fw, bytes.NewReader(buf))
	if err != nil {
		return err
	}

	writer.Close()

	req, err := http.NewRequestWithContext(context.Background(), "PUT", url, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

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

func (c *Client) DownloadChunk(chunkID string, buf []byte) error {
	url := fmt.Sprintf("http://%s/?id=%s", c.address, chunkID)

	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
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

	defer resp.Body.Close()

	_, err = resp.Body.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	return nil
}
