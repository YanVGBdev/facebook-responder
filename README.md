# Facebook Comments AI Responder

App desktop para Windows (Go + Wails) que ajuda o dono de uma Página do Facebook a responder comentários com apoio de IA. **Nada é publicado sem revisão e clique explícito do usuário.**

> Veja a especificação completa em [`../especificacao-facebook-responder.md`](../especificacao-facebook-responder.md).

## ✨ O que o app faz

1. Lê posts e comentários da sua Página via **Meta Graph API**.
2. Para cada comentário pendente, pede ao **Google Gemini** uma sugestão de resposta, usando o "perfil" da empresa como prompt de sistema.
3. Você revisa, edita se quiser, salva rascunho ou clica **Enviar**.
4. O app publica via Graph API e marca o comentário como respondido.

Fora de escopo (decisão consciente): automação de "convidir para seguir" — não existe na API oficial.

## 🧱 Stack

- **Go 1.21+** + **Wails v2** (HTML/CSS/JS no WebView2 do Windows, lógica em Go)
- **Meta Graph API v20.0** — endpoints `/{page-id}/posts`, `/{post-id}/comments`, `/{comment-id}/comments`
- **Google Gemini 1.5 Flash** via `generativelanguage.googleapis.com`
- **Persistência local** em `data/config.json` + `data/comments.json` (JSON, sem banco)

## 🚀 Como rodar

### Pré-requisitos

- **Windows 10/11** com [Go 1.21+](https://go.dev/dl/) e [Wails v2](https://wails.io/docs/gettingstarted/installation)
- (Opcional) Node.js só para rebuild do frontend; o app já vem com `frontend/dist/` pronto após `wails build`

### Build

```bash
wails build              # gera build/bin/facebook-responder.exe
.\build\bin\facebook-responder.exe
```

### Dev (hot reload)

```bash
wails dev                # UI em janela + devtools; recarrega ao salvar
```

## 🧪 Testes

```bash
go test ./...
```

Cobre as regras do `storage` (status, não-sobrescrita de respondidos, preservação de sugestão), `ai` (montagem de prompt, branches do mock), `config` (round-trip, flags) e `facebook` (mock).

## 🤖 CI (GitHub Actions)

O workflow `.github/workflows/build-windows.yml` builda o `.exe` em `windows-latest` a cada push/PR e sobe como artifact.

- Roda `go vet`, `go test ./...`, depois `wails build -clean -platform windows/amd64`.
- Artifact: `facebook-responder-windows-amd64` (contém `facebook-responder.exe`, retenção 14 dias).
- Pode ser disparado manualmente em **Actions → build-windows → Run workflow**.

Para ativar: faça push do `.github/workflows/build-windows.yml` para o repositório.

## 🔑 Como obter as credenciais

O app abre **em modo mock** na primeira execução (dados fictícios, ideal para testar o fluxo). Para usar com sua Página real, abra a aba **Configuração**, desmarque "Usar modo mock" e preencha:

### 1. Page ID e Page Access Token (Meta)

1. Acesse <https://developers.facebook.com/apps/> e crie um app do tipo **"Empresa"** ou **"Outro" → "Negócios"**.
2. Em **Adicionar produto → Graph API**, clique em **Configurar**.
3. Em **Permissões e recursos**, solicite:
   - `pages_read_engagement` (ler posts e comentários)
   - `pages_manage_engagement` (publicar respostas)
4. Gere um **token de acesso da Página** (long-lived, 60 dias) em **Ferramentas → Graph API Explorer**:
   - Selecione seu app
   - Adicione as duas permissões acima
   - Selecione a Página alvo (não "usuário")
   - Gere o token, **depois estenda para long-lived** via `https://graph.facebook.com/v20.0/oauth/access_token?grant_type=fb_exchange_token&client_id={app_id}&client_secret={app_secret}&fb_exchange_token={short_token}` e em seguida para **never-expiring** com `?grant_type=fb_exchange_token&...` apontando para `/me/accounts` (ver doc "Long-Lived Page Access Token").
5. Copie o **Page ID** (em **Configurações → Sua Página → Sobre**) e o token para a tela de Configuração.

⚠️ **Nunca compartilhe o token.** Ele fica só em `data/config.json` (ignorado no `.gitignore`).

### 2. Gemini API Key

1. Acesse <https://aistudio.google.com/app/apikey> com sua conta Google.
2. Clique em **Create API Key** e copie.
3. Cole na tela de Configuração.

A chave é independente da sua assinatura do Gemini (não consome plano pago do Workspace).

### 3. Perfil da empresa

Texto livre com tom de voz, produtos, faixas de preço e **regras do que nunca inventar** (ex.: "NUNCA prometa frete grátis; nunca cite produto X"). Quanto mais específico, melhor a resposta. Esse texto vai inteiro como `system_instruction` para a Gemini.

## 🧩 Estrutura

```
facebook-responder/
├── main.go                 # entry, carrega config+store, escolhe clients mock/real
├── app.go                  # struct Wails, métodos expostos ao frontend
├── internal/
│   ├── storage/            # JSON store (data/comments.json) com mutex
│   ├── config/             # config.json com flag usar_mock
│   ├── facebook/           # GraphClient (real) + MockClient (sem rede)
│   └── ai/                 # GeminiClient (real) + MockClient + prompt builder
├── frontend/
│   ├── index.html          # shell pt-BR
│   ├── src/
│   │   ├── main.js         # roteador de abas
│   │   ├── api.js          # carrega bindings gerados
│   │   ├── style.css       # tema dark
│   │   └── screens/        # Config.js, Posts.js
│   ├── wailsjs/            # bindings gerados (NÃO editar)
│   └── dist/               # build Vite
├── build/                  # ícones + manifesto Windows
└── data/                   # config.json + comments.json (runtime, gitignored)
```

## 🔄 Alternando entre mock e real

- **Mock** (padrão inicial): use sem credenciais. 3 posts fictícios, ~8 comentários por post, sugestões heurísticas (preço/entrega/elogio/genérico).
- **Real**: desmarque "Usar modo mock", preencha as credenciais, salve. O app recria os clients automaticamente — não precisa reiniciar.

Os arquivos em `data/` continuam existindo nos dois modos; só o que muda é de onde vêm os dados.

## 🛡️ Boas práticas aplicadas

- **Sem publicação automática**: todo envio exige `confirm()` + clique em **Enviar**.
- **Sem scraping/automação de navegador**: só Graph API oficial.
- **Tokens locais**: nunca commitados, ignorados pelo `.gitignore`.
- **Comentários já respondidos nunca são sobrescritos** por novas buscas.
- **Defesa em profundidade**: o store assume `status=pendente` se o caller não setar; o `app.go` seta explicitamente.

## 📋 Limites

- A app não implementa backoff automático para rate limit da Graph API — respeite os limites e evite buscas em loop.
- Não há UI para revisar tokens antes de salvar — confira na próxima execução.
- Cross-compile para Windows a partir de Linux requer MinGW e `libwebkit2gtk-4.0-dev`; o caminho recomendado é buildar direto no Windows.
