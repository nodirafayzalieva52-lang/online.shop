export function escapeHtml(value) {
  return String(value ?? "")
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#39;");
}

export function formatPrice(n) {
  const num = Number(n) || 0;
  return (
    new Intl.NumberFormat("ru-RU", {
      minimumFractionDigits: 0,
      maximumFractionDigits: 2,
    }).format(num) + " ₽"
  );
}

export function formatDate(iso) {
  if (!iso) return "";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "";
  return d.toLocaleString("ru-RU", {
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function toast(message, type = "info") {
  const box = document.getElementById("toasts");
  if (!box) return;
  const el = document.createElement("div");
  el.className = `toast ${type}`;
  el.textContent = message;
  box.appendChild(el);
  setTimeout(() => {
    el.style.animation = "toastOut 0.25s ease forwards";
    setTimeout(() => el.remove(), 250);
  }, 2800);
}

export function categoryName(categories, id) {
  const cat = (categories || []).find((c) => c.id === id);
  return cat ? cat.name : "Без категории";
}

export function storeName(stores, id) {
  const s = (stores || []).find((x) => x.id === id);
  return s ? s.name : `Магазин #${id}`;
}

export function imgTag(url, alt, cls) {
  if (url) {
    return `<img class="${cls || ""}" src="${escapeHtml(url)}" alt="${escapeHtml(alt)}" onerror="this.replaceWith(Object.assign(document.createElement('div'),{className:'product-img-placeholder',textContent:'◆'}))" />`;
  }
  return `<div class="product-img-placeholder">◆</div>`;
}

export function statusClass(status) {
  if (status === "paid") return "status-paid";
  if (status === "cancelled") return "status-cancelled";
  return "status-pending";
}

export function statusLabel(status) {
  const map = { pending: "Ожидает", paid: "Оплачен", cancelled: "Отменён" };
  return map[status] || status;
}

export function initials(email) {
  const name = (email || "?").split("@")[0];
  return name.slice(0, 2).toUpperCase();
}
