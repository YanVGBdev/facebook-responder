import './style.css';
import { renderConfig } from './screens/Config.js';
import { renderPosts } from './screens/Posts.js';
import { loadApi } from './api.js';

const ctx = {
  api: null,
  currentConfig: null,
  setStatus(msg, isError) {
    const el = document.getElementById('status');
    if (!el) return;
    el.textContent = msg;
    el.style.color = isError ? 'var(--danger)' : 'var(--muted)';
    if (!isError) setTimeout(() => {
      if (el.textContent === msg) el.textContent = 'pronto';
    }, 4000);
  },
};

function setActiveTab(name) {
  document.querySelectorAll('.tab').forEach(b => {
    b.classList.toggle('active', b.dataset.view === name);
  });
  const view = document.getElementById('view');
  if (name === 'config') renderConfig(view, ctx);
  else renderPosts(view, ctx);
}

document.querySelectorAll('.tab').forEach(btn => {
  btn.addEventListener('click', () => setActiveTab(btn.dataset.view));
});

(async function init() {
  ctx.api = await loadApi();
  try {
    ctx.currentConfig = await ctx.api.GetConfig();
  } catch (e) {
    ctx.currentConfig = { usar_mock: true, configurado: false };
  }
  setActiveTab('posts');
})();
