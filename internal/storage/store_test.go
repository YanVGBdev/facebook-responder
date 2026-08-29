package storage

import (
	"path/filepath"
	"testing"
)

func tempStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	s, err := NewStore(filepath.Join(dir, "comments.json"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return s
}

func TestNewStore_CriaArquivoVazio(t *testing.T) {
	s := tempStore(t)
	if s == nil {
		t.Fatal("store nulo")
	}
	if got := s.Snapshot().Posts; len(got) != 0 {
		t.Fatalf("esperava 0 posts, veio %d", len(got))
	}
}

func TestUpsertPost_CriaESobrescreve(t *testing.T) {
	s := tempStore(t)
	if err := s.UpsertPost(Post{ID: "p1", TextoResumo: "oi"}); err != nil {
		t.Fatal(err)
	}
	if err := s.UpsertPost(Post{ID: "p1", TextoResumo: "oi2"}); err != nil {
		t.Fatal(err)
	}
	if got := s.Snapshot().Posts["p1"].TextoResumo; got != "oi2" {
		t.Fatalf("esperava 'oi2', veio %q", got)
	}
}

func TestUpsertComment_NovoEPendente(t *testing.T) {
	s := tempStore(t)
	_ = s.UpsertPost(Post{ID: "p1", TextoResumo: "post"})
	if err := s.UpsertComment("p1", Comment{ID: "c1", Autor: "Ana", Texto: "olá"}); err != nil {
		t.Fatal(err)
	}
	got := s.Snapshot().Posts["p1"].Comentarios["c1"]
	if got.Status != StatusPendente {
		t.Fatalf("status inicial deveria ser 'pendente', veio %q", got.Status)
	}
	if got.Texto != "olá" {
		t.Fatalf("texto perdido: %q", got.Texto)
	}
}

func TestUpsertComment_NaoSobrescreveRespondido(t *testing.T) {
	s := tempStore(t)
	_ = s.UpsertPost(Post{ID: "p1", TextoResumo: "post"})

	// Comentário já respondido.
	_ = s.UpsertComment("p1", Comment{ID: "c1", Texto: "original", Status: StatusRespondido, RespostaFinal: "ok!"})
	// Tentativa de regravar — NÃO pode sobrescrever.
	_ = s.UpsertComment("p1", Comment{ID: "c1", Texto: "modificado", Status: StatusPendente})

	got := s.Snapshot().Posts["p1"].Comentarios["c1"]
	if got.Texto != "original" {
		t.Fatalf("texto de comentário respondido foi sobrescrito: %q", got.Texto)
	}
	if got.Status != StatusRespondido {
		t.Fatalf("status mudou: %q", got.Status)
	}
	if got.RespostaFinal != "ok!" {
		t.Fatalf("resposta_final perdida: %q", got.RespostaFinal)
	}
}

func TestSetSugestao_PendenteParaSugerido(t *testing.T) {
	s := tempStore(t)
	_ = s.UpsertPost(Post{ID: "p1", TextoResumo: "post"})
	_ = s.UpsertComment("p1", Comment{ID: "c1", Texto: "oi"})
	if err := s.SetSugestao("p1", "c1", "olá!"); err != nil {
		t.Fatal(err)
	}
	got := s.Snapshot().Posts["p1"].Comentarios["c1"]
	if got.Status != StatusSugerido {
		t.Fatalf("esperava 'sugerido', veio %q", got.Status)
	}
	if got.SugestaoIA != "olá!" {
		t.Fatalf("sugestão não gravada: %q", got.SugestaoIA)
	}
}

func TestSetSugestao_IgnoraRespondido(t *testing.T) {
	s := tempStore(t)
	_ = s.UpsertPost(Post{ID: "p1"})
	_ = s.UpsertComment("p1", Comment{ID: "c1", Texto: "x", Status: StatusRespondido, SugestaoIA: "antiga"})
	_ = s.SetSugestao("p1", "c1", "nova")
	got := s.Snapshot().Posts["p1"].Comentarios["c1"]
	if got.SugestaoIA != "antiga" {
		t.Fatalf("sugestão de comentário respondido foi mexida: %q", got.SugestaoIA)
	}
}

func TestMarcarRespondido_GravaRespostaEData(t *testing.T) {
	s := tempStore(t)
	_ = s.UpsertPost(Post{ID: "p1"})
	_ = s.UpsertComment("p1", Comment{ID: "c1", Texto: "x", Status: StatusSugerido, SugestaoIA: "s"})
	if err := s.MarcarRespondido("p1", "c1", "s", "2025-01-01T00:00:00Z"); err != nil {
		t.Fatal(err)
	}
	got := s.Snapshot().Posts["p1"].Comentarios["c1"]
	if got.Status != StatusRespondido {
		t.Fatalf("status: %q", got.Status)
	}
	if got.RespostaFinal != "s" {
		t.Fatalf("resposta: %q", got.RespostaFinal)
	}
	if got.RespondidoEm == "" {
		t.Fatal("respondido_em vazio")
	}
}

func TestSetRespostaFinal_NaoMexeNoStatus(t *testing.T) {
	s := tempStore(t)
	_ = s.UpsertPost(Post{ID: "p1"})
	_ = s.UpsertComment("p1", Comment{ID: "c1", Texto: "x", Status: StatusSugerido})
	_ = s.SetRespostaFinal("p1", "c1", "editado")
	got := s.Snapshot().Posts["p1"].Comentarios["c1"]
	if got.Status != StatusSugerido {
		t.Fatalf("status mudou: %q", got.Status)
	}
	if got.RespostaFinal != "editado" {
		t.Fatalf("rascunho não salvo: %q", got.RespostaFinal)
	}
}

func TestUpsertComment_PreservaSugestaoEmAtualizacao(t *testing.T) {
	s := tempStore(t)
	_ = s.UpsertPost(Post{ID: "p1"})
	_ = s.UpsertComment("p1", Comment{ID: "c1", Texto: "oi", Status: StatusSugerido, SugestaoIA: "rascunho"})
	// Nova busca traz o comentário sem SugestaoIA/Status — não deve apagar a sugestão existente.
	_ = s.UpsertComment("p1", Comment{ID: "c1", Texto: "oi", Autor: "Ana"})
	got := s.Snapshot().Posts["p1"].Comentarios["c1"]
	if got.SugestaoIA != "rascunho" {
		t.Fatalf("sugestão perdida em nova busca: %q", got.SugestaoIA)
	}
	if got.Status != StatusSugerido {
		t.Fatalf("status perdido: %q", got.Status)
	}
	if got.Autor != "Ana" {
		t.Fatalf("autor não atualizado: %q", got.Autor)
	}
}

func TestErros_PostOuComentarioInexistente(t *testing.T) {
	s := tempStore(t)
	if err := s.SetSugestao("nao_existe", "c", "x"); err == nil {
		t.Fatal("esperava erro para post inexistente")
	}
	if err := s.MarcarRespondido("nao_existe", "c", "x", "agora"); err == nil {
		t.Fatal("esperava erro")
	}
	_ = s.UpsertPost(Post{ID: "p1"})
	if err := s.MarcarRespondido("p1", "nao_existe", "x", "agora"); err == nil {
		t.Fatal("esperava erro para comentário inexistente")
	}
}
