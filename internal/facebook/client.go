package facebook

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// FBPost é o post que chega do Graph API (forma mínima que nos interessa).
type FBPost struct {
	ID          string `json:"id"`
	Message     string `json:"message,omitempty"`
	Story       string `json:"story,omitempty"`
	CreatedTime string `json:"created_time,omitempty"`
}

// FBComment é o comentário retornado pela Graph API.
type FBComment struct {
	ID          string `json:"id"`
	Message     string `json:"message"`
	From        struct {
		Name string `json:"name"`
		ID   string `json:"id"`
	} `json:"from"`
	CreatedTime string `json:"created_time"`
}

// Client é a interface que o app usa. Trocar entre mock e Graph é só
// escolher a implementação.
type Client interface {
	ListPosts(ctx context.Context, pageAccessToken string, limit int) ([]FBPost, error)
	ListComments(ctx context.Context, postID, pageAccessToken string, limit int) ([]FBComment, error)
	Reply(ctx context.Context, commentID, message, pageAccessToken string) (string, error)
}

// GraphClient implementa Client usando a Meta Graph API.
type GraphClient struct {
	HTTP    *http.Client
	APIBase string // default: https://graph.facebook.com/v20.0
	PageID  string
}

// NewGraphClient cria cliente HTTP com timeout razoável.
func NewGraphClient(pageID string) *GraphClient {
	return &GraphClient{
		HTTP:    &http.Client{Timeout: 20 * time.Second},
		APIBase: "https://graph.facebook.com/v20.0",
		PageID:  pageID,
	}
}

func (g *GraphClient) endpoint(path string) string {
	return g.APIBase + path
}

// ListPosts — GET /{page-id}/posts?fields=id,message,story,created_time
func (g *GraphClient) ListPosts(ctx context.Context, pageAccessToken string, limit int) ([]FBPost, error) {
	if limit <= 0 {
		limit = 10
	}
	q := url.Values{}
	q.Set("fields", "id,message,story,created_time")
	q.Set("limit", fmt.Sprintf("%d", limit))
	q.Set("access_token", pageAccessToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.endpoint("/"+g.PageID+"/posts"), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.URL.RawQuery = q.Encode()
	var out struct {
		Data []FBPost `json:"data"`
	}
	if err := doJSON(req, g.HTTP, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

// ListComments — GET /{post-id}/comments?fields=id,message,from,created_time
func (g *GraphClient) ListComments(ctx context.Context, postID, pageAccessToken string, limit int) ([]FBComment, error) {
	if limit <= 0 {
		limit = 25
	}
	q := url.Values{}
	q.Set("fields", "id,message,from,created_time")
	q.Set("limit", fmt.Sprintf("%d", limit))
	q.Set("access_token", pageAccessToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.endpoint("/"+postID+"/comments"), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.URL.RawQuery = q.Encode()
	var out struct {
		Data []FBComment `json:"data"`
	}
	if err := doJSON(req, g.HTTP, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

// Reply — POST /{comment-id}/comments  body: message=...&access_token=...
func (g *GraphClient) Reply(ctx context.Context, commentID, message, pageAccessToken string) (string, error) {
	form := url.Values{}
	form.Set("message", message)
	form.Set("access_token", pageAccessToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.endpoint("/"+commentID+"/comments"), strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	var out struct {
		ID string `json:"id"`
	}
	if err := doJSON(req, g.HTTP, &out); err != nil {
		return "", err
	}
	return out.ID, nil
}

func doJSON(req *http.Request, hc *http.Client, out any) error {
	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		// Lê o corpo do erro pra mostrar a mensagem real da Meta
		// (em vez de só o status code). Limita a 512 chars pra não estourar o toast.
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		msg := strings.TrimSpace(string(body))
		if msg == "" {
			msg = "(sem corpo)"
		}
		return fmt.Errorf("graph %s %s -> %d: %s", req.Method, req.URL.Path, resp.StatusCode, msg)
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
