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

const pricingGrid = document.getElementById('pricingGrid');

async function loadPlans() {
    try {
        if (typeof paymentAPI === 'undefined' || typeof paymentAPI.listRates !== 'function') {
            console.error('paymentAPI is not available or listRates is not a function');
            return;
        }

        const data = await paymentAPI.listRates();
        renderRates(data.list);
    } catch (err) {
        console.error('Не удалось загрузить тарифы', err);
        pricingGrid.innerHTML = '<p class="plans-error">Не удалось загрузить тарифы. Попробуйте позже.</p>';
    }
}

function renderRates(rates) {
    const grid = document.getElementById('pricingGrid');
    if (!grid) return;

    // clear skeletons
    grid.querySelectorAll('.pricing-skeleton').forEach(el => el.remove());

    if (!rates || !rates.length) {
        grid.insertAdjacentHTML('beforeend', '<p class="plans-empty">Нет доступных тарифов</p>');
        return;
    }

    rates.forEach(rate => {
        const card = buildRateCard(rate);
        grid.appendChild(card);
    });
}

function buildRateCard(rate) {
    const div = document.createElement('div');
    div.className = 'pricing-card';

    const price = rate.amount_penny / 100;
    const priceFormatted = new Intl.NumberFormat('ru-RU').format(price);

    div.innerHTML = `
        <h3 class="pricing-name">${escapeHtml(rate.name)}</h3>
        <div class="pricing-price">
            <span class="pricing-amount">${priceFormatted}</span>
            <span class="pricing-currency">₽</span>
        </div>
        <p class="pricing-period">${rate.paid_days} дней</p>
        <a href="/auth?action=register&rate=${rate.id}" class="pricing-btn">Выбрать тариф</a>
    `;

    return div;
}

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

    loadPlans();
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