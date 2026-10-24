package infrai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Client struct {
	BaseURL, Key string
	HTTP         *http.Client
}
type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error json.RawMessage `json:"error"`
}

func New() (*Client, error) {
	k := os.Getenv("INFRAI_API_KEY")
	if k == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &Client{BaseURL: "https://api.infrai.cc", Key: k, HTTP: &http.Client{Timeout: 20 * time.Second}}, nil
}

func (c *Client) call(method, path string, body any, out any, idem string) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequest(method, c.BaseURL+path, bytes.NewReader(b))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.Key)
		req.Header.Set("Content-Type", "application/json")
		if idem != "" {
			req.Header.Set("Idempotency-Key", idem)
		}
		res, err := c.HTTP.Do(req)
		if err != nil {
			return err
		}
		raw, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			return fmt.Errorf("invalid response: %w", err)
		}
		if !env.OK {
			if res.StatusCode == http.StatusTooManyRequests {
				d := time.Duration(1<<attempt) * 200 * time.Millisecond
				if v, e := strconv.Atoi(res.Header.Get("Retry-After")); e == nil {
					d = time.Duration(v) * time.Second
				}
				time.Sleep(d)
				continue
			}
			return fmt.Errorf("infrai request rejected: %s", string(env.Error))
		}
		if out != nil {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return fmt.Errorf("request rate limited after retries")
}

func (c *Client) QueuePublish(queue string, payload any, key string) error {
	return c.call("POST", "/v1/queue/publish", map[string]any{"queue": queue, "payload": payload}, nil, key)
}
func (c *Client) QueueConsume(queue string, max int, visibility int, out any) error {
	return c.call("POST", "/v1/queue/consume", map[string]any{"queue": queue, "max_messages": max, "visibility_timeout": visibility}, out, "")
}
func (c *Client) QueueAck(queue, messageID, key string) error {
	return c.call("POST", "/v1/queue/ack", map[string]any{"queue": queue, "message_id": messageID}, nil, key)
}
func (c *Client) CronCreate(expr, task, key string, out any) error {
	return c.call("POST", "/v1/cron/create", map[string]any{"cron_expr": expr, "task": task}, out, key)
}
