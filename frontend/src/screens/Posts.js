// Tela de posts: lista, busca novos e abre o detalhe de comentários.

export async function renderPosts(view, ctx) {
  view.innerHTML = `
    <div class="card">
      <div class="row">
        <button class="btn" id="buscar">🔄 Buscar posts e comentários</button>
        <span class="muted" id="busca-info"></span>
        <span class="spacer"></span>
        <label class="muted" style="margin:0;">
          Posts:
          <input type="number" id="limite-posts" value="5" min="1" max="25" style="width:60px; margin-left:6px;"/>
        </label>
        <label class="muted" style="margin:0;">
          Comentários/post:
          <input type="number" id="limite-cmts" value="10" min="1" max="50" style="width:60px; margin-left:6px;"/>
        </label>
      </div>
    </div>
    <div id="lista-posts"></div>
  `;

  view.querySelector('#buscar').addEventListener('click', async () => {
    const lp = parseInt(view.querySelector('#limite-posts').value, 10) || 5;
    const lc = parseInt(view.querySelector('#limite-cmts').value, 10) || 10;
    ctx.setStatus('Buscando…');
    try {
      const res = await ctx.api.BuscarPosts(lp, lc);
      ctx.setStatus(`${res.posts.length} post(s) carregados`);
      paint();
    } catch (e) {
      ctx.setStatus('Erro: ' + e, true);
    }
  });

  paint();

  async function paint() {
    const data = await ctx.api.ListarPosts();
    const lista = view.querySelector('#lista-posts');
    if (!data.posts || data.posts.length === 0) {
      lista.innerHTML = `<div class="empty">Nenhum post ainda. Clique em <strong>Buscar posts e comentários</strong> para começar.</div>`;
      return;
    }
    lista.innerHTML = data.posts.map(p => `
      <div class="card" data-post="${escapeAttr(p.id)}">
        <div class="row">
          <div>
            <div class="post-title">${escapeText(p.texto_resumo)}</div>
            <div class="post-meta">ID: ${escapeText(p.id)} ${p.created_time ? '• ' + escapeText(p.created_time) : ''}</div>
          </div>
          <span class="spacer"></span>
          <span class="badge pendente">${p.pendentes} pendente(s)</span>
          <span class="badge respondido">${p.total - p.pendentes} respondido(s)</span>
          <button class="btn secondary toggle">Ver comentários</button>
        </div>
        <div class="comentarios" style="display:none;"></div>
      </div>
    `).join('');

    lista.querySelectorAll('.card').forEach(card => {
      const pid = card.getAttribute('data-post');
      const btn = card.querySelector('.toggle');
      const box = card.querySelector('.comentarios');
      btn.addEventListener('click', async () => {
        const open = box.style.display !== 'none';
        if (open) { box.style.display = 'none'; btn.textContent = 'Ver comentários'; return; }
        box.style.display = '';
        btn.textContent = 'Ocultar';
        await renderComments(box, ctx, pid);
      });
    });
  }
}

async function renderComments(box, ctx, postID) {
  box.innerHTML = '<div class="muted">Carregando…</div>';
  const data = await ctx.api.ListarPosts();
  const p = (data.posts || []).find(x => x.id === postID);
  if (!p) { box.innerHTML = '<div class="muted">Post não encontrado.</div>'; return; }

  const pendentes = p.comentarios.filter(c => c.status !== 'respondido');
  const respondidos = p.comentarios.filter(c => c.status === 'respondido');

  const html = `
    <div class="row" style="margin-top:10px;">
      <button class="btn secondary" id="gerar">✨ Gerar sugestões para pendentes</button>
    </div>
    <h4 style="margin:14px 0 4px;">Pendentes (${pendentes.length})</h4>
    ${pendentes.length === 0 ? '<div class="muted">Nada pendente. 🎉</div>' : pendentes.map(c => commentCard(c, postID, true)).join('')}
    <h4 style="margin:14px 0 4px;">Respondidos (${respondidos.length})</h4>
    ${respondidos.length === 0 ? '<div class="muted">Nenhum ainda.</div>' : respondidos.map(c => commentCard(c, postID, false)).join('')}
  `;
  box.innerHTML = html;

  box.querySelector('#gerar').addEventListener('click', async () => {
    ctx.setStatus('Gerando sugestões…');
    try {
      await ctx.api.GerarSugestoes(postID);
      ctx.setStatus('Sugestões geradas');
      await renderComments(box, ctx, postID);
    } catch (e) {
      ctx.setStatus('Erro: ' + e, true);
    }
  });

  box.querySelectorAll('.comment').forEach(node => {
    const cid = node.getAttribute('data-cid');
    const ta = node.querySelector('textarea');
    const btnSalvar = node.querySelector('.salvar');
    const btnEnviar = node.querySelector('.enviar');
    if (ta && btnSalvar) {
      btnSalvar.addEventListener('click', async () => {
        try {
          await ctx.api.EditarResposta(postID, cid, ta.value);
          ctx.setStatus('Rascunho salvo');
        } catch (e) { ctx.setStatus('Erro: ' + e, true); }
      });
    }
    if (ta && btnEnviar) {
      btnEnviar.addEventListener('click', async () => {
        const texto = (ta.value || '').trim();
        if (!texto) {
          ctx.setStatus('Digite uma resposta ou gere uma sugestão antes de enviar.', true);
          ta.focus();
          return;
        }
        if (!confirm(`Publicar esta resposta na Página do Facebook?\n\n"${texto}"`)) return;
        ctx.setStatus('Publicando…');
        try {
          await ctx.api.EnviarResposta(postID, cid, texto);
          ctx.setStatus('✅ Publicado');
          await renderComments(box, ctx, postID);
        } catch (e) {
          ctx.setStatus('Erro ao publicar: ' + (e?.message || e), true);
        }
      });
    }
  });
}

function commentCard(c, postID, editavel) {
  const resposta = c.resposta_final || c.sugestao_ia || '';
  return `
    <div class="comment ${c.status}" data-cid="${escapeAttr(c.id)}">
      <div class="row">
        <strong>${escapeText(c.autor || 'Anônimo')}</strong>
        <span class="muted">• ${escapeText(c.id)}</span>
        <span class="spacer"></span>
        <span class="badge ${c.status}">${c.status}</span>
      </div>
      <div style="margin-top:6px;">${escapeText(c.texto)}</div>
      ${c.sugestao_ia ? `<div class="suggestion">💡 Sugestão da IA: ${escapeText(c.sugestao_ia)}</div>` : ''}
      ${editavel ? `
        <div class="col" style="margin-top:8px;">
          <label>Sua resposta (edite à vontade antes de enviar)</label>
          <textarea data-original="${escapeAttr(resposta)}" placeholder="Digite ou gere uma sugestão acima…">${escapeText(resposta)}</textarea>
          <div class="row">
            <button class="btn secondary salvar">💾 Salvar rascunho</button>
            <button class="btn enviar">📤 Enviar para o Facebook</button>
            <span class="muted" style="font-size:12px;">(você pode digitar e enviar mesmo sem sugestão da IA)</span>
          </div>
        </div>
      ` : `
        <div class="suggestion">✅ Resposta enviada${c.respondido_em ? ' em ' + escapeText(c.respondido_em) : ''}: ${escapeText(c.resposta_final)}</div>
      `}
    </div>
  `;
}

function escapeText(s) {
  return (s || '').replace(/[&<>]/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;'}[c]));
}
function escapeAttr(s) { return escapeText(s).replace(/"/g, '&quot;'); }
