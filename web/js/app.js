import {
  AuthAPI,
  CatalogAPI,
  OrderAPI,
  SellerAPI,
  clearTokens,
  currentUser,
  getRefreshToken,
  saveTokens,
} from "./api.js";
import {
  addToCart,
  cartCount,
  clearCart,
  getCart,
  getFavorites,
  needsRole,
  setNeedRole,
  toggleFavorite,
  updateCartQty,
} from "./state.js";
import { toast } from "./ui.js";
import {
  authView,
  cartDrawerView,
  catalogView,
  favoritesView,
  loadingView,
  ordersView,
  productDetailView,
  productFormView,
  roleView,
  sellerDashboard,
  sellerStoreSetup,
  shell,
} from "./views.js";

const root = document.getElementById("root");

const cache = {
  products: null,
  categories: null,
  stores: null,
};

let cartOpen = false;
let detailQty = 1;
let filters = { category: "", store: "", sort: "new", q: "" };

function route() {
  const raw = (location.hash || "#/").replace(/^#/, "") || "/";
  const parts = raw.split("/").filter(Boolean);
  return { path: "/" + parts.join("/"), parts };
}

function go(hash) {
  if (location.hash === hash) render();
  else location.hash = hash;
}

function isAuthed() {
  return Boolean(currentUser());
}

async function loadCatalogData() {
  const [products, categories, stores] = await Promise.all([
    cache.products ? Promise.resolve(cache.products) : CatalogAPI.products(),
    cache.categories ? Promise.resolve(cache.categories) : CatalogAPI.categories(),
    cache.stores ? Promise.resolve(cache.stores) : CatalogAPI.stores(),
  ]);
  cache.products = products || [];
  cache.categories = categories || [];
  cache.stores = stores || [];
  return cache;
}

function applyFilters(products) {
  let list = [...products];
  if (filters.category) list = list.filter((p) => p.category_id === Number(filters.category));
  if (filters.store) list = list.filter((p) => p.store_id === Number(filters.store));
  if (filters.q) {
    const q = filters.q.toLowerCase();
    list = list.filter((p) => (p.name || "").toLowerCase().includes(q) || (p.description || "").toLowerCase().includes(q));
  }
  if (filters.sort === "price-asc") list.sort((a, b) => a.price - b.price);
  else if (filters.sort === "price-desc") list.sort((a, b) => b.price - a.price);
  else if (filters.sort === "name") list.sort((a, b) => a.name.localeCompare(b.name, "ru"));
  else list.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
  return list;
}

function paintCart() {
  const drawer = document.getElementById("cart-drawer");
  if (drawer) drawer.innerHTML = cartDrawerView(getCart());
  const badge = document.getElementById("cart-badge");
  if (badge) badge.textContent = String(cartCount());
}

function mountShell(inner) {
  root.innerHTML = shell(inner, { cartOpen });
  paintCart();
  const search = document.getElementById("global-search");
  if (search) {
    search.value = filters.q;
    search.addEventListener("keydown", (e) => {
      if (e.key === "Enter") {
        filters.q = search.value.trim();
        if ((location.hash || "#/catalog") === "#/catalog") render();
        else go("#/catalog");
      }
    });
  }
  document.querySelectorAll(".nav-link").forEach((el) => {
    const r = el.dataset.route;
    if (r && location.hash.startsWith(r)) el.classList.add("active");
  });
}

async function render() {
  const { path, parts } = route();
  const user = currentUser();

  if (!user) {
    if (path.startsWith("/register")) root.innerHTML = authView("register");
    else root.innerHTML = authView("login");
    bind();
    return;
  }

  if (needsRole()) {
    root.innerHTML = roleView();
    bind();
    return;
  }

  try {
    if (parts[0] === "product" && parts[1]) {
      mountShell(loadingView());
      const [{ product, categories, stores }] = await Promise.all([
        loadCatalogData().then(async (data) => {
          const product = data.products.find((p) => p.id === Number(parts[1])) || (await CatalogAPI.product(parts[1]));
          return { product, categories: data.categories, stores: data.stores };
        }),
      ]);
      detailQty = 1;
      mountShell(productDetailView(product, categories, stores, detailQty));
    } else if (path === "/favorites") {
      const data = await loadCatalogData();
      const ids = getFavorites();
      const products = data.products.filter((p) => ids.includes(p.id));
      mountShell(favoritesView(products, data.categories));
    } else if (path === "/orders") {
      mountShell(loadingView());
      const orders = await OrderAPI.mine();
      mountShell(ordersView(orders || []));
    } else if (parts[0] === "seller") {
      await renderSeller(parts);
    } else {
      const data = await loadCatalogData();
      const products = applyFilters(data.products);
      mountShell(catalogView({ products, categories: data.categories, stores: data.stores, filters }));
    }
  } catch (err) {
    mountShell(`<div class="page"><div class="empty-state"><div class="empty-state-title">Ошибка</div><div class="empty-state-sub">${err.message}</div></div></div>`);
  }
  bind();
}

async function renderSeller(parts) {
  const user = currentUser();
  if (user.role !== "seller" && user.role !== "admin") {
    go("#/catalog");
    toast("Кабинет доступен продавцам", "error");
    return;
  }

  mountShell(loadingView());
  let store = null;
  try {
    store = await CatalogAPI.storeBySeller(user.id);
  } catch (err) {
    if (err.status === 404) {
      mountShell(sellerStoreSetup());
      return;
    }
    throw err;
  }

  const [products, categories, orders] = await Promise.all([
    CatalogAPI.productsByStore(store.id),
    cache.categories || CatalogAPI.categories(),
    SellerAPI.storeOrders(store.id).catch(() => []),
  ]);
  cache.categories = categories || [];

  if (parts[1] === "product") {
    let product = null;
    if (parts[2]) {
      product = (products || []).find((p) => p.id === Number(parts[2])) || (await CatalogAPI.product(parts[2]));
    }
    mountShell(productFormView({ product, categories: cache.categories, store }));
    return;
  }

  mountShell(sellerDashboard({ store, products: products || [], orders: orders || [], categories: cache.categories }));
}

function bind() {
  root.onclick = onClick;
  const authForm = document.getElementById("auth-form");
  if (authForm) authForm.onsubmit = onAuth;
  const storeForm = document.getElementById("store-form");
  if (storeForm) storeForm.onsubmit = onCreateStore;
  const productForm = document.getElementById("product-form");
  if (productForm) productForm.onsubmit = onSaveProduct;
  const sort = document.getElementById("sort-select");
  if (sort) {
    sort.onchange = () => {
      filters.sort = sort.value;
      render();
    };
  }
}

async function onAuth(e) {
  e.preventDefault();
  const form = e.target;
  const email = form.email.value.trim();
  const password = form.password.value;
  const errorEl = document.getElementById("auth-error");
  errorEl.classList.add("hidden");
  try {
    if (form.dataset.mode === "register") {
      await AuthAPI.register(email, password);
      const tokens = await AuthAPI.login(email, password);
      saveTokens(tokens.access_token, tokens.refresh_token);
      setNeedRole(true);
      toast("Аккаунт создан. Выберите роль.", "success");
      render();
      return;
    }
    const tokens = await AuthAPI.login(email, password);
    saveTokens(tokens.access_token, tokens.refresh_token);
    setNeedRole(false);
    toast("Добро пожаловать в Kinetiq", "success");
    go("#/catalog");
  } catch (err) {
    errorEl.textContent = err.message;
    errorEl.classList.remove("hidden");
  }
}

async function onCreateStore(e) {
  e.preventDefault();
  const form = e.target;
  const errorEl = document.getElementById("store-error");
  try {
    await SellerAPI.createStore(form.name.value.trim(), form.description.value.trim());
    cache.stores = null;
    toast("Магазин открыт", "success");
    render();
  } catch (err) {
    errorEl.textContent = err.message;
    errorEl.classList.remove("hidden");
  }
}

async function onSaveProduct(e) {
  e.preventDefault();
  const form = e.target;
  const errorEl = document.getElementById("product-error");
  const user = currentUser();
  try {
    const store = await CatalogAPI.storeBySeller(user.id);
    const payload = {
      store_id: store.id,
      category_id: Number(form.category_id.value) || 0,
      name: form.name.value.trim(),
      description: form.description.value.trim(),
      price: Number(form.price.value),
      stock: Number(form.stock.value),
      image_url: form.image_url.value.trim(),
    };
    if (form.dataset.id) {
      await SellerAPI.updateProduct(form.dataset.id, payload);
      toast("Товар обновлён", "success");
    } else {
      await SellerAPI.createProduct(payload);
      toast("Товар выставлен", "success");
    }
    cache.products = null;
    go("#/seller");
  } catch (err) {
    errorEl.textContent = err.message;
    errorEl.classList.remove("hidden");
  }
}

async function onClick(e) {
  const t = e.target.closest("[data-action]");
  if (!t) return;
  const action = t.dataset.action;
  e.preventDefault();

  if (action === "auth-tab") {
    root.innerHTML = authView(t.dataset.tab);
    bind();
    return;
  }
  if (action === "nav") {
    cartOpen = false;
    go(t.dataset.route);
    return;
  }
  if (action === "logout") {
    try {
      await AuthAPI.logout(getRefreshToken());
    } catch {
      /* ignore */
    }
    clearTokens();
    setNeedRole(false);
    cartOpen = false;
    toast("Сессия закрыта", "info");
    go("#/login");
    render();
    return;
  }
  if (action === "pick-role") {
    try {
      const tokens = await AuthAPI.setRole(t.dataset.role);
      saveTokens(tokens.access_token, tokens.refresh_token);
      setNeedRole(false);
      toast(t.dataset.role === "seller" ? "Режим продавца включён" : "Режим покупателя включён", "success");
      go(t.dataset.role === "seller" ? "#/seller" : "#/catalog");
    } catch (err) {
      toast(err.message, "error");
    }
    return;
  }
  if (action === "filter-cat") {
    filters.category = t.dataset.id;
    render();
    return;
  }
  if (action === "filter-store") {
    filters.store = t.dataset.id;
    render();
    return;
  }
  if (action === "open-product") {
    if (e.target.closest("[data-action='toggle-fav'], [data-action='add-cart']")) return;
    go(`#/product/${t.dataset.id}`);
    return;
  }
  if (action === "toggle-fav") {
    toggleFavorite(Number(t.dataset.id));
    render();
    return;
  }
  if (action === "add-cart") {
    const id = Number(t.dataset.id);
    const product = (cache.products || []).find((p) => p.id === id);
    const qty = document.getElementById("detail-qty") ? detailQty : 1;
    if (!product) {
      try {
        const p = await CatalogAPI.product(id);
        addToCart(p, qty);
      } catch (err) {
        toast(err.message, "error");
        return;
      }
    } else {
      addToCart(product, qty);
    }
    paintCart();
    toast("Добавлено в корзину", "success");
    return;
  }
  if (action === "qty") {
    const productId = Number(location.hash.split("/").pop());
    const product = (cache.products || []).find((p) => p.id === productId);
    const max = product?.stock || 99;
    detailQty = Math.max(1, Math.min(max, detailQty + Number(t.dataset.delta)));
    const el = document.getElementById("detail-qty");
    if (el) el.textContent = String(detailQty);
    return;
  }
  if (action === "open-cart") {
    cartOpen = true;
    document.querySelector(".cart-overlay")?.classList.add("open");
    document.getElementById("cart-drawer")?.classList.add("open");
    paintCart();
    return;
  }
  if (action === "close-cart") {
    cartOpen = false;
    document.querySelector(".cart-overlay")?.classList.remove("open");
    document.getElementById("cart-drawer")?.classList.remove("open");
    return;
  }
  if (action === "cart-qty") {
    const id = Number(t.dataset.id);
    const item = getCart().find((i) => i.product_id === id);
    if (!item) return;
    updateCartQty(id, item.quantity + Number(t.dataset.delta));
    paintCart();
    return;
  }
  if (action === "cart-remove") {
    updateCartQty(Number(t.dataset.id), 0);
    paintCart();
    return;
  }
  if (action === "checkout") {
    await checkout();
  }
}

async function checkout() {
  const cart = getCart();
  if (!cart.length) return;
  const byStore = new Map();
  for (const item of cart) {
    const list = byStore.get(item.store_id) || [];
    list.push({ product_id: item.product_id, quantity: item.quantity });
    byStore.set(item.store_id, list);
  }
  try {
    for (const items of byStore.values()) {
      await OrderAPI.create(items);
    }
    clearCart();
    cartOpen = false;
    cache.products = null;
    toast("Заказ оформлен", "success");
    go("#/orders");
  } catch (err) {
    toast(err.message, "error");
  }
}

window.addEventListener("hashchange", render);
render();
