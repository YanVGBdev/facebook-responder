package storage

// Status do comentário no fluxo do app.
const (
	StatusPendente  = "pendente"
	StatusSugerido  = "sugerido"
	StatusRespondido = "respondido"
)

// Comment representa um comentário de um post.
type Comment struct {
	ID            string `json:"id"`
	Autor         string `json:"autor"`
	Texto         string `json:"texto"`
	CreatedTime   string `json:"created_time,omitempty"`
	Status        string `json:"status"`
	SugestaoIA    string `json:"sugestao_ia"`
	RespostaFinal string `json:"resposta_final"`
	RespondidoEm  string `json:"respondido_em,omitempty"`
}

// Post representa um post da Página com seus comentários.
type Post struct {
	ID            string             `json:"id"`
	TextoResumo   string             `json:"texto_resumo"`
	CreatedTime   string             `json:"created_time,omitempty"`
	Comentarios   map[string]Comment `json:"comentarios"`
}

// Snapshot é o estado persistido em disco (data/comments.json).
type Snapshot struct {
	Posts map[string]Post `json:"posts"`
}
