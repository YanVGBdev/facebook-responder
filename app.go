package main

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"facebook-responder/internal/ai"
	"facebook-responder/internal/config"
	"facebook-responder/internal/facebook"
	"facebook-responder/internal/storage"
)

// App é o ponto de integração Wails. Os métodos públicos aqui viram
// funções JavaScript disponíveis em window.go.main.App.
type App struct {
	ctx       context.Context
	mu        sync.Mutex
	cfg       *config.Manager
	store     *storage.Store
	fbClient  facebook.Client
	aiClient  ai.Client
	lastError string
}

func NewApp(cfg *config.Manager, store *storage.Store, fb facebook.Client, aiC ai.Client) *App {
	return &App{cfg: cfg, store: store, fbClient: fb, aiClient: aiC}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// --- DTOs expostos ao frontend (Wails gera o TS) ---

type ConfigDTO struct {
	PageID          string `json:"page_id"`
	PageAccessToken string `json:"page_access_token"`
	GeminiAPIKey    string `json:"gemini_api_key"`
	PerfilEmpresa   string `json:"perfil_empresa"`
	UsarMock        bool   `json:"usar_mock"`
	Configurado     bool   `json:"configurado"`
}

type CommentDTO struct {
	ID            string `json:"id"`
	Autor         string `json:"autor"`
	Texto         string `json:"texto"`
	CreatedTime   string `json:"created_time,omitempty"`
	Status        string `json:"status"`
	SugestaoIA    string `json:"sugestao_ia"`
	RespostaFinal string `json:"resposta_final"`
	RespondidoEm  string `json:"respondido_em,omitempty"`
}

type PostDTO struct {
	ID          string       `json:"id"`
	TextoResumo string       `json:"texto_resumo"`
	CreatedTime string       `json:"created_time,omitempty"`
	Total       int          `json:"total"`
	Pendentes   int          `json:"pendentes"`
	Comentarios []CommentDTO `json:"comentarios"`
}

type PostsListDTO struct {
	Posts []PostDTO `json:"posts"`
}

// --- Métodos Wails ---

// GetConfig devolve o config atual (oculta campos sensíveis se você quiser
// futuramente — por ora devolve como está para a tela de configuração).
func (a *App) GetConfig() ConfigDTO {
	c := a.cfg.Get()
	return ConfigDTO{
		PageID:          c.PageID,
		PageAccessToken: c.PageAccessToken,
		GeminiAPIKey:    c.GeminiAPIKey,
		PerfilEmpresa:   c.PerfilEmpresa,
		UsarMock:        c.UsarMock,
		Configurado:     a.cfg.Configurado(),
	}
}

// SaveConfig persiste e re-instancia os clients conforme o modo (mock/real).
func (a *App) SaveConfig(dto ConfigDTO) error {
	c := config.Config{
		PageID:          dto.PageID,
		PageAccessToken: dto.PageAccessToken,
		GeminiAPIKey:    dto.GeminiAPIKey,
		PerfilEmpresa:   dto.PerfilEmpresa,
		UsarMock:        dto.UsarMock,
	}
	if err := a.cfg.Save(c); err != nil {
		return err
	}
	a.rebuildClients(c)
	return nil
}

// ListarPosts retorna a lista de posts persistida.
func (a *App) ListarPosts() PostsListDTO {
	snap := a.store.Snapshot()
	out := make([]PostDTO, 0, len(snap.Posts))
	for _, p := range snap.Posts {
		pendentes := 0
		cmts := make([]CommentDTO, 0, len(p.Comentarios))
		for _, c := range p.Comentarios {
			if c.Status != storage.StatusRespondido {
				pendentes++
			}
			cmts = append(cmts, CommentDTO{
				ID:            c.ID,
				Autor:         c.Autor,
				Texto:         c.Texto,
				CreatedTime:   c.CreatedTime,
				Status:        c.Status,
				SugestaoIA:    c.SugestaoIA,
				RespostaFinal: c.RespostaFinal,
				RespondidoEm:  c.RespondidoEm,
			})
		}
		out = append(out, PostDTO{
			ID:          p.ID,
			TextoResumo: p.TextoResumo,
			CreatedTime: p.CreatedTime,
			Total:       len(p.Comentarios),
			Pendentes:   pendentes,
			Comentarios: cmts,
		})
	}
	return PostsListDTO{Posts: out}
}

// BuscarPosts consulta a Graph API (ou mock) e atualiza o store.
// limit define quantos posts pegar; commentsLimit por post.
func (a *App) BuscarPosts(limit, commentsLimit int) (PostsListDTO, error) {
	cfg := a.cfg.Get()
	token := cfg.PageAccessToken
	if !cfg.UsarMock && (cfg.PageID == "" || token == "") {
		return PostsListDTO{}, fmt.Errorf("configure o Page ID e o Page Access Token (ou ative o modo mock)")
	}
	posts, err := a.fbClient.ListPosts(a.ctx, token, limit)
	if err != nil {
		a.setErr(err)
		return PostsListDTO{}, err
	}
	for _, p := range posts {
		texto := p.Message
		if texto == "" {
			texto = p.Story
		}
		if texto == "" {
			texto = "(post sem texto)"
		}
		if len(texto) > 200 {
			texto = texto[:200] + "…"
		}
		if err := a.store.UpsertPost(storage.Post{
			ID:          p.ID,
			TextoResumo: texto,
			CreatedTime: p.CreatedTime,
		}); err != nil {
			return PostsListDTO{}, err
		}
		cmts, err := a.fbClient.ListComments(a.ctx, p.ID, token, commentsLimit)
		if err != nil {
			a.setErr(err)
			return PostsListDTO{}, err
		}
		for _, c := range cmts {
			if err := a.store.UpsertComment(p.ID, storage.Comment{
				ID:          c.ID,
				Autor:       c.From.Name,
				Texto:       c.Message,
				CreatedTime: c.CreatedTime,
				Status:      storage.StatusPendente,
			}); err != nil {
				return PostsListDTO{}, err
			}
		}
	}
	return a.ListarPosts(), nil
}

// GerarSugestoes chama a IA (ou mock) para todos os comentários pendentes
// do post que ainda não têm sugestão.
func (a *App) GerarSugestoes(postID string) error {
	cfg := a.cfg.Get()
	snap := a.store.Snapshot()
	p, ok := snap.Posts[postID]
	if !ok {
		return fmt.Errorf("post não encontrado")
	}
	for _, c := range p.Comentarios {
		if c.Status == storage.StatusRespondido {
			continue
		}
		if c.SugestaoIA != "" {
			continue
		}
		sug, err := a.aiClient.Sugerir(a.ctx, ai.PromptInput{
			PerfilEmpresa: cfg.PerfilEmpresa,
			PostResumo:    p.TextoResumo,
			Autor:         c.Autor,
			Comentario:    c.Texto,
		})
		if err != nil {
			a.setErr(err)
			return err
		}
		if err := a.store.SetSugestao(postID, c.ID, sug); err != nil {
			return err
		}
	}
	return nil
}

// EditarResposta salva o texto editado pelo humano (não publica).
func (a *App) EditarResposta(postID, commentID, texto string) error {
	return a.store.SetRespostaFinal(postID, commentID, texto)
}

// EnviarResposta publica via Graph API e marca como respondido.
func (a *App) EnviarResposta(postID, commentID string) (CommentDTO, error) {
	cfg := a.cfg.Get()
	snap := a.store.Snapshot()
	p, ok := snap.Posts[postID]
	if !ok {
		return CommentDTO{}, fmt.Errorf("post não encontrado")
	}
	c, ok := p.Comentarios[commentID]
	if !ok {
		return CommentDTO{}, fmt.Errorf("comentário não encontrado")
	}
	if c.Status == storage.StatusRespondido {
		return CommentDTO{}, fmt.Errorf("comentário já respondido")
	}
	final := c.RespostaFinal
	if final == "" {
		final = c.SugestaoIA
	}
	if final == "" {
		return CommentDTO{}, fmt.Errorf("nenhuma resposta definida")
	}
	if !cfg.UsarMock && cfg.PageAccessToken == "" {
		return CommentDTO{}, fmt.Errorf("Page Access Token ausente (ou ative o modo mock)")
	}
	if _, err := a.fbClient.Reply(a.ctx, commentID, final, cfg.PageAccessToken); err != nil {
		a.setErr(err)
		return CommentDTO{}, err
	}
	quando := time.Now().UTC().Format(time.RFC3339)
	if err := a.store.MarcarRespondido(postID, commentID, final, quando); err != nil {
		return CommentDTO{}, err
	}
	return CommentDTO{
		ID: c.ID, Autor: c.Autor, Texto: c.Texto,
		Status: storage.StatusRespondido, SugestaoIA: c.SugestaoIA,
		RespostaFinal: final, RespondidoEm: quando,
	}, nil
}

// LastError devolve o último erro registrado (apenas para o frontend exibir).
func (a *App) LastError() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.lastError
}

func (a *App) setErr(err error) {
	a.mu.Lock()
	a.lastError = err.Error()
	a.mu.Unlock()
}

func (a *App) rebuildClients(c config.Config) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if c.UsarMock {
		a.fbClient = facebook.NewMockClient()
		a.aiClient = ai.NewMock()
	} else {
		a.fbClient = facebook.NewGraphClient(c.PageID)
		a.aiClient = ai.NewGemini(c.GeminiAPIKey)
	}
}

// resolveDataPath resolve um caminho relativo ao diretório do executável
// (em dev fica ao lado do binário; em produção também).
func resolveDataPath(rel string) string {
	abs, err := filepath.Abs(rel)
	if err != nil {
		return rel
	}
	return abs
}
