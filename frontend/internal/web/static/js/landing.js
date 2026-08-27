const burgerBtn = document.getElementById('burgerBtn');
const mobileMenuOverflow = document.getElementById('mobileMenuOverflow');

function toggleMobileMenu() {
    burgerBtn.classList.toggle('active');
    mobileMenuOverflow.classList.toggle('open');
    document.body.classList.toggle('block-scroll');
}

burgerBtn.addEventListener('click', toggleMobileMenu);

const menuLinks = document.querySelectorAll('#mobileMenuOverflow .menu-link');
menuLinks.forEach(link => {
    link.addEventListener('click', toggleMobileMenu);
});

// Проверка авторизации (проверяем наличие токена в куках)
function checkAuth() {
    // Проверяем, есть ли токен в куках
    const cookies = document.cookie.split(';');
    for (let cookie of cookies) {
        cookie = cookie.trim();
        if (cookie.startsWith('access_token=') || cookie.startsWith('refresh_token=')) {
            return true;
        }
    }
    return false;
}

// Загрузка тарифов для лендинга
async function loadLandingRates() {
    const grid = document.getElementById('pricingGrid');
    if (!grid) return;

    if (typeof paymentAPI === 'undefined') {
        await loadRatesDirect(grid);
        return;
    }

    try {
        const data = await paymentAPI.listRates();
        renderLandingRates(grid, data.list);
    } catch (err) {
        console.error('Не удалось загрузить тарифы через paymentAPI:', err);
        await loadRatesDirect(grid);
    }
}

// Прямая загрузка тарифов через fetch (на случай, если paymentAPI не инициализирован)
async function loadRatesDirect(grid) {
    try {
        const apiUrl = window.APP_CONFIG?.apiUrl || '';
        const response = await fetch(apiUrl + '/payments/rates', {
            credentials: 'include',
            headers: { 'Content-Type': 'application/json' }
        });
        
        if (!response.ok) {
            throw new Error('Ошибка загрузки тарифов: ' + response.status);
        }
        
        const data = await response.json();
        if (data && data.list) {
            renderLandingRates(grid, data.list);
        } else {
            throw new Error('Некорректный ответ от сервера');
        }
    } catch (err) {
        console.error('Не удалось загрузить тарифы:', err);
        showRatesError(grid);
    }
}

function showRatesError(grid) {
    const skeletons = grid.querySelectorAll('.pricing-card.loading');
    if (skeletons.length > 0) {
        skeletons.forEach(s => s.remove());
        grid.insertAdjacentHTML('beforeend', '<p class="pricing-error">Не удалось загрузить тарифы</p>');
    }
}

function renderLandingRates(grid, rates) {
    // Удаляем скелетоны
    grid.querySelectorAll('.pricing-card.loading').forEach(el => el.remove());

    if (!rates || rates.length === 0) {
        grid.insertAdjacentHTML('beforeend', '<p class="pricing-empty">Нет доступных тарифов</p>');
        return;
    }

    rates.forEach(rate => {
        const card = buildLandingRateCard(rate);
        grid.appendChild(card);
    });

    // Принудительно добавляем класс visible для новых карточек
    grid.querySelectorAll('.pricing-card:not(.loading)').forEach(el => {
        el.classList.add('visible');
    });
}

function buildLandingRateCard(rate) {
    const card = document.createElement('div');
    card.className = 'pricing-card';

    const daily = rate.amount_penny / 100 / rate.paid_days;
    const dailyFormatted = '~' + Math.max(1, Math.round(daily)) + ' ₽ в день';

    card.innerHTML =
        '<h3 class="pricing-name">' + escapeHtml(rate.name) + '</h3>' +
        '<div class="pricing-price">' +
        '<span class="pricing-amount">' + (rate.amount_penny / 100) + '</span>' +
        '<span class="pricing-currency">₽</span>' +
        '</div>' +
        '<p class="pricing-period">' + rate.paid_days + ' дней</p>' +
        '<p class="pricing-daily">' + dailyFormatted + '</p>' +
        '<button class="pricing-btn">Выбрать тариф</button>';

    const btn = card.querySelector('.pricing-btn');
    btn.addEventListener('click', async () => {
        // Проверяем, авторизован ли пользователь
        if (!checkAuth()) {
            window.location.href = '/auth';
            return;
        }

        btn.classList.add('loading');
        btn.textContent = 'Создаём платёж…';
        try {
            const result = await paymentAPI.create(rate.id);
            if (result.url) {
                window.location.href = result.url;
            } else {
                btn.textContent = 'Ошибка';
                btn.classList.remove('loading');
            }
        } catch (err) {
            console.error('Ошибка создания платежа', err);
            if (err.status === 401) {
                window.location.href = '/auth?action=login';
                return;
            }
            btn.textContent = 'Ошибка';
            btn.classList.remove('loading');
        }
    });

    return card;
}

// Простая функция для экранирования HTML
function escapeHtml(str) {
    const div = document.createElement('div');
    div.appendChild(document.createTextNode(str));
    return div.innerHTML;
}

document.addEventListener('DOMContentLoaded', () => {
    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.classList.add('visible');
                observer.unobserve(entry.target);
            }
        });
    }, { threshold: 0.1 });

    document.querySelectorAll(`
        .section-title, .section-subtitle, .feature-card,
        .pricing-card, .faq-item
    `)
        .forEach(el => observer.observe(el));

    // Загружаем тарифы после того, как страница полностью загружена
    loadLandingRates();
});

const faqItems = document.querySelectorAll('.faq-item');
faqItems.forEach(item => {
    item.addEventListener('click', () => {
        const isOpen = item.classList.contains('open');

        faqItems.forEach(item => {
            item.classList.remove('open');
            item.querySelector('.faq-answer').style.maxHeight = '0px';
        })

        if (!isOpen) {
            item.classList.add('open');
            const answer = item.querySelector('.faq-answer');
            const height = answer.querySelector('p').scrollHeight;
            answer.style.maxHeight = height + 'px';
        }
    });
})