import { currentUser } from "./api.js";
import { cartCount, isFavorite } from "./state.js";
import {
  escapeHtml,
  formatPrice,
  formatDate,
  categoryName,
  storeName,
  statusClass,
  statusLabel,
  initials,
} from "./ui.js";

export function authView(tab = "login") {
  const isLogin = tab === "login";
  return `
    <div class="auth-shell">
      <div class="auth-left">
        <div class="auth-brand">KINETIQ<span>.SHOP</span></div>
        <p class="auth-tagline">Кибермаркет будущего. Плотная витрина, мгновенный заказ, контроль продавца.</p>
      </div>
      <div class="auth-right">
        <div class="auth-tabs">
          <div class="auth-tab ${isLogin ? "active" : ""}" data-action="auth-tab" data-tab="login">Вход</div>
          <div class="auth-tab ${!isLogin ? "active" : ""}" data-action="auth-tab" data-tab="register">Регистрация</div>
        </div>
        <form class="auth-form" id="auth-form" data-mode="${isLogin ? "login" : "register"}">
          <div class="form-group">
            <label class="form-label">Email</label>
            <input class="form-input" type="email" name="email" placeholder="you@kinetiq.shop" required />
          </div>
          <div class="form-group">
            <label class="form-label">Пароль</label>
            <input class="form-input" type="password" name="password" placeholder="минимум 6 символов" minlength="6" required />
          </div>
          <p class="form-error hidden" id="auth-error"></p>
          <button class="btn btn-primary btn-full" type="submit">${isLogin ? "Войти в сеть" : "Создать аккаунт"}</button>
          <p class="form-note">${isLogin ? "Нет аккаунта — переключитесь на регистрацию." : "После регистрации выберите роль: покупатель или продавец."}</p>
        </form>
      </div>
    </div>
  `;
}

export function roleView() {
  return `
    <div class="role-shell">
      <div class="auth-brand" style="font-size:2rem;margin-bottom:24px">KINETIQ<span>.SHOP</span></div>
      <h1 class="role-title">Кто вы в Kinetiq?</h1>
      <p class="role-sub">Роль можно выбрать один раз. Продавец не сможет стать покупателем.</p>
      <div class="role-cards">
        <div class="role-card buyer" data-action="pick-role" data-role="customer">
          <span class="role-icon">◈</span>
          <h3>Покупатель</h3>
          <p>Каталог, избранное, корзина и оформление заказов в магазинах площадки.</p>
        </div>
        <div class="role-card seller" data-action="pick-role" data-role="seller">
          <span class="role-icon">⬡</span>
          <h3>Продавец</h3>
          <p>Свой магазин, витрина товаров, остатки и входящие заказы.</p>
        </div>
      </div>
    </div>
  `;
}

export function shell(inner, { cartOpen = false } = {}) {
  const user = currentUser();
  const seller = user && (user.role === "seller" || user.role === "admin");
  return `
    <div class="app-shell">
      <header class="navbar">
        <div class="navbar-logo" data-action="nav" data-route="#/catalog">KINETIQ<span>.SHOP</span></div>
        <nav class="navbar-nav">
          <a class="nav-link" data-action="nav" data-route="#/catalog">Каталог</a>
          <a class="nav-link" data-action="nav" data-route="#/favorites">Избранное</a>
          <a class="nav-link" data-action="nav" data-route="#/orders">Заказы</a>
          ${seller ? `<a class="nav-link" data-action="nav" data-route="#/seller">Кабинет</a>` : ""}
        </nav>
        <div class="navbar-spacer"></div>
        <div class="navbar-actions">
          <div class="navbar-search">
            <input id="global-search" type="search" placeholder="Поиск по витрине..." />
            <span class="search-icon">⌕</span>
          </div>
          <button class="icon-btn" data-action="open-cart" title="Корзина" type="button">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M6 6h15l-1.5 9h-12z"/><circle cx="9" cy="20" r="1.5"/><circle cx="18" cy="20" r="1.5"/><path d="M6 6L5 3H2"/></svg>
            <span class="badge" id="cart-badge">${cartCount()}</span>
          </button>
          <button class="avatar-btn" data-action="logout" type="button" title="Выйти">
            <span class="avatar-circle">${escapeHtml(initials(user?.email))}</span>
            <span class="user-email">${escapeHtml(user?.email || "")}</span>
          </button>
        </div>
      </header>
      <main id="app">${inner}</main>
    </div>
    <div class="cart-overlay ${cartOpen ? "open" : ""}" data-action="close-cart"></div>
    <aside class="cart-drawer ${cartOpen ? "open" : ""}" id="cart-drawer"></aside>
  `;
}

