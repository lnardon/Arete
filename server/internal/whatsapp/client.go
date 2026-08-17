package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/lnardon/arete/internal/config"
)

type Client struct {
	baseURL  string
	apiKey   string
	instance string
	http     *http.Client
}

func NewClient(cfg config.EvolutionConfig) *Client {
	return &Client{
		baseURL:  cfg.BaseURL,
		apiKey:   cfg.APIKey,
		instance: cfg.InstanceName,
		http:     &http.Client{Timeout: 15 * time.Second},
	}
}

type sendTextRequest struct {
	Number string `json:"number"`
	Text   string `json:"text"`
}

func (c *Client) SendText(ctx context.Context, toPhone, text string) error {
	body, err := json.Marshal(sendTextRequest{Number: toPhone, Text: text})
	if err != nil {
		return fmt.Errorf("marshal send-text request: %w", err)
	}

	url := fmt.Sprintf("%s/message/sendText/%s", c.baseURL, c.instance)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build send-text request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("send-text request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("evolution api returned %d: %s", resp.StatusCode, respBody)
	}
	return nil
}

type getBase64MediaRequest struct {
	Message struct {
		Key struct {
			RemoteJid string `json:"remoteJid"`
			ID        string `json:"id"`
			FromMe    bool   `json:"fromMe"`
		} `json:"key"`
	} `json:"message"`
}

type MediaResult struct {
	Base64   string `json:"base64"`
	Mimetype string `json:"mimetype"`
}

func (c *Client) GetBase64Media(ctx context.Context, remoteJid, messageID string, fromMe bool) (*MediaResult, error) {
	var reqBody getBase64MediaRequest
	reqBody.Message.Key.RemoteJid = remoteJid
	reqBody.Message.Key.ID = messageID
	reqBody.Message.Key.FromMe = fromMe

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal get-media request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/getBase64FromMediaMessage/%s", c.baseURL, c.instance)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build get-media request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get-media request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("evolution api returned %d: %s", resp.StatusCode, respBody)
	}

	var result MediaResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode get-media response: %w", err)
	}
	if result.Base64 == "" {
		return nil, fmt.Errorf("evolution api returned empty media payload")
	}
	return &result, nil
}
