// Tela de configuração: salva no backend e força o app a recriar os clients.

export function renderConfig(view, ctx) {
  const cfg = ctx.currentConfig || { page_id: '', page_access_token: '', gemini_api_key: '', perfil_empresa: '', usar_mock: true, configurado: false };

  view.innerHTML = `
    <div class="card col">
      <h2 style="margin:0 0 8px 0">Configuração</h2>
      <p class="muted">Suas credenciais ficam só no seu computador (<code>data/config.json</code>) e nunca são enviadas a lugar nenhum além das APIs oficiais.</p>

      <div class="row">
        <label style="display:flex;align-items:center;gap:8px;">
          <input type="checkbox" id="usar_mock" ${cfg.usar_mock ? 'checked' : ''}/>
          <span>Usar modo mock (dados fictícios — ideal para testar sem credenciais)</span>
        </label>
      </div>

      <div id="bloco-credenciais" ${cfg.usar_mock ? 'style="opacity:0.5;pointer-events:none"' : ''}>
        <label>Page ID</label>
        <input type="text" id="page_id" value="${escapeAttr(cfg.page_id)}" placeholder="Ex.: 1234567890"/>

        <label style="margin-top:10px;">Page Access Token (long-lived)</label>
        <input type="password" id="page_access_token" value="${escapeAttr(cfg.page_access_token)}" placeholder="EAAB..."/>

        <label style="margin-top:10px;">Gemini API Key</label>
        <input type="password" id="gemini_api_key" value="${escapeAttr(cfg.gemini_api_key)}" placeholder="AIza..."/>
      </div>

      <label style="margin-top:10px;">Perfil da empresa (prompt de sistema)</label>
      <textarea id="perfil_empresa" placeholder="Ex.: somos a Pizzaria X, atendimento de 18h às 23h, entregas no raio de 5km, pizza grande R$ 55, nunca inventar promoções que não existam...">${escapeText(cfg.perfil_empresa)}</textarea>

      <div class="row" style="margin-top:10px;">
        <button class="btn" id="salvar">Salvar configuração</button>
        <span id="cfg-status" class="muted"></span>
        <span class="spacer"></span>
        <span class="muted">Status: ${cfg.configurado ? '✅ configurado' : '⚠️ incompleto'}</span>
      </div>
    </div>
  `;

  const usarMock = view.querySelector('#usar_mock');
  const bloco = view.querySelector('#bloco-credenciais');
  usarMock.addEventListener('change', () => {
    bloco.style.opacity = usarMock.checked ? '0.5' : '1';
    bloco.style.pointerEvents = usarMock.checked ? 'none' : 'auto';
  });

  view.querySelector('#salvar').addEventListener('click', async () => {
    const dto = {
      page_id: view.querySelector('#page_id').value.trim(),
      page_access_token: view.querySelector('#page_access_token').value.trim(),
      gemini_api_key: view.querySelector('#gemini_api_key').value.trim(),
      perfil_empresa: view.querySelector('#perfil_empresa').value,
      usar_mock: usarMock.checked,
    };
    try {
      await ctx.api.SaveConfig(dto);
      ctx.setStatus('Configuração salva');
      ctx.currentConfig = await ctx.api.GetConfig();
      renderConfig(view, ctx);
    } catch (e) {
      ctx.setStatus('Erro ao salvar: ' + e, true);
    }
  });
}

function escapeText(s) {
  return (s || '').replace(/[&<>]/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;'}[c]));
}
function escapeAttr(s) { return escapeText(s).replace(/"/g, '&quot;'); }