export function productCard(p, categories) {
  const fav = isFavorite(p.id);
  const out = !p.stock;
  return `
    <article class="product-card" data-action="open-product" data-id="${p.id}">
      <div class="product-img-wrap">
        ${p.image_url ? `<img class="product-img" src="${escapeHtml(p.image_url)}" alt="${escapeHtml(p.name)}" />` : `<div class="product-img-placeholder">◆</div>`}
        <div class="product-img-overlay"></div>
        <button class="fav-btn ${fav ? "active" : ""}" type="button" data-action="toggle-fav" data-id="${p.id}" title="В избранное">${fav ? "♥" : "♡"}</button>
        <span class="stock-badge ${out ? "out-stock" : "in-stock"}">${out ? "Нет в наличии" : "В наличии"}</span>
      </div>
      <div class="product-body">
        <span class="product-cat-tag">${escapeHtml(categoryName(categories, p.category_id))}</span>
        <div class="product-name">${escapeHtml(p.name)}</div>
        <div class="product-price">${formatPrice(p.price)}</div>
        <div class="product-actions">
          <button class="add-cart-btn" type="button" data-action="add-cart" data-id="${p.id}" ${out ? "disabled" : ""}>В корзину</button>
        </div>
      </div>
    </article>
  `;
}

export function catalogView({ products, categories, stores, filters }) {
  const catItems = [`<div class="cat-item ${!filters.category ? "active" : ""}" data-action="filter-cat" data-id="">Все категории</div>`]
    .concat(
      (categories || []).map(
        (c) =>
          `<div class="cat-item ${Number(filters.category) === c.id ? "active" : ""}" data-action="filter-cat" data-id="${c.id}">
            ${escapeHtml(c.name)}
            <span class="cat-count">${products.filter((p) => p.category_id === c.id).length}</span>
          </div>`
      )
    )
    .join("");

  const storeItems = [`<div class="cat-item ${!filters.store ? "active" : ""}" data-action="filter-store" data-id="">Все магазины</div>`]
    .concat(
      (stores || []).map(
        (s) =>
          `<div class="cat-item ${Number(filters.store) === s.id ? "active" : ""}" data-action="filter-store" data-id="${s.id}">${escapeHtml(s.name)}</div>`
      )
    )
    .join("");

  const grid =
    products.length === 0
      ? `<div class="empty-state"><div class="empty-state-icon">◇</div><div class="empty-state-title">Витрина пуста</div><div class="empty-state-sub">Измените фильтры или подождите появления товаров.</div></div>`
      : `<div class="products-grid">${products.map((p) => productCard(p, categories)).join("")}</div>`;

  return `
    <div class="page-wide">
      <section class="hero-banner">
        <h1 class="hero-title">Рынок <em>Kinetiq</em></h1>
        <p class="hero-sub">Киберпанк-витрина в духе маркетплейса: плотная сетка, быстрые фильтры, избранное и корзина.</p>
      </section>
      <div class="catalog-layout">
        <aside class="catalog-sidebar">
          <div class="sidebar-card">
            <div class="sidebar-title">Категории</div>
            <div class="cat-list">${catItems}</div>
          </div>
          <div class="sidebar-card">
            <div class="sidebar-title">Магазины</div>
            <div class="cat-list">${storeItems}</div>
          </div>
        </aside>
        <section class="catalog-main">
          <div class="catalog-header">
            <div>
              <h2 class="catalog-title">Каталог</h2>
              <div class="catalog-meta">${products.length} товаров</div>
            </div>
            <select class="sort-select" id="sort-select">
              <option value="new" ${filters.sort === "new" ? "selected" : ""}>Сначала новые</option>
              <option value="price-asc" ${filters.sort === "price-asc" ? "selected" : ""}>Цена ↑</option>
              <option value="price-desc" ${filters.sort === "price-desc" ? "selected" : ""}>Цена ↓</option>
              <option value="name" ${filters.sort === "name" ? "selected" : ""}>По названию</option>
            </select>
          </div>
          ${grid}
        </section>
      </div>
    </div>
  `;
}

