const CONFIG = window.__CONFIG__

async function apiRequest(url, data) {
    const response = await fetch(CONFIG.apiBase + url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
    });

    const result = await response.json();

    if (!response.ok) {
        throw new Error(result.error || 'Ошибка запроса');
    }

    return result;
}

function setAuthCookie(token, rememberMe) {
    const maxAge = rememberMe ? 30 * 24 * 60 * 60 : 0;
    document.cookie = `token=${token}; path=/; max-age=${maxAge}`;
}

const passwordToggles = document.querySelectorAll('.password-toggle');
passwordToggles.forEach(el => {
    el.addEventListener('click', function() {
        const wrapper = this.closest('.password-wrapper');
        const input = wrapper.querySelector('.form-input');

        const isPassword = input.type === 'password';
        input.type = isPassword ? 'text' : 'password';

        if (isPassword) {
            this.innerHTML = `
            <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M16.86 17.1975C15.4614 18.3376 13.7584 18.9693 12 19C6.27273 19 3 12 3 12C4.01773 9.97167 5.28929 8.39703 7 7M10.2818 5.21002C10.845 5.06904 11.4216 4.99857 12 5.00002C17.7273 5.00002 21 12 21 12C20.5033 12.9937 19.911 13.9292 19.2327 14.7913M13.7345 13.855C13.5098 14.1129 13.2389 14.3198 12.9378 14.4632C12.6367 14.6067 12.3116 14.6839 11.9821 14.6901C11.6525 14.6963 11.3251 14.6315 11.0195 14.4994C10.7139 14.3674 10.4362 14.1709 10.2031 13.9216C9.97007 13.6724 9.78632 13.3755 9.66287 13.0486C9.53942 12.7217 9.47879 12.3716 9.48461 12.0192C9.49042 11.6667 9.56256 11.3191 9.69672 10.9971C9.83087 10.6751 9.96199 10.5626 10.2031 10.3222"
                    stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                <path d="M3 3L21 21"
                    stroke-width="2" stroke-linecap="round"/>
            </svg>
            `;
        } else {
            this.innerHTML = `
            <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M3 12C3 12 6.27273 5 12 5C17.7273 5 21 12 21 12C21 12 17.7273 19 12 19C6.27273 19 3 12 3 12Z"
                    stroke-width="2" stroke-linejoin="round"/>
                <path d="M12 15C13.6569 15 15 13.6569 15 12C15 10.3431 13.6569 9 12 9C10.3431 9 9 10.3431 9 12C9 13.6569 10.3431 15 12 15Z"
                    stroke-width="2"/>
            </svg>
            `;
        }
    });
});

document.querySelectorAll('.switch-btn').forEach(el => {
    el.addEventListener('click', () => {
        clearMessages();
        document.getElementById('loginForm').classList.toggle('active');
        document.getElementById('registerForm').classList.toggle('active');
    });
});

const reEmail = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const rePhone = /^(7|8)\d{10}$/;

function isValidEmail(email) {
    return reEmail.test(email);
}

function isValidPhone(phone) {
    return rePhone.test(phone.replace(/[\s\-\(\)]/g, ''));
}

function validateLogin(data) {
    if (!data.email) return 'Введите email';
    if (!isValidEmail(data.email)) return 'Введите корректный email';
    if (!data.password) return 'Введите пароль';
    return null;
}

function validateRegister(data) {
    if (!data.firstname) return 'Введите имя';
    if (!data.email) return 'Введите email';
    if (!isValidEmail(data.email)) return 'Введите корректный email';
    if (data.phone && !isValidPhone(data.phone)) return 'Введите корректный телефон';
    if (!data.password) return 'Введите пароль';
    if (data.password.length < 6) return 'Пароль должен быть не менее 6 символов';
    if (data.password.length > 32) return 'Пароль должен быть не более 32 символов';
    if (!data.agreeTerms) return 'Необходимо принять пользовательское соглашение';
    return null;
}

function showError(msg) {
    const el = document.getElementById('errorMessage');
    el.textContent = msg;
}

function showSuccess(msg) {
    const el = document.getElementById('successMessage');
    el.textContent = msg;
}

function clearMessages() {
    document.getElementById('errorMessage').textContent = '';
    document.getElementById('successMessage').textContent = '';
}

async function login(e) {
    const data = {
        email: document.getElementById('loginEmail').value.trim(),
        password: document.getElementById('loginPassword').value,
        rememberMe: document.getElementById('rememberMe').checked
    };

    clearMessages();

    const error = validateLogin(data);
    if (error) {
        showError(error);
        return;
    }

    try {
        const result = await apiRequest('/login', data);

        if (result.token) {
            setAuthCookie(result.token, data.rememberMe);
        }

        showSuccessMessage('Успешный вход');
        setTimeout(() => window.location.href = CONFIG.urls.dashboard, 1500);
    } catch (err) {
        console.error(err);
        showError('Ошибка. Попробуйте позже');
    }
}

async function register(e) {
    const data = {
        firstname: document.getElementById('regFirstname').value.trim(),
        lastname: document.getElementById('regLastname').value.trim(),
        email: document.getElementById('regEmail').value.trim(),
        phone: document.getElementById('regPhone').value.trim(),
        password: document.getElementById('regPassword').value,
        agreeTerms: document.getElementById('agreeTerms').checked
    };

    clearMessages();

    const error = validateRegister(data);
    if (error) {
        showError(error);
        return;
    }

    try {
        const result = await apiRequest('/register', data);

        if (result.token) {
            setAuthCookie(result.token, true);
        }

        showSuccessMessage('Успешная регистрация');
        setTimeout(() => window.location.href = CONFIG.urls.dashboard, 1500);
    } catch (err) {
        console.error(err);
        showError('Ошибка. Попробуйте позже');
    }
}

document.getElementById('loginBtn').addEventListener('click', login);
document.getElementById('registerBtn').addEventListener('click', register);
