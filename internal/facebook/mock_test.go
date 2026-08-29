package facebook

import (
	"context"
	"strings"
	"testing"
)

func TestMockListPosts_RetornaAteOLimite(t *testing.T) {
	m := NewMockClient()
	posts, err := m.ListPosts(context.Background(), "tok", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) == 0 {
		t.Fatal("esperava posts mock")
	}
	posts2, _ := m.ListPosts(context.Background(), "tok", 2)
	if len(posts2) != 2 {
		t.Fatalf("limite não respeitado: %d", len(posts2))
	}
}

func TestMockListComments_GeraIDsUnicos(t *testing.T) {
	m := NewMockClient()
	c1, _ := m.ListComments(context.Background(), "p1", "tok", 3)
	c2, _ := m.ListComments(context.Background(), "p1", "tok", 3)
	seen := map[string]bool{}
	for _, c := range append(c1, c2...) {
		if seen[c.ID] {
			t.Fatalf("ID duplicado: %s", c.ID)
		}
		seen[c.ID] = true
		if c.Message == "" {
			t.Fatal("comentário sem texto")
		}
		if c.From.Name == "" {
			t.Fatal("comentário sem autor")
		}
	}
}

func TestMockReply_RejeitaVazio(t *testing.T) {
	m := NewMockClient()
	if _, err := m.Reply(context.Background(), "c1", "  ", "tok"); err == nil {
		t.Fatal("esperava erro com mensagem vazia")
	}
	id, err := m.Reply(context.Background(), "c1", "oi", "tok")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(id, "mock_reply_") {
		t.Fatalf("ID inesperado: %s", id)
	}
}