export function productDetailView(p, categories, stores, qty = 1) {
  const fav = isFavorite(p.id);
  const out = !p.stock;
  return `
    <div class="page">
      <div class="breadcrumb">
        <a data-action="nav" data-route="#/catalog">Каталог</a>
        <span class="breadcrumb-sep">/</span>
        <span class="breadcrumb-current">${escapeHtml(p.name)}</span>
      </div>
      <div class="product-detail">
        <div class="product-detail-img" style="position:relative">
          ${p.image_url ? `<img src="${escapeHtml(p.image_url)}" alt="${escapeHtml(p.name)}" />` : "◆"}
          <button class="fav-btn ${fav ? "active" : ""}" type="button" data-action="toggle-fav" data-id="${p.id}" style="top:16px;right:16px">${fav ? "♥" : "♡"}</button>
        </div>
        <div class="product-detail-info">
          <span class="product-cat-tag product-detail-category">${escapeHtml(categoryName(categories, p.category_id))}</span>
          <h1 class="product-detail-name">${escapeHtml(p.name)}</h1>
          <div class="product-detail-price">${formatPrice(p.price)}</div>
          <p class="product-detail-desc">${escapeHtml(p.description || "Описание появится позже.")}</p>
          <div class="product-detail-meta">
            <div class="meta-item">
              <div class="meta-label">Магазин</div>
              <div class="meta-value">${escapeHtml(storeName(stores, p.store_id))}</div>
            </div>
            <div class="meta-item">
              <div class="meta-label">Остаток</div>
              <div class="meta-value">${p.stock} шт.</div>
            </div>
          </div>
          <div class="qty-control">
            <button class="qty-btn" type="button" data-action="qty" data-delta="-1">−</button>
            <span class="qty-display" id="detail-qty">${qty}</span>
            <button class="qty-btn" type="button" data-action="qty" data-delta="1">+</button>
          </div>
          <div class="detail-actions">
            <button class="btn btn-primary" type="button" data-action="add-cart" data-id="${p.id}" ${out ? "disabled" : ""}>В корзину</button>
            <button class="btn btn-ghost" type="button" data-action="toggle-fav" data-id="${p.id}">${fav ? "Убрать из избранного" : "В избранное"}</button>
          </div>
        </div>
      </div>
    </div>
  `;
}

export function favoritesView(products, categories) {
  return `
    <div class="page">
      <div class="favorites-header">
        <h1 class="favorites-title">Избранное</h1>
        <span class="fav-count">${products.length}</span>
      </div>
      ${
        products.length
          ? `<div class="products-grid">${products.map((p) => productCard(p, categories)).join("")}</div>`
          : `<div class="empty-state"><div class="empty-state-icon">♡</div><div class="empty-state-title">Пока пусто</div><div class="empty-state-sub">Нажмите сердечко на карточке товара, чтобы сохранить его здесь.</div></div>`
      }
    </div>
  `;
}

export function ordersView(orders) {
  if (!orders?.length) {
    return `
      <div class="page">
        <h1 class="catalog-title mb-24">Мои заказы</h1>
        <div class="empty-state"><div class="empty-state-icon">▤</div><div class="empty-state-title">Заказов нет</div><div class="empty-state-sub">Соберите корзину и оформите первый заказ.</div></div>
      </div>`;
  }
  const cards = orders
    .map((o) => {
      const items = (o.items || [])
        .map(
          (it) => `
          <div class="order-item-row">
            <span class="order-item-name">${escapeHtml(it.product?.name || "Товар #" + it.product_id)}</span>
            <span class="order-item-qty">× ${it.quantity}</span>
            <span class="order-item-price">${formatPrice(it.price * it.quantity)}</span>
          </div>`
        )
        .join("");
      return `
        <article class="order-card">
          <div class="order-card-header">
            <div>
              <div class="order-id">Заказ #${o.id}</div>
              <div class="order-date">${formatDate(o.created_at)}</div>
            </div>
            <span class="status-badge ${statusClass(o.status)}">${statusLabel(o.status)}</span>
          </div>
          <div class="order-items-list">${items}</div>
          <div class="order-total-row">
            <span class="order-total-label">Итого</span>
            <span class="order-total-value">${formatPrice(o.total_price)}</span>
          </div>
        </article>`;
    })
    .join("");
  return `<div class="page"><h1 class="catalog-title mb-24">Мои заказы</h1><div class="orders-list">${cards}</div></div>`;
}

