package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

// Store gerencia o snapshot persistido em disco (data/comments.json).
// Usa mutex para serializar leituras/escritas concorrentes vindas da UI.
type Store struct {
	mu   sync.Mutex
	path string
	data Snapshot
}

// NewStore cria/abre o arquivo de estado. Se o arquivo não existir,
// inicia com um snapshot vazio e o persiste.
func NewStore(path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("storage: path vazio")
	}
	s := &Store{path: path, data: Snapshot{Posts: map[string]Post{}}}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		if err := s.persistLocked(); err != nil {
			return nil, err
		}
		return s, nil
	} else if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(b) > 0 {
		if err := json.Unmarshal(b, &s.data); err != nil {
			return nil, err
		}
	}
	if s.data.Posts == nil {
		s.data.Posts = map[string]Post{}
	}
	return s, nil
}

// Snapshot retorna uma cópia do estado atual.
func (s *Store) Snapshot() Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data
}

// UpsertPost cria/atualiza um post sem mexer nos comentários existentes.
func (s *Store) UpsertPost(p Post) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.data.Posts == nil {
		s.data.Posts = map[string]Post{}
	}
	existente, ok := s.data.Posts[p.ID]
	if !ok {
		if p.Comentarios == nil {
			p.Comentarios = map[string]Comment{}
		}
		s.data.Posts[p.ID] = p
	} else {
		existente.TextoResumo = p.TextoResumo
		if existente.TextoResumo == "" {
			existente.TextoResumo = p.TextoResumo
		}
		if p.CreatedTime != "" {
			existente.CreatedTime = p.CreatedTime
		}
		if existente.Comentarios == nil {
			existente.Comentarios = map[string]Comment{}
		}
		for cid, c := range p.Comentarios {
			existente.Comentarios[cid] = c
		}
		s.data.Posts[p.ID] = existente
	}
	return s.persistLocked()
}

// UpsertComment cria/atualiza um comentário dentro de um post.
// NÃO sobrescreve status/resposta_final se já estiver respondido.
func (s *Store) UpsertComment(postID string, c Comment) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.data.Posts[postID]
	if !ok {
		p = Post{ID: postID, Comentarios: map[string]Comment{}}
	}
	if p.Comentarios == nil {
		p.Comentarios = map[string]Comment{}
	}
	existente, existe := p.Comentarios[c.ID]
	if existe && existente.Status == StatusRespondido {
		// Não mexe em comentário já respondido.
		return nil
	}
	if existe {
		// Mantém sugestão/resposta em edição se já houver.
		if c.SugestaoIA == "" {
			c.SugestaoIA = existente.SugestaoIA
		}
		if c.Status == "" {
			c.Status = existente.Status
		}
		if c.Autor == "" {
			c.Autor = existente.Autor
		}
		if c.Texto == "" {
			c.Texto = existente.Texto
		}
		if c.CreatedTime == "" {
			c.CreatedTime = existente.CreatedTime
		}
	} else if c.Status == "" {
		// Comentário novo: status padrão "pendente" (defesa em profundidade —
		// o app já passa isso explicitamente, mas o store deve assumir um
		// default razoável se o caller esquecer).
		c.Status = StatusPendente
	}
	p.Comentarios[c.ID] = c
	s.data.Posts[postID] = p
	return s.persistLocked()
}

// SetSugestao grava a sugestão da IA e marca como "sugerido".
func (s *Store) SetSugestao(postID, commentID, sugestao string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.data.Posts[postID]
	if !ok {
		return errors.New("post não encontrado")
	}
	c, ok := p.Comentarios[commentID]
	if !ok {
		return errors.New("comentário não encontrado")
	}
	if c.Status == StatusRespondido {
		return nil
	}
	c.SugestaoIA = sugestao
	c.Status = StatusSugerido
	p.Comentarios[commentID] = c
	s.data.Posts[postID] = p
	return s.persistLocked()
}

// SetRespostaFinal salva o texto editado pelo humano (sem publicar).
func (s *Store) SetRespostaFinal(postID, commentID, texto string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.data.Posts[postID]
	if !ok {
		return errors.New("post não encontrado")
	}
	c, ok := p.Comentarios[commentID]
	if !ok {
		return errors.New("comentário não encontrado")
	}
	c.RespostaFinal = texto
	p.Comentarios[commentID] = c
	s.data.Posts[postID] = p
	return s.persistLocked()
}

// MarcarRespondido move o comentário para "respondido".
func (s *Store) MarcarRespondido(postID, commentID, respostaFinal, quando string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.data.Posts[postID]
	if !ok {
		return errors.New("post não encontrado")
	}
	c, ok := p.Comentarios[commentID]
	if !ok {
		return errors.New("comentário não encontrado")
	}
	c.Status = StatusRespondido
	c.RespostaFinal = respostaFinal
	c.RespondidoEm = quando
	p.Comentarios[commentID] = c
	s.data.Posts[postID] = p
	return s.persistLocked()
}

func (s *Store) persistLocked() error {
	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
