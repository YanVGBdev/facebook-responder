package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

// Config é o conteúdo do config.json (tokens + perfil da empresa).
type Config struct {
	PageID          string `json:"page_id"`
	PageAccessToken string `json:"page_access_token"`
	GeminiAPIKey    string `json:"gemini_api_key"`
	PerfilEmpresa   string `json:"perfil_empresa"`

	// Modo "mock" (sem credenciais reais). Quando true, o app usa
	// dados fictícios para Facebook e Gemini, ideal para teste.
	UsarMock bool `json:"usar_mock"`
}

// Manager lê e grava config.json com proteção contra concorrência.
type Manager struct {
	mu   sync.Mutex
	path string
	data Config
}

// New carrega o config; se o arquivo não existir, cria um com UsarMock=true.
func New(path string) (*Manager, error) {
	if path == "" {
		return nil, errors.New("config: path vazio")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	m := &Manager{path: path, data: Config{UsarMock: true}}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		if err := m.saveLocked(); err != nil {
			return nil, err
		}
		return m, nil
	} else if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(b) > 0 {
		if err := json.Unmarshal(b, &m.data); err != nil {
			return nil, err
		}
	}
	return m, nil
}

// Get retorna uma cópia do config atual.
func (m *Manager) Get() Config {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.data
}

// Save persiste o config recebido.
func (m *Manager) Save(c Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = c
	return m.saveLocked()
}

func (m *Manager) saveLocked() error {
	b, err := json.MarshalIndent(m.data, "", "  ")
	if err != nil {
		return err
	}
	tmp := m.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, m.path)
}

// Configurado retorna true se as credenciais mínimas estão presentes
// (e o modo mock está desligado).
func (m *Manager) Configurado() bool {
	c := m.Get()
	if c.UsarMock {
		return true
	}
	return c.PageID != "" && c.PageAccessToken != "" && c.GeminiAPIKey != ""
}
