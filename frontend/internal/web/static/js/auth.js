// ===== КОНФИГ =====
const REDIRECT_AFTER_LOGIN = '/services';
const LS_EMAIL = 'auth_confirm_email';
const LS_EXPIRES = 'auth_resend_expires_at';

// ===== УТИЛИТЫ =====

/** Конвертирует next_send_at (unix, секунды) → миллисекунды, минимум +3с, фолбек +60с. */
function normalizeNextSendAt(resp) {
    let ms = resp?.next_send_at ? resp.next_send_at * 1000 : Date.now() + 60_000;
    const floor = Date.now() + 3_000;
    return ms < floor ? floor : ms;
}

// ===== DOM-ХЕЛПЕРЫ =====

const $ = (id) => document.getElementById(id);

function clearFieldErrors(formSelector) {
    document.querySelectorAll(`${formSelector} .form-input, #confirmCode`).forEach(el => el.classList.remove('error'));
    const checkmark = document.querySelector(`${formSelector} .checkmark`);
    if (checkmark) checkmark.classList.remove('error');
}

function showForm(formId) {
    ['loginForm', 'registerForm', 'confirmForm', 'forgotForm'].forEach(id => $(id).classList.toggle('active', id === formId));
}

function unlockResendBtn() {
    const btn = $('resendBtn');
    btn.disabled = false;
    btn.classList.remove('disabled');
    btn.textContent = 'Отправить код снова';
}

function lockResendBtn() {
    const btn = $('resendBtn');
    btn.disabled = true;
    btn.classList.add('disabled');
}

function clearConfirmState() {
    localStorage.removeItem(LS_EMAIL);
    localStorage.removeItem(LS_EXPIRES);
    if (window._resendTimer) {
        clearInterval(window._resendTimer);
        window._resendTimer = null;
    }
}

// ===== УВЕДОМЛЕНИЯ =====

function showNotification(message, type = 'info') {
    const existing = document.querySelector('.notification-toast');
    if (existing) existing.remove();

    const toast = document.createElement('div');
    toast.className = `notification-toast ${type}`;
    toast.textContent = message;
    document.body.appendChild(toast);

    requestAnimationFrame(() => toast.classList.add('show'));

    setTimeout(() => {
        toast.classList.remove('show');
        setTimeout(() => toast.remove(), 300);
    }, 4000);
}

// ===== ВАЛИДАЦИЯ НА ЛЕТУ =====

function setupLiveValidation() {
    const rules = [
        { id: 'regEmail',    test: (v) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v), emptyIsOk: false },
        { id: 'regPassword', test: (v) => v.length >= 6,                            emptyIsOk: true  },
        { id: 'regFirstname',test: (v) => v.length >= 3,                            emptyIsOk: false },
        { id: 'regLastname', test: (v) => v.length >= 3,                            emptyIsOk: true  },
        { id: 'regPhone',    test: (v) => /^(\+7|8)\d{10}$/.test(v.replace(/[\s\-\(\)]/g, '')), emptyIsOk: true },
    ];

    rules.forEach(({ id, test, emptyIsOk }) => {
        const el = $(id);
        if (!el) return;
        el.addEventListener('input', function () {
            const val = this.value.trim();
            if (!val && emptyIsOk) {
                this.classList.remove('error', 'success');
                return;
            }
            const ok = val && test(val);
            this.classList.toggle('error', !ok);
            this.classList.toggle('success', ok);
        });
    });

    const checkbox = $('agreeTerms');
    if (checkbox) {
        checkbox.addEventListener('change', function () {
            const mark = this.closest('.checkbox-label').querySelector('.checkmark');
            mark.classList.toggle('error', !this.checked);
            mark.classList.toggle('success', this.checked);
        });
    }
}

// ===== TOGGLE ПАРОЛЯ =====

