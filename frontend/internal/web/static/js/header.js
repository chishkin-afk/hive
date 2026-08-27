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

// ---------- Paid until ----------
async function loadPaidUntil() {
    const paidUntilEl = document.getElementById('paidUntil');
    const menuPaidUntilEl = document.getElementById('menuPaidUntil');
    if (!paidUntilEl && !menuPaidUntilEl) return;

    try {
        const user = await authAPI.self();
        if (user && user.paid_until) {
            const now = Math.floor(Date.now() / 1000);
            if (user.paid_until > now) {
                const date = new Date(user.paid_until * 1000);
                const formatted = formatDate(date);
                if (paidUntilEl) paidUntilEl.textContent = 'до ' + formatted;
                if (menuPaidUntilEl) menuPaidUntilEl.textContent = 'до ' + formatted;
            } else {
                // Подписка истекла
                if (paidUntilEl) paidUntilEl.textContent = 'Тарифы';
                if (menuPaidUntilEl) menuPaidUntilEl.textContent = 'Тарифы';
            }
        } else {
            // Если нет paid_until — показываем «Тарифы»
            if (paidUntilEl) paidUntilEl.textContent = 'Тарифы';
            if (menuPaidUntilEl) menuPaidUntilEl.textContent = 'Тарифы';
        }
    } catch (err) {
        console.error('Не удалось загрузить информацию о подписке:', err);
        // Если пользователь не авторизован — показываем «Тарифы»
        if (paidUntilEl) paidUntilEl.textContent = 'Тарифы';
        if (menuPaidUntilEl) menuPaidUntilEl.textContent = 'Тарифы';
    }
}

function formatDate(date) {
    const day = String(date.getDate()).padStart(2, '0');
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const year = date.getFullYear();
    return day + '.' + month + '.' + year;
}

// Загружаем информацию о подписке после загрузки страницы
document.addEventListener('DOMContentLoaded', loadPaidUntil);
