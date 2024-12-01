package external

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"effective_mobile/internal/clients"
	"effective_mobile/internal/domain/models"
	"effective_mobile/internal/lib/logger/sl"
)

type Client struct {
	baseURL string
	log     *slog.Logger
	client  *http.Client
}

func New(log *slog.Logger, baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		log:     log,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) FetchSongDetails(group, song string) (*models.SongData, error) {
	const op = "clients.external.FetchSong"

	encodedGroup := url.QueryEscape(group)
	encodedSong := url.QueryEscape(song)

	url := fmt.Sprintf("%s/info?group=%s&song=%s", c.baseURL, encodedGroup, encodedSong)

	c.log.Info("fetching song details", slog.String("url", url))

	resp, err := c.client.Get(url)
	if err != nil {
		c.log.Error("failed to make get reqest", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusBadRequest {
			c.log.Error("get request status code: ", sl.Err(clients.ErrBadRequest))

			return nil, fmt.Errorf("%s: %w", op, clients.ErrBadRequest)
		}

		if resp.StatusCode == http.StatusInternalServerError {
			c.log.Error("get request status code: ", sl.Err(clients.ErrBadRequest))

			return nil, fmt.Errorf("%s: %w", op, clients.ErrInternal)
		}
	}

	var detail models.SongData
	if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
		c.log.Error("failed to decode external response", sl.Err(err))

		return nil, fmt.Errorf("%s: failed to decode external API response: %w", op, err)
	}

	return &detail, nil
}