document.querySelectorAll('.password-toggle').forEach(el => {
    el.addEventListener('click', function () {
        const input = this.closest('.password-wrapper').querySelector('.form-input');
        const hidden = input.type === 'password';
        input.type = hidden ? 'text' : 'password';
        this.innerHTML = hidden
            ? `<svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M16.86 17.1975C15.4614 18.3376 13.7584 18.9693 12 19C6.27273 19 3 12 3 12C4.01773 9.97167 5.28929 8.39703 7 7M10.2818 5.21002C10.845 5.06904 11.4216 4.99857 12 5.00002C17.7273 5.00002 21 12 21 12C20.5033 12.9937 19.911 13.9292 19.2327 14.7913M13.7345 13.855C13.5098 14.1129 13.2389 14.3198 12.9378 14.4632C12.6367 14.6067 12.3116 14.6839 11.9821 14.6901C11.6525 14.6963 11.3251 14.6315 11.0195 14.4994C10.7139 14.3674 10.4362 14.1709 10.2031 13.9216C9.97007 13.6724 9.78632 13.3755 9.66287 13.0486C9.53942 12.7217 9.47879 12.3716 9.48461 12.0192C9.49042 11.6667 9.56256 11.3191 9.69672 10.9971C9.83087 10.6751 9.96199 10.5626 10.2031 10.3222" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                <path d="M3 3L21 21" stroke-width="2" stroke-linecap="round"/>
            </svg>`
            : `<svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M3 12C3 12 6.27273 5 12 5C17.7273 5 21 12 21 12C21 12 17.7273 19 12 19C6.27273 19 3 12 3 12Z" stroke-width="2" stroke-linejoin="round"/>
                <path d="M12 15C13.6569 15 15 13.6569 15 12C15 10.3431 13.6569 9 12 9C10.3431 9 9 10.3431 9 12C9 13.6569 10.3431 15 12 15Z" stroke-width="2"/>
            </svg>`;
    });
});

// ===== ПЕРЕКЛЮЧЕНИЕ ФОРМ =====

document.querySelectorAll('.switch-btn').forEach(btn => {
    btn.addEventListener('click', () => {
        $('loginForm').classList.toggle('active');
        $('registerForm').classList.toggle('active');
    });
});

// ===== АДАПТАЦИЯ ОШИБОК =====

const ERROR_MAP = [
    { field: null,        match: (s) => s === 409,                                                     text: 'Пользователь с таким email уже существует.' },
    { field: null,        match: (s) => s === 401,                                                     text: 'Неверный email или пароль.' },
    { field: null,        match: (s) => s === 403,                                                     text: 'Доступ запрещён. Возможно, аккаунт не подтверждён.' },
    { field: null,        match: (s) => s === 429,                                                     text: 'Слишком много попыток. Подождите немного.' },
    { field: null,        match: (s) => s >= 500,                                                      text: 'Ошибка на сервере. Попробуйте позже.' },
    { field: 'regEmail',  match: (s, m) => m.includes('email') || m.includes('e-mail'),                 text: 'Неверный формат email.' },
    { field: 'regPassword',match:(s, m) => m.includes('password') || m.includes('пароль'),              text: 'Неверный пароль.' },
    { field: 'regFirstname',match:(s,m)=> m.includes('firstname') || m.includes('first name') || m.includes('имя'), text: 'Некорректное имя.' },
    { field: 'regPhone',  match: (s, m) => m.includes('phone') || m.includes('телефон'),                text: 'Неверный формат телефона.' },
    { field: 'confirmCode',match:(s,m)=> m.includes('code') || m.includes('код') || m.includes('confirm'), text: 'Неверный код подтверждения.' },
    { field: 'regEmail',  match: (s, m) => m.includes('exist') || m.includes('существует') || m.includes('занят'), text: 'Пользователь с таким email уже существует.' },
];

// 404 различается по контексту
const CONTEXT_404 = {
    login:   { field: 'loginEmail', text: 'Пользователь с таким email не найден.' },
    confirm: { field: 'confirmCode', text: 'Неверный код подтверждения.' },
    forgot:  { field: 'forgotEmail', text: 'Пользователь с таким email не найден.' },
};

function formatError(err, context = '') {
    const status = err.status || 0;
    const msg = (err.message || '').toLowerCase();

    // 404 обрабатываем отдельно — зависит от контекста
    if (status === 404 && CONTEXT_404[context]) {
        return CONTEXT_404[context];
    }

    for (const rule of ERROR_MAP) {
        if (rule.match(status, msg)) return rule;
    }
    return { field: 'regEmail', text: err.message || 'Произошла неизвестная ошибка.' };
}

// ===== ВАЛИДАЦИЯ ПОЛЕЙ =====

const RE_EMAIL = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

function isValidPhone(phone) {
    return /^(\+7|8)\d{10}$/.test(phone.replace(/[\s\-\(\)]/g, ''));
}

// ===== ЛОГИН =====

