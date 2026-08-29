// Carrega os bindings gerados pelo Wails (frontend/wailsjs/go/main/App.js).
// Em preview puro (sem Wails), devolve um stub que avisa o usuário.
export async function loadApi() {
  try {
    const mod = await import('../wailsjs/go/main/App.js');
    return mod;
  } catch (e) {
    const noop = async () => {
      throw new Error('Backend Wails indisponível — rode `wails dev` ou abra o app compilado.');
    };
    return new Proxy({}, { get: () => noop });
  }
}