export function cartDrawerView(items) {
  if (!items.length) {
    return `
      <div class="cart-header"><h3>Корзина</h3><button class="cart-close" type="button" data-action="close-cart">✕</button></div>
      <div class="cart-empty"><div class="cart-empty-icon">◇</div><div>Корзина пуста</div></div>
    `;
  }
  const rows = items
    .map(
      (i) => `
      <div class="cart-item">
        <div class="cart-item-img">${i.image_url ? `<img src="${escapeHtml(i.image_url)}" alt="" />` : "◆"}</div>
        <div class="cart-item-info">
          <div class="cart-item-name">${escapeHtml(i.name)}</div>
          <div class="cart-item-price">${formatPrice(i.price)}</div>
          <div class="cart-item-controls">
            <button class="cart-qty-btn" type="button" data-action="cart-qty" data-id="${i.product_id}" data-delta="-1">−</button>
            <span class="cart-item-qty">${i.quantity}</span>
            <button class="cart-qty-btn" type="button" data-action="cart-qty" data-id="${i.product_id}" data-delta="1">+</button>
            <span class="cart-remove" data-action="cart-remove" data-id="${i.product_id}">Удалить</span>
          </div>
        </div>
      </div>`
    )
    .join("");
  const total = items.reduce((s, i) => s + i.price * i.quantity, 0);
  return `
    <div class="cart-header"><h3>Корзина</h3><button class="cart-close" type="button" data-action="close-cart">✕</button></div>
    <div class="cart-items">${rows}</div>
    <div class="cart-footer">
      <div class="cart-total-row">
        <span class="cart-total-label">К оплате</span>
        <span class="cart-total-value">${formatPrice(total)}</span>
      </div>
      <p class="form-note mb-16">Заказ с товарами из разных магазинов будет разбит автоматически.</p>
      <button class="btn btn-primary btn-full" type="button" data-action="checkout">Оформить заказ</button>
    </div>
  `;
}

export function sellerStoreSetup() {
  return `
    <div class="page">
      <div class="store-setup-wrap">
        <div class="store-setup-icon">⬡</div>
        <h1 class="store-setup-title">Создайте свой магазин</h1>
        <p class="store-setup-sub">У продавца один магазин. После создания вы сможете выставлять товары.</p>
        <form class="form-card store-form-card" id="store-form">
          <div class="form-group mb-16">
            <label class="form-label">Название</label>
            <input class="form-input" name="name" required placeholder="Neon Forge" />
          </div>
          <div class="form-group mb-16">
            <label class="form-label">Описание</label>
            <textarea class="form-input" name="description" placeholder="Чем торгуете"></textarea>
          </div>
          <p class="form-error hidden" id="store-error"></p>
          <button class="btn btn-violet btn-full" type="submit">Открыть магазин</button>
        </form>
      </div>
    </div>
  `;
}

