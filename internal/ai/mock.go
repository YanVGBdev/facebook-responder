package ai

import (
	"context"
	"fmt"
	"strings"
)

// MockClient devolve uma sugestão plausível sem chamar a API.
// Usado quando config.usar_mock=true.
type MockClient struct{}

func NewMock() *MockClient { return &MockClient{} }

func (m *MockClient) Sugerir(_ context.Context, in PromptInput) (string, error) {
	perfil := strings.TrimSpace(in.PerfilEmpresa)
	if perfil == "" {
		perfil = "empresa genérica"
	}
	texto := strings.TrimSpace(in.Comentario)
	autor := strings.TrimSpace(in.Autor)
	if autor == "" {
		autor = "amigo(a)"
	}
	// Regras simples só para simular; em produção isso é a Gemini.
	low := strings.ToLower(texto)
	switch {
	case strings.Contains(low, "preço") || strings.Contains(low, "preco") || strings.Contains(low, "valor") || strings.Contains(low, "quanto"):
		return fmt.Sprintf("Oi, %s! Os valores estão no nosso perfil da empresa. Qualquer dúvida, é só chamar aqui. 😊", autor), nil
	case strings.Contains(low, "entrega") || strings.Contains(low, "prazo"):
		return fmt.Sprintf("Oi, %s! Atendemos várias regiões. Me chama no direct com seu CEP que te passo o prazo certinho. 📦", autor), nil
	case strings.Contains(low, "obrigad") || strings.Contains(low, "amei") || strings.Contains(low, "recomendo"):
		return fmt.Sprintf("%s, obrigado pelo carinho! Você é demais. 💜", autor), nil
	default:
		return fmt.Sprintf("Oi, %s! Obrigado pelo comentário. A gente segue alinhado com nosso perfil: %s. Se quiser, me conta mais!", autor, truncate(perfil, 80)), nil
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
