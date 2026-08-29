const TOKEN_ACCESS = "kinetiq.access";
const TOKEN_REFRESH = "kinetiq.refresh";

export function getAccessToken() {
  return localStorage.getItem(TOKEN_ACCESS) || "";
}

export function getRefreshToken() {
  return localStorage.getItem(TOKEN_REFRESH) || "";
}

export function saveTokens(access, refresh) {
  if (access) localStorage.setItem(TOKEN_ACCESS, access);
  if (refresh) localStorage.setItem(TOKEN_REFRESH, refresh);
}

export function clearTokens() {
  localStorage.removeItem(TOKEN_ACCESS);
  localStorage.removeItem(TOKEN_REFRESH);
}

function decodeJwt(token) {
  try {
    const payload = token.split(".")[1];
    const json = atob(payload.replace(/-/g, "+").replace(/_/g, "/"));
    return JSON.parse(decodeURIComponent(escape(json)));
  } catch {
    return null;
  }
}

export function currentUser() {
  const claims = decodeJwt(getAccessToken());
  if (!claims) return null;
  return {
    id: claims.user_id,
    email: claims.email,
    role: claims.role,
  };
}

async function parseBody(res) {
  const text = await res.text();
  if (!text) return null;
  try {
    return JSON.parse(text);
  } catch {
    return { error: text };
  }
}

let refreshPromise = null;

async function refreshSession() {
  const refresh_token = getRefreshToken();
  if (!refresh_token) return false;
  if (refreshPromise) return refreshPromise;

  refreshPromise = (async () => {
    const res = await fetch("/refresh", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token }),
    });
    const data = await parseBody(res);
    if (!res.ok) {
      clearTokens();
      return false;
    }
    saveTokens(data.access_token, data.refresh_token);
    return true;
  })();

  try {
    return await refreshPromise;
  } finally {
    refreshPromise = null;
  }
}

export async function api(path, options = {}) {
  const headers = { ...(options.headers || {}) };
  if (options.body && !(options.body instanceof FormData)) {
    headers["Content-Type"] = "application/json";
  }
  const token = getAccessToken();
  if (token) headers.Authorization = `Bearer ${token}`;

  let res = await fetch(path, { ...options, headers });
  if (res.status === 401 && getRefreshToken() && path !== "/refresh") {
    const ok = await refreshSession();
    if (ok) {
      headers.Authorization = `Bearer ${getAccessToken()}`;
      res = await fetch(path, { ...options, headers });
    }
  }

  const data = await parseBody(res);
  if (!res.ok) {
    const err = new Error((data && data.error) || `HTTP ${res.status}`);
    err.status = res.status;
    err.data = data;
    throw err;
  }
  return data;
}

export const AuthAPI = {
  register: (email, password) =>
    api("/register", { method: "POST", body: JSON.stringify({ email, password }) }),
  login: (email, password) =>
    api("/login", { method: "POST", body: JSON.stringify({ email, password }) }),
  logout: (refresh_token) =>
    api("/logout", { method: "POST", body: JSON.stringify({ refresh_token }) }),
  setRole: (role) =>
    api("/users/me/role", { method: "PATCH", body: JSON.stringify({ role }) }),
};

export const CatalogAPI = {
  categories: () => api("/categories"),
  stores: () => api("/stores"),
  store: (id) => api(`/stores/${id}`),
  storeBySeller: (sellerId) => api(`/stores/seller/${sellerId}`),
  products: () => api("/products"),
  product: (id) => api(`/products/${id}`),
  productsByStore: (storeId) => api(`/products/store/${storeId}`),
};

export const SellerAPI = {
  createStore: (name, description) =>
    api("/stores", { method: "POST", body: JSON.stringify({ name, description }) }),
  createProduct: (payload) =>
    api("/products", { method: "POST", body: JSON.stringify(payload) }),
  updateProduct: (id, payload) =>
    api(`/products/${id}`, { method: "PUT", body: JSON.stringify(payload) }),
  storeOrders: (storeId) => api(`/orders/store/${storeId}`),
};

export const OrderAPI = {
  create: (items) =>
    api("/orders", { method: "POST", body: JSON.stringify({ items }) }),
  mine: () => api("/orders"),
  one: (id) => api(`/orders/${id}`),
};