async function login() {
    const emailInput = $('loginEmail');
    const passInput = $('loginPassword');
    const email = emailInput.value.trim();
    const password = passInput.value;

    emailInput.classList.remove('error');
    passInput.classList.remove('error');

    if (!email)              { emailInput.classList.add('error'); showNotification('Введите email', 'error'); return; }
    if (!RE_EMAIL.test(email)) { emailInput.classList.add('error'); showNotification('Введите корректный email', 'error'); return; }
    if (!password)           { passInput.classList.add('error'); showNotification('Введите пароль', 'error'); return; }

    try {
        await authAPI.login(email, password);
        showNotification('Успешный вход!', 'success');
        setTimeout(() => { window.location.href = REDIRECT_AFTER_LOGIN; }, 1500);
    } catch (err) {
        const e = formatError(err, 'login');
        if (e.field === 'loginEmail') emailInput.classList.add('error');
        if (e.field === 'loginPassword') passInput.classList.add('error');
        showNotification(e.text, 'error');
    }
}

// ===== РЕГИСТРАЦИЯ =====

async function register() {
    const inputs = {
        firstname: $('regFirstname'),
        lastname:  $('regLastname'),
        email:     $('regEmail'),
        phone:     $('regPhone'),
        password:  $('regPassword'),
    };
    const agreeCheckbox = $('agreeTerms');
    const checkmark = agreeCheckbox.closest('.checkbox-label').querySelector('.checkmark');

    const f = (el) => el.value.trim();
    const firstname = f(inputs.firstname);
    const lastname  = f(inputs.lastname) || null;
    const email     = f(inputs.email);
    const phone     = f(inputs.phone) || null;
    const password  = inputs.password.value;
    const agree     = agreeCheckbox.checked;

    clearFieldErrors('#registerForm');

    // Валидация
    const fail = (el, msg) => { el.classList.add('error'); showNotification(msg, 'error'); };

    if (!firstname)              { fail(inputs.firstname, 'Введите имя'); return; }
    if (firstname.length < 3)    { fail(inputs.firstname, 'Имя должно быть не менее 3 символов'); return; }
    if (lastname && lastname.length < 3) { fail(inputs.lastname, 'Фамилия должна быть не менее 3 символов'); return; }
    if (!email)                  { fail(inputs.email, 'Введите email'); return; }
    if (!RE_EMAIL.test(email))   { fail(inputs.email, 'Введите корректный email'); return; }
    if (phone && !isValidPhone(phone)) { fail(inputs.phone, 'Введите корректный телефон (+7 или 8)'); return; }
    if (!password)               { fail(inputs.password, 'Введите пароль'); return; }
    if (password.length < 6)     { fail(inputs.password, 'Пароль должен быть не менее 6 символов'); return; }
    if (password.length > 32)    { fail(inputs.password, 'Пароль должен быть не более 32 символов'); return; }
    if (!agree)                  { checkmark.classList.add('error'); showNotification('Примите пользовательское соглашение', 'error'); return; }

    try {
        const result = await authAPI.prepare(firstname, lastname, email, phone, password);
        localStorage.setItem(LS_EMAIL, email);
        localStorage.setItem(LS_EXPIRES, String(normalizeNextSendAt(result)));
        showNotification('Код подтверждения отправлен на email!', 'success');
        showForm('confirmForm');
        $('confirmCode').value = '';
        startResendTimer();
    } catch (err) {
        const e = formatError(err);
        const fieldEl = e.field ? $(e.field) : null;
        if (fieldEl) fieldEl.classList.add('error');
        showNotification(e.text, 'error');
    }
}

// ===== ПОДТВЕРЖДЕНИЕ КОДА =====

async function confirmEmail() {
    const email = localStorage.getItem(LS_EMAIL);
    const codeInput = $('confirmCode');
    const code = codeInput.value.trim();

    codeInput.classList.remove('error');

    if (!email) { showNotification('Сессия регистрации истекла. Попробуйте зарегистрироваться снова', 'error'); return; }
    if (!code)  { codeInput.classList.add('error'); showNotification('Введите код из письма', 'error'); return; }

    try {
        await authAPI.confirm(email, parseInt(code, 10));
        clearConfirmState();
        showNotification('Регистрация завершена!', 'success');
        setTimeout(() => { window.location.href = REDIRECT_AFTER_LOGIN; }, 1500);
    } catch (err) {
        const e = formatError(err, 'confirm');
        codeInput.classList.add('error');
        showNotification(e.text, 'error');
    }
}

// ===== ПОВТОРНАЯ ОТПРАВКА КОДА =====

