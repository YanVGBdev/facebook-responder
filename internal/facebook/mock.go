package facebook

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

// MockClient implementa Client sem rede. Útil para o modo "usar_mock=true"
// do config (testes locais sem credenciais).
type MockClient struct {
	posts    []FBPost
	nextCID  int64
	postedTo map[string]string // commentID -> message publicado
}

// NewMockClient devolve um conjunto pequeno de posts e comentários fictícios.
func NewMockClient() *MockClient {
	m := &MockClient{
		postedTo: map[string]string{},
	}
	now := time.Now().UTC().Format(time.RFC3339)
	m.posts = []FBPost{
		{ID: "mock_post_1", Message: "🔥 Promoção de inauguração: 20% de desconto hoje!", CreatedTime: now},
		{ID: "mock_post_2", Message: "Nosso horário de atendimento é de seg a sex, 9h às 18h.", CreatedTime: now},
		{ID: "mock_post_3", Message: "Novo produto chegando — fique de olho!", CreatedTime: now},
	}
	return m
}

func (m *MockClient) ListPosts(_ context.Context, _ string, limit int) ([]FBPost, error) {
	if limit <= 0 || limit > len(m.posts) {
		limit = len(m.posts)
	}
	out := make([]FBPost, limit)
	copy(out, m.posts[:limit])
	return out, nil
}

func (m *MockClient) ListComments(_ context.Context, postID, _ string, limit int) ([]FBComment, error) {
	// Gera comentários fictícios variando por post.
	base := []string{
		"Adorei! Vocês entregam para minha cidade?",
		"Qual o preço?",
		"Top demais!",
		"Como faço para comprar?",
		"Tem versão em outras cores?",
		"Vocês têm loja física?",
		"Já comprei e recomendo!",
		"Chegou certinho aqui. Obrigado!",
	}
	if limit <= 0 || limit > len(base) {
		limit = len(base)
	}
	out := make([]FBComment, 0, limit)
	seed := int64(0)
	for _, ch := range postID {
		seed += int64(ch)
	}
	autores := []string{"Maria Souza", "João Lima", "Ana Clara", "Carlos R.", "Bea Mendes", "Pedro H.", "Lucas T.", "Renata P."}
	now := time.Now().UTC().Format(time.RFC3339)
	for i := 0; i < limit; i++ {
		idx := (seed + int64(i)) % int64(len(base))
		cid := atomic.AddInt64(&m.nextCID, 1)
		c := FBComment{
			ID:          fmt.Sprintf("mock_cmt_%s_%d", postID, cid),
			Message:     base[idx],
			CreatedTime: now,
		}
		c.From.ID = fmt.Sprintf("user_%d", 1000+idx)
		c.From.Name = autores[idx%int64(len(autores))]
		out = append(out, c)
	}
	return out, nil
}

func (m *MockClient) Reply(_ context.Context, commentID, message, _ string) (string, error) {
	if strings.TrimSpace(message) == "" {
		return "", fmt.Errorf("mock: mensagem vazia")
	}
	m.postedTo[commentID] = message
	return "mock_reply_" + commentID, nil
}
