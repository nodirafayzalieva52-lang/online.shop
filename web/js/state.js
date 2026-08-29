import { currentUser } from "./api.js";

const KEY_NEED_ROLE = "kinetiq.need_role";

function uid() {
  const user = currentUser();
  return user ? String(user.id) : "guest";
}

function read(key, fallback) {
  try {
    const raw = localStorage.getItem(key);
    return raw ? JSON.parse(raw) : fallback;
  } catch {
    return fallback;
  }
}

function write(key, value) {
  localStorage.setItem(key, JSON.stringify(value));
}

export function setNeedRole(value) {
  if (value) sessionStorage.setItem(KEY_NEED_ROLE, "1");
  else sessionStorage.removeItem(KEY_NEED_ROLE);
}

export function needsRole() {
  return sessionStorage.getItem(KEY_NEED_ROLE) === "1";
}

export function getCart() {
  return read(`kinetiq.cart.${uid()}`, []);
}

export function setCart(items) {
  write(`kinetiq.cart.${uid()}`, items);
}

export function addToCart(product, qty = 1) {
  const cart = getCart();
  const existing = cart.find((i) => i.product_id === product.id);
  const nextQty = (existing ? existing.quantity : 0) + qty;
  const capped = Math.min(nextQty, Math.max(product.stock || nextQty, 1));
  if (existing) {
    existing.quantity = capped;
    existing.name = product.name;
    existing.price = product.price;
    existing.image_url = product.image_url;
    existing.store_id = product.store_id;
    existing.stock = product.stock;
  } else {
    cart.push({
      product_id: product.id,
      store_id: product.store_id,
      name: product.name,
      price: product.price,
      image_url: product.image_url || "",
      quantity: Math.min(qty, product.stock || qty),
      stock: product.stock,
    });
  }
  setCart(cart);
  return cart;
}

export function updateCartQty(productId, quantity) {
  let cart = getCart();
  if (quantity <= 0) {
    cart = cart.filter((i) => i.product_id !== productId);
  } else {
    const item = cart.find((i) => i.product_id === productId);
    if (item) item.quantity = quantity;
  }
  setCart(cart);
  return cart;
}

export function clearCart() {
  setCart([]);
}

export function cartCount() {
  return getCart().reduce((sum, i) => sum + i.quantity, 0);
}

export function cartTotal() {
  return getCart().reduce((sum, i) => sum + i.price * i.quantity, 0);
}

export function getFavorites() {
  return read(`kinetiq.fav.${uid()}`, []);
}

export function isFavorite(productId) {
  return getFavorites().includes(productId);
}

export function toggleFavorite(productId) {
  const favs = getFavorites();
  const idx = favs.indexOf(productId);
  if (idx >= 0) favs.splice(idx, 1);
  else favs.push(productId);
  write(`kinetiq.fav.${uid()}`, favs);
  return favs;
}