async function resendCode() {
    const email = localStorage.getItem(LS_EMAIL);

    if (!email) {
        showNotification('Сессия регистрации истекла. Зарегистрируйтесь заново.', 'error');
        clearConfirmState();
        showForm('registerForm');
        return;
    }

    const storedExpires = localStorage.getItem(LS_EXPIRES);
    if (storedExpires && Date.now() < parseInt(storedExpires, 10)) {
        showNotification('Подождите, пока истечёт таймер.', 'error');
        return;
    }

    lockResendBtn();

    try {
        const result = await authAPI.resend_code(email);
        localStorage.setItem(LS_EXPIRES, String(normalizeNextSendAt(result)));
        showNotification('Код отправлен повторно!', 'success');
        showForm('confirmForm');
        $('confirmCode').value = '';
        startResendTimer();
    } catch (err) {
        const e = formatError(err);
        unlockResendBtn();

        if (e.text.includes('уже существует')) {
            showNotification('Пользователь уже зарегистрирован. Проверьте почту или войдите.', 'error');
            showForm('confirmForm');
            return;
        }
        showNotification(e.text, 'error');
    }
}

// ===== ВЫХОД ИЗ ПОДТВЕРЖДЕНИЯ =====

function cancelConfirmation() {
    clearConfirmState();
    showForm('registerForm');
    $('confirmCode').value = '';
    unlockResendBtn();
    showNotification('Вы вернулись к регистрации', 'info');
}

// ===== ТАЙМЕР =====

function startResendTimer() {
    const btn = $('resendBtn');
    const expiresAt = localStorage.getItem(LS_EXPIRES);

    if (!expiresAt) { unlockResendBtn(); return; }

    const expiresNum = parseInt(expiresAt, 10);
    const remaining = Math.max(0, expiresNum - Date.now());

    if (remaining <= 0) { unlockResendBtn(); return; }

    lockResendBtn();
    if (window._resendTimer) clearInterval(window._resendTimer);

    const tick = () => {
        const left = Math.max(0, expiresNum - Date.now());
        if (left <= 0) {
            clearInterval(window._resendTimer);
            window._resendTimer = null;
            unlockResendBtn();
            localStorage.removeItem(LS_EXPIRES);
            return;
        }
        btn.textContent = `Отправить снова (${Math.ceil(left / 1000)}с)`;
    };

    tick();
    window._resendTimer = setInterval(tick, 1000);
}

// ===== ВОССТАНОВЛЕНИЕ ПАРОЛЯ: ШАГ 1 — ОТПРАВКА EMAIL =====

async function forgotPassword() {
    const emailInput = $('forgotEmail');
    const email = emailInput.value.trim();

    emailInput.classList.remove('error');

    if (!email)              { emailInput.classList.add('error'); showNotification('Введите email', 'error'); return; }
    if (!RE_EMAIL.test(email)) { emailInput.classList.add('error'); showNotification('Введите корректный email', 'error'); return; }

    try {
        await authAPI.forget_password(email);
        showNotification('Инструкция отправлена на email! Проверьте почту.', 'success');
    } catch (err) {
        const e = formatError(err, 'forgot');
        emailInput.classList.add('error');
        showNotification(e.text, 'error');
    }
}

// ===== ВОССТАНОВЛЕНИЕ =====

function restoreAuthState() {
    if (localStorage.getItem(LS_EMAIL)) {
        showForm('confirmForm');
        startResendTimer();
    }
}

// ===== СОБЫТИЯ =====

document.addEventListener('DOMContentLoaded', () => {
    // На странице /reset/ свой скрипт, auth.js не нужен
    if (!$('loginForm') && !$('registerForm')) return;

    setupLiveValidation();
    restoreAuthState();

    $('loginBtn').addEventListener('click', login);
    $('registerBtn').addEventListener('click', register);
    $('confirmBtn').addEventListener('click', confirmEmail);
    $('resendBtn').addEventListener('click', resendCode);

    const cancelBtn = $('cancelConfirmBtn');
    if (cancelBtn) cancelBtn.addEventListener('click', cancelConfirmation);

    // Восстановление пароля
    const forgotLink = $('forgotPasswordLink');
    if (forgotLink) forgotLink.addEventListener('click', (e) => { e.preventDefault(); showForm('forgotForm'); });

    $('forgotBtn').addEventListener('click', forgotPassword);
    $('cancelForgotBtn').addEventListener('click', () => showForm('loginForm'));

    $('loginForm').addEventListener('keydown', (e) => { if (e.key === 'Enter') { e.preventDefault(); login(); } });
    $('registerForm').addEventListener('keydown', (e) => { if (e.key === 'Enter') { e.preventDefault(); register(); } });
    $('confirmForm').addEventListener('keydown', (e) => { if (e.key === 'Enter') { e.preventDefault(); confirmEmail(); } });
    $('forgotForm').addEventListener('keydown', (e) => { if (e.key === 'Enter') { e.preventDefault(); forgotPassword(); } });
});
