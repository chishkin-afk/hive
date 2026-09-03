document.addEventListener('DOMContentLoaded', () => {
  loadTrialStatus();
  loadRates();
  loadHistory();
});

const STATUS_MAP = {
  0: { label: 'Неизвестно', badge: 'badge-unknown' },
  1: { label: 'Ожидает',    badge: 'badge-pending' },
  2: { label: 'Оплачен',    badge: 'badge-confirmed' },
  3: { label: 'Отменён',    badge: 'badge-canceled' },
};

function statusLabel(statusInt) {
  return (STATUS_MAP[statusInt] || {}).label || '—';
}

function badgeClassByStatus(statusInt) {
  return (STATUS_MAP[statusInt] || {}).badge || 'badge-unknown';
}

function dailyPrice(penny, days) {
  if (!days) return '—';
  const daily = penny / days / 100;
  return '~' + Math.max(1, Math.round(daily)) + ' ₽ в день';
}

function fmtDate(ts) {
  if (!ts) return '—';
  const d = new Date(ts * 1000);
  const day = String(d.getDate()).padStart(2, '0');
  const month = String(d.getMonth() + 1).padStart(2, '0');
  return day + '.' + month + '.' + d.getFullYear();
}

async function loadTrialStatus() {
  const btn = document.getElementById('trialBtn');
  if (!btn) return;

  try {
    const user = await authAPI.self();
    if (user && user.paid_until) {
      const now = Math.floor(Date.now() / 1000);
      if (user.paid_until > now) {
        const section = document.querySelector('.trial-section');
        if (section) section.style.display = 'none';
        return;
      }
    }
  } catch (err) {
    console.log('Не удалось проверить paid_until, проверяем платежи');
  }

  try {
    const data = await paymentAPI.listPayments(1, 50);
    const hasActivePayment = data.list.some(p => p.status_int === 2);
    if (hasActivePayment) {
      const section = document.querySelector('.trial-section');
      if (section) section.style.display = 'none';
      return;
    }
  } catch (err) {
    console.log('Не удалось проверить историю платежей, показываем кнопку');
  }

  btn.addEventListener('click', onTrialClick);
}

async function onTrialClick(e) {
  const btn = e.currentTarget;
  if (btn.classList.contains('trial-used') || btn.classList.contains('loading')) return;

  btn.classList.add('loading');
  btn.textContent = 'Активируем...';

  try {
    await paymentAPI.takeTrial();
    btn.classList.add('trial-used');
    btn.classList.remove('loading');
    btn.textContent = 'Пробный период активирован';
    setTimeout(() => location.reload(), 1500);
  } catch (err) {
    console.error('Ошибка активации пробного периода', err);
    btn.classList.remove('loading');
    btn.textContent = 'Активировать пробный период';
    if (err.status === 400 || err.status === 409) {
      btn.classList.add('trial-used');
      btn.textContent = 'Пробный период активирован';
    }
  }
}

async function loadRates() {
  const grid = document.querySelector('.plans-grid');
  if (!grid) return;

  try {
    const data = await paymentAPI.listRates();
    renderRates(grid, data.list);
  } catch (err) {
    console.error('Не удалось загрузить тарифы', err);
    showRatesError(grid);
  }
}

function clearSkeletons(grid) {
  grid.querySelectorAll('.plan-card-skeleton').forEach(el => el.remove());
}

function renderRates(grid, rates) {
  clearSkeletons(grid);

  if (!rates.length) {
    grid.insertAdjacentHTML('beforeend',
      '<p class="plans-empty">Нет доступных тарифов</p>');
    return;
  }

  rates.forEach(rate => {
    const card = buildRateCard(rate);
    grid.appendChild(card);
  });
}

function showRatesError(grid) {
  clearSkeletons(grid);
  grid.insertAdjacentHTML('beforeend',
    '<p class="plans-error">Ошибка загрузки тарифов</p>');
}

function buildRateCard(rate) {
  const article = document.createElement('article');
  article.className = 'plan-card';

  article.innerHTML =
    '<h3 class="plan-name">' + escapeHtml(rate.name) + '</h3>' +
    '<p class="plan-price">' +
    '<span class="plan-amount">' + (rate.amount_penny / 100) + '</span>' +
    '<span class="plan-currency">₽</span>' +
    '</p>' +
    '<p class="plan-period">' + rate.paid_days + ' дней</p>' +
    '<p class="plan-daily">' + dailyPrice(rate.amount_penny, rate.paid_days) + '</p>' +
    '<button type="button" class="plan-btn">Выбрать тариф</button>';

  const btn = article.querySelector('.plan-btn');
  btn.addEventListener('click', () => onSelectRate(rate, btn));

  return article;
}

async function onSelectRate(rate, btn) {
  btn.classList.add('loading');
  btn.textContent = 'Создаём платёж…';

  try {
    const result = await paymentAPI.create(rate.id);
    if (result.url) {
      window.location.href = result.url;
    } else {
      btn.textContent = 'Готово';
      btn.classList.remove('loading');
    }
  } catch (err) {
    console.error('Ошибка создания платежа', err);
    btn.textContent = 'Ошибка';
    btn.classList.remove('loading');
  }
}

async function loadHistory() {
  const list = document.getElementById('historyList');
  if (!list) return;

  try {
    const data = await paymentAPI.listPayments(1, 50);
    renderHistory(list, data.list);
  } catch (err) {
    console.error('Не удалось загрузить историю', err);
    showHistoryError(list);
  }
}

function renderHistory(list, payments) {
  list.innerHTML = '';

  if (!payments.length) {
    list.insertAdjacentHTML('beforeend',
      '<p class="history-empty">Платежей пока нет</p>');
    return;
  }

  payments.forEach(p => {
    const row = buildHistoryRow(p);
    list.appendChild(row);
  });
}

function showHistoryError(list) {
  list.innerHTML = '';
  list.insertAdjacentHTML('beforeend',
    '<p class="history-error">Ошибка загрузки истории</p>');
}

function buildHistoryRow(p) {
  const div = document.createElement('div');
  div.className = 'history-row';

  const badgeClass = badgeClassByStatus(p.status_int);

  div.innerHTML =
    '<span class="history-icon" aria-hidden="true">' +
    '<svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">' +
    '<path d="M6 3H18V21L15 19L12 21L9 19L6 21V3Z" stroke-width="2" stroke-linejoin="round"/>' +
    '<path d="M9 8H15M9 12H15" stroke-width="2" stroke-linecap="round"/>' +
    '</svg>' +
    '</span>' +
    '<div class="history-main">' +
    '<p class="history-title">Тариф</p>' +
    '<p class="history-date">' + fmtDate(p.created_at) + '</p>' +
    '</div>' +
    '<span class="history-side">' +
    '<span class="history-amount">' + (p.amount_penny / 100) + '</span>' +
    '<span class="history-badge ' + badgeClass + '">' + statusLabel(p.status_int) + '</span>' +
    '</span>';

  div.style.cursor = 'pointer';
  div.title = 'Перейти к платежу';
  div.setAttribute('role', 'button');
  div.setAttribute('tabindex', '0');
  div.addEventListener('click', () => goToPayment(p));

  return div;
}

function goToPayment(payment) {
  const orderId = payment.payment_id || payment.id;
  const url = 'https://yoomoney.ru/checkout/payments/v2/contract?orderId=' + encodeURIComponent(orderId);
  window.location.href = url;
}

function escapeHtml(str) {
  const div = document.createElement('div');
  div.appendChild(document.createTextNode(str));
  return div.innerHTML;
}
