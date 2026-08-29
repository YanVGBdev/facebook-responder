package ai

import (
	"context"
	"strings"
	"testing"
)

func TestMontaPrompt_IncluiPerfilEComentario(t *testing.T) {
	sys, user := MontaPrompt(PromptInput{
		PerfilEmpresa: "Pizzaria X, 18h–23h",
		PostResumo:    "Post de inauguração",
		Autor:         "Maria",
		Comentario:    "Vocês entregam?",
	})
	if !strings.Contains(sys, "Pizzaria X, 18h–23h") {
		t.Fatalf("system prompt não inclui perfil: %q", sys)
	}
	if !strings.Contains(sys, "português do Brasil") {
		t.Fatalf("system prompt não força pt-BR: %q", sys)
	}
	if !strings.Contains(user, "Maria") || !strings.Contains(user, "Vocês entregam?") {
		t.Fatalf("user prompt incompleto: %q", user)
	}
	if !strings.Contains(user, "APENAS o texto da resposta") {
		t.Fatalf("user prompt sem instrução de saída: %q", user)
	}
}

func TestMockSugerir_BranchPreco(t *testing.T) {
	c := NewMock()
	out, err := c.Sugerir(context.Background(), PromptInput{
		PerfilEmpresa: "loja x",
		Autor:         "João",
		Comentario:    "qual o preço?",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "João") {
		t.Fatalf("não citou o autor: %q", out)
	}
}

func TestMockSugerir_BranchEntrega(t *testing.T) {
	c := NewMock()
	out, err := c.Sugerir(context.Background(), PromptInput{
		PerfilEmpresa: "loja x",
		Autor:         "Ana",
		Comentario:    "qual o prazo de entrega?",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ToLower(out), "direct") && !strings.Contains(strings.ToLower(out), "entrega") {
		t.Fatalf("resposta não parece cobrir entrega: %q", out)
	}
}

func TestMockSugerir_BranchElogio(t *testing.T) {
	c := NewMock()
	out, err := c.Sugerir(context.Background(), PromptInput{
		Autor:      "Bea",
		Comentario: "amei, recomendo!",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ToLower(out), "obrigad") {
		t.Fatalf("resposta não agradece: %q", out)
	}
}

func TestMockSugerir_BranchGenericoIncluiPerfil(t *testing.T) {
	c := NewMock()
	out, err := c.Sugerir(context.Background(), PromptInput{
		PerfilEmpresa: "MARCAROMA — especializada em cafeterias",
		Autor:         "Leo",
		Comentario:    "olá",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "MARCAROMA") {
		t.Fatalf("resposta genérica não citou perfil: %q", out)
	}
}

func TestMockSugerir_SemAutorUsaAmigo(t *testing.T) {
	c := NewMock()
	out, err := c.Sugerir(context.Background(), PromptInput{
		Comentario: "olá",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "amigo(a)") {
		t.Fatalf("não usou fallback de autor: %q", out)
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("abcdef", 3); got != "ab…" {
		t.Fatalf("truncate: %q", got)
	}
	if truncate("abc", 10) != "abc" {
		t.Fatal("truncate não-idempotente para tamanho menor")
	}
}
