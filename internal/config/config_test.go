package config

import (
	"path/filepath"
	"testing"
)

func TestNew_CriaPadraoComUsarMock(t *testing.T) {
	dir := t.TempDir()
	m, err := New(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	c := m.Get()
	if !c.UsarMock {
		t.Fatal("config padrão deveria vir com UsarMock=true")
	}
	if c.PageID != "" || c.PageAccessToken != "" || c.GeminiAPIKey != "" {
		t.Fatal("config padrão deveria ter credenciais vazias")
	}
}

func TestSaveAndGet_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	m, _ := New(filepath.Join(dir, "config.json"))
	in := Config{
		PageID:          "123",
		PageAccessToken: "EAABxx",
		GeminiAPIKey:    "AIzaxxx",
		PerfilEmpresa:   "sou a Pizzaria X",
		UsarMock:        false,
	}
	if err := m.Save(in); err != nil {
		t.Fatal(err)
	}
	// Reabrir do disco.
	m2, err := New(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	out := m2.Get()
	if out.PageID != in.PageID || out.PageAccessToken != in.PageAccessToken ||
		out.GeminiAPIKey != in.GeminiAPIKey || out.PerfilEmpresa != in.PerfilEmpresa ||
		out.UsarMock != in.UsarMock {
		t.Fatalf("round-trip divergiu: %+v vs %+v", in, out)
	}
}

func TestConfigurado_ComMockSempreTrue(t *testing.T) {
	dir := t.TempDir()
	m, _ := New(filepath.Join(dir, "config.json"))
	if !m.Configurado() {
		t.Fatal("com mock=true, Configurado() deveria ser true")
	}
}

func TestConfigurado_RealExigeCredenciais(t *testing.T) {
	dir := t.TempDir()
	m, _ := New(filepath.Join(dir, "config.json"))
	_ = m.Save(Config{UsarMock: false, PageID: "1", PageAccessToken: "t", GeminiAPIKey: "k"})
	if !m.Configurado() {
		t.Fatal("com credenciais + UsarMock=false, deveria ser true")
	}
	_ = m.Save(Config{UsarMock: false})
	if m.Configurado() {
		t.Fatal("sem credenciais e sem mock, deveria ser false")
	}
}
