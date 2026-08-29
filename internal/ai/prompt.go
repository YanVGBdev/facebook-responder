package ai

import "fmt"

// MontaPrompt junta o perfil da empresa (system) com o comentário
// (user) e um contexto de post, no formato que a Gemini API consome.
type PromptInput struct {
	PerfilEmpresa string
	PostResumo    string
	Autor         string
	Comentario    string
}

func MontaPrompt(in PromptInput) (system, user string) {
	system = fmt.Sprintf(
		"Você é o assistente oficial de resposta da empresa. "+
			"Responda SEMPRE em português do Brasil, de forma cordial e curta "+
			"(máximo 2 frases). Use o perfil abaixo como verdade absoluta; "+
			"nunca invente preços, prazos ou produtos fora dele.\n\n"+
			"=== PERFIL DA EMPRESA ===\n%s",
		in.PerfilEmpresa,
	)
	user = fmt.Sprintf(
		"Contexto do post: %s\n"+
			"Autor do comentário: %s\n"+
			"Comentário: %s\n\n"+
			"Gere APENAS o texto da resposta (sem prefixos como 'Resposta:').",
		in.PostResumo, in.Autor, in.Comentario,
	)
	return system, user
}