export function sellerDashboard({ store, products, orders, categories }) {
  const revenue = (orders || []).reduce((s, o) => s + (o.total_price || 0), 0);
  const rows = (products || [])
    .map(
      (p) => `
      <tr>
        <td>${escapeHtml(p.name)}</td>
        <td>${escapeHtml(categoryName(categories, p.category_id))}</td>
        <td>${formatPrice(p.price)}</td>
        <td>${p.stock}</td>
        <td>
          <div class="table-actions">
            <button class="btn btn-ghost btn-sm" type="button" data-action="nav" data-route="#/seller/product/${p.id}">Изменить</button>
          </div>
        </td>
      </tr>`
    )
    .join("");

  const orderCards = (orders || [])
    .slice(0, 8)
    .map(
      (o) => `
      <article class="order-card">
        <div class="order-card-header">
          <div>
            <div class="order-id">Заказ #${o.id}</div>
            <div class="order-date">${formatDate(o.created_at)} · покупатель #${o.customer_id}</div>
          </div>
          <span class="status-badge ${statusClass(o.status)}">${statusLabel(o.status)}</span>
        </div>
        <div class="order-total-row">
          <span class="order-total-label">${(o.items || []).length} позиций</span>
          <span class="order-total-value">${formatPrice(o.total_price)}</span>
        </div>
      </article>`
    )
    .join("");

  return `
    <div class="page">
      <div class="seller-header">
        <div>
          <h1 class="seller-title">${escapeHtml(store.name)}</h1>
          <p class="seller-sub">${escapeHtml(store.description || "Панель продавца Kinetiq.Shop")}</p>
        </div>
        <button class="btn btn-violet" type="button" data-action="nav" data-route="#/seller/product">+ Новый товар</button>
      </div>
      <div class="stats-grid">
        <div class="stat-card cyan"><div class="stat-label">Товары</div><div class="stat-value">${products.length}</div></div>
        <div class="stat-card violet"><div class="stat-label">Заказы</div><div class="stat-value">${orders.length}</div></div>
        <div class="stat-card green"><div class="stat-label">Оборот</div><div class="stat-value" style="font-size:1.3rem">${formatPrice(revenue)}</div></div>
        <div class="stat-card orange"><div class="stat-label">Остаток SKU</div><div class="stat-value">${products.reduce((s, p) => s + (p.stock || 0), 0)}</div></div>
      </div>
      <div class="data-table-wrap mb-24">
        <div class="data-table-header">
          <div class="data-table-title">Витрина магазина</div>
        </div>
        <div style="overflow-x:auto">
          <table class="data-table">
            <thead><tr><th>Наименование</th><th>Категория</th><th>Цена</th><th>Сток</th><th></th></tr></thead>
            <tbody>${rows || `<tr><td colspan="5" class="text-muted">Товаров пока нет</td></tr>`}</tbody>
          </table>
        </div>
      </div>
      <h2 class="data-table-title mb-16">Входящие заказы</h2>
      <div class="orders-list">${orderCards || `<div class="empty-state"><div class="empty-state-title">Заказов нет</div></div>`}</div>
    </div>
  `;
}

export function productFormView({ product, categories, store }) {
  const p = product || {};
  const opts = (categories || [])
    .map((c) => `<option value="${c.id}" ${c.id === p.category_id ? "selected" : ""}>${escapeHtml(c.name)}</option>`)
    .join("");
  return `
    <div class="page">
      <div class="breadcrumb">
        <a data-action="nav" data-route="#/seller">Кабинет</a>
        <span class="breadcrumb-sep">/</span>
        <span class="breadcrumb-current">${p.id ? "Редактирование" : "Новый товар"}</span>
      </div>
      <form class="form-card" id="product-form" data-id="${p.id || ""}">
        <h2 class="form-card-title">${p.id ? "Изменить товар" : "Добавить товар"}</h2>
        <p class="text-sm text-muted mb-16">Магазин: <strong class="text-cyan">${escapeHtml(store.name)}</strong></p>
        <div class="form-grid">
          <div class="form-group form-col-span">
            <label class="form-label">Наименование</label>
            <input class="form-input" name="name" required value="${escapeHtml(p.name || "")}" />
          </div>
          <div class="form-group form-col-span">
            <label class="form-label">Описание</label>
            <textarea class="form-input" name="description">${escapeHtml(p.description || "")}</textarea>
          </div>
          <div class="form-group">
            <label class="form-label">Цена, ₽</label>
            <input class="form-input" name="price" type="number" min="0.01" step="0.01" required value="${p.price ?? ""}" />
          </div>
          <div class="form-group">
            <label class="form-label">Остаток</label>
            <input class="form-input" name="stock" type="number" min="0" step="1" required value="${p.stock ?? 0}" />
          </div>
          <div class="form-group">
            <label class="form-label">Категория</label>
            <select class="form-input" name="category_id">${opts}</select>
          </div>
          <div class="form-group">
            <label class="form-label">Ссылка на фото (https)</label>
            <input class="form-input" name="image_url" type="url" placeholder="https://..." value="${escapeHtml(p.image_url || "")}" />
          </div>
        </div>
        <p class="form-error hidden mt-16" id="product-error"></p>
        <div class="modal-actions">
          <button class="btn btn-ghost" type="button" data-action="nav" data-route="#/seller">Отмена</button>
          <button class="btn btn-primary" type="submit">Сохранить</button>
        </div>
      </form>
    </div>
  `;
}

export function loadingView() {
  return `<div class="loading-center"><div class="spinner"></div></div>`;
}
