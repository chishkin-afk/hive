(function () {
    'use strict';

    const form = document.getElementById('profileForm');
    const navItems = document.querySelectorAll('.profile-nav-item');
    const saveBtn = document.getElementById('profileSaveBtn');
    const heroName = document.querySelector('.profile-hero-name');
    const initialsEl = document.querySelector('.profile-avatar-initials');
    const skeleton = document.getElementById('profileSkeleton');
    const profileContent = document.getElementById('profileContent');
    const tgConnect = document.getElementById('tgConnect');
    const tgStatus = document.getElementById('tgStatus');
    const tgBindBtn = document.getElementById('tgBindBtn');
    const changePasswordBtn = document.getElementById('changePasswordBtn');
    const logoutBtn = document.getElementById('profileLogoutBtn');

    if (!form) return;

    let initialState = {};
    let tgBound = false;
    // --- Toast system (как в auth) ---
    function showToast(message, type) {
        const existing = document.querySelector('.profile-toast');
        if (existing) existing.remove();

        const toast = document.createElement('div');
        toast.className = 'profile-toast';
        if (type) toast.classList.add(type);
        toast.textContent = message;
        document.body.appendChild(toast);

        requestAnimationFrame(() => toast.classList.add('show'));

        setTimeout(() => {
            toast.classList.remove('show');
            setTimeout(() => toast.remove(), 400);
        }, 4000);
    }

    function showSkeleton(show) {
        if (!skeleton || !profileContent) return;
        skeleton.style.display = show ? 'flex' : 'none';
        profileContent.style.display = show ? 'none' : '';
    }

    function updateHero(firstname, lastname) {
        const full = [firstname, lastname].filter(Boolean).join(' ') || 'Пользователь';
        if (heroName) heroName.textContent = full;
        if (initialsEl) {
            const f = (firstname || '').charAt(0).toUpperCase();
            const l = (lastname || '').charAt(0).toUpperCase();
            initialsEl.textContent = (f + l) || '?';
        }
    }

    // --- Capture / detect changes ---
    function captureFormState() {
        const snapshot = {};
        ['firstname', 'lastname', 'phone'].forEach((name) => {
            const el = form.elements[name];
            if (el) snapshot[name] = el.value;
        });
        ['notify_new_booking', 'on_vacation'].forEach((name) => {
            const el = form.elements[name];
            if (el) snapshot[name] = el.checked;
        });
        return snapshot;
    }

    function isDirty() {
        const current = captureFormState();
        for (const key of Object.keys(initialState)) {
            if (current[key] !== initialState[key]) return true;
        }
        for (const key of Object.keys(current)) {
            if (!(key in initialState) || current[key] !== initialState[key]) return true;
        }
        return false;
    }

    function toggleSaveBtn() {
        if (!saveBtn) return;
        if (isDirty()) {
            saveBtn.style.display = '';
        } else {
            saveBtn.style.display = 'none';
        }
    }

    // --- Populate form from API response ---
    function populateForm(user) {
        if (user.firstname != null) {
            const el = form.elements['firstname'];
            if (el) el.value = user.firstname;
        }
        if (user.lastname != null) {
            const el = form.elements['lastname'];
            if (el) el.value = user.lastname;
        }
        const emailEls = form.querySelectorAll('input[name="email"]');
        emailEls.forEach((el) => {
            if (user.email != null) el.value = user.email;
        });
        if (user.phone_number != null) {
            const el = form.elements['phone'];
            if (el) el.value = formatPhone(user.phone_number);
        }
        if (user.notification != null && user.notification.telegram_enable != null) {
            const cb = form.elements['notify_new_booking'];
            if (cb) cb.checked = user.notification.telegram_enable;
        }
        if (user.on_vacation != null) {
            const cb = form.elements['on_vacation'];
            if (cb) cb.checked = user.on_vacation;
        }
        tgBound = !!(user.telegram_id || user.notification?.telegram_chat_id);
        updateTgUI();

        initialState = captureFormState();
        toggleSaveBtn();
        updateHero(user.firstname || '', user.lastname || '');
    }

    // --- Telegram UI ---
    function updateTgUI() {
        if (!tgConnect) return;
        if (tgBound) {
            if (tgStatus) {
                tgStatus.textContent = 'Подключено';
                tgStatus.className = 'profile-connect-status active';
            }
            if (tgBindBtn) {
                tgBindBtn.textContent = 'Отвязать';
                tgBindBtn.className = 'profile-btn profile-btn-ghost';
            }
        } else {
            if (tgStatus) {
                tgStatus.textContent = 'Не подключено';
                tgStatus.className = 'profile-connect-status';
            }
            if (tgBindBtn) {
                tgBindBtn.textContent = 'Привязать';
                tgBindBtn.className = 'profile-btn profile-btn-primary';
            }
        }
    }

    // --- Wrap height management ---
    function updateWrapHeight() {
        const wrap = document.querySelector('.profile-sections-wrap');
        const active = document.querySelector('.profile-section.is-active');
        if (!wrap || !active) return;
        active.style.position = 'relative';
        active.style.display = 'flex';
        wrap.style.minHeight = active.offsetHeight + 'px';
        active.style.position = '';
        active.style.display = '';
    }

    function switchSection(sectionName) {
        const currentActive = document.querySelector('.profile-section.is-active');
        const nextActive = document.querySelector(`.profile-section[data-section="${sectionName}"]`);
        if (!nextActive || currentActive === nextActive) return;

        navItems.forEach((btn) => {
            btn.classList.toggle('active', btn.dataset.section === sectionName);
        });

        const wrap = document.querySelector('.profile-sections-wrap');

        if (currentActive && wrap) {
            wrap.style.minHeight = currentActive.offsetHeight + 'px';
        }

        if (currentActive) {
            currentActive.classList.remove('is-active');
            currentActive.classList.add('is-leaving');
            currentActive.removeAttribute('hidden');
            setTimeout(() => {
                currentActive.classList.remove('is-leaving');
                currentActive.setAttribute('hidden', '');
            }, 250);
        }

        nextActive.removeAttribute('hidden');
        nextActive.style.position = 'relative';
        nextActive.style.display = 'flex';
        const newHeight = nextActive.offsetHeight;
        nextActive.style.position = '';
        nextActive.style.display = '';

        void nextActive.offsetHeight;
        nextActive.classList.add('is-active');

        if (wrap) wrap.style.minHeight = newHeight + 'px';
    }

    // ===== ВАЛИДАЦИЯ НА ЛЕТУ (как в auth) =====
    const PHONE_RE = /^[\+\d\s\-]*$/;

    /** Убирает всё лишнее, оставляет только цифры */
    function phoneDigits(val) {
        return (val || '').replace(/[^\d]/g, '');
    }

    /** Форматирует цифры в +7 999 123-45-67 */
    function formatPhone(val) {
        let d = phoneDigits(val);
        if (d.length === 0) return '';
        // Нормализуем: 8 в начале -> 7, убираем +7/8 если уже есть
        if (d.startsWith('8')) d = '7' + d.slice(1);
        // Если больше 11 цифр — обрезаем
        if (d.length > 11) d = d.slice(0, 11);
        // Если начинается с 7 и это российский номер (11 цифр или строится)
        if (d.startsWith('7')) {
            const len = d.length;
            if (len <= 1) return '+7';
            if (len <= 4) return '+7 ' + d.slice(1);
            if (len <= 7) return '+7 ' + d.slice(1, 4) + ' ' + d.slice(4);
            if (len <= 9) return '+7 ' + d.slice(1, 4) + ' ' + d.slice(4, 7) + '-' + d.slice(7);
            return '+7 ' + d.slice(1, 4) + ' ' + d.slice(4, 7) + '-' + d.slice(7, 9) + '-' + d.slice(9);
        }
        return '+7 ' + d;
    }

    function isValidPhone(val) {
        const d = phoneDigits(val);
        return d.length === 11 && (d.startsWith('7') || d.startsWith('8'));
    }

    function setupLiveValidation() {
        const rules = [
            { name: 'firstname', test: (v) => v.length >= 3, emptyIsOk: false },
            { name: 'lastname',  test: (v) => v.length >= 3, emptyIsOk: true  },
            { name: 'phone',     test: (v) => isValidPhone(v), emptyIsOk: true },
        ];

        rules.forEach(({ name, test, emptyIsOk }) => {
            const el = form.elements[name];
            if (!el) return;
            el.addEventListener('input', function () {
                const val = this.value.trim();
                if (!val && emptyIsOk) {
                    this.classList.remove('input-error', 'input-success');
                    return;
                }
                const ok = val && test(val);
                this.classList.toggle('input-error', !ok);
                this.classList.toggle('input-success', ok);
            });
        });

        // Маска ввода для телефона — форматирует в +7 999 123-45-67
        const phoneEl = form.elements['phone'];
        if (phoneEl) {
            phoneEl.addEventListener('input', function () {
                const cursor = this.selectionStart;
                const oldLen = (this.value || '').length;
                const digits = phoneDigits(this.value);
                if (digits.length > 11) return; // не даём ввести больше 11 цифр
                this.value = formatPhone(digits);
                // Восстанавливаем позицию курсора
                const newLen = this.value.length;
                const shift = newLen - oldLen;
                this.setSelectionRange(cursor + shift, cursor + shift);
            });
            phoneEl.addEventListener('keydown', function (e) {
                // Разрешаем backspace/delete даже если уже 11 цифр
                if (e.key === 'Backspace' || e.key === 'Delete') return;
                if (phoneDigits(this.value).length >= 11) {
                    e.preventDefault();
                }
            });
        }
    }

    // Проверка всех полей перед отправкой
    function validateAllFields() {
        const firstname = form.elements['firstname'];
        const phone = form.elements['phone'];
        const notifyCb = form.elements['notify_new_booking'];
        let valid = true;

        if (firstname && firstname.value.trim().length < 3) {
            firstname.classList.add('input-error');
            showToast('Имя должно содержать минимум 3 символа', 'error');
            valid = false;
        }
        if (phone && phone.value.trim() && !isValidPhone(phone.value.trim())) {
            phone.classList.add('input-error');
            showToast('Введите корректный номер телефона', 'error');
            valid = false;
        }
        // Не даём включить уведомления, если Telegram не подключен
        if (notifyCb && notifyCb.checked && !tgBound) {
            notifyCb.checked = false;
            showToast('Подключите Telegram, чтобы включить уведомления', 'error');
            toggleSaveBtn();
            valid = false;
        }
        return valid;
    }

    // ===== АДАПТАЦИЯ ОШИБОК (как в auth) =====
    const ERROR_MAP = [
        { field: null,          match: (s) => s === 409,                                                     text: 'Пользователь с таким email уже существует.' },
        { field: null,          match: (s) => s === 401,                                                     text: 'Неверный email или пароль.' },
        { field: null,          match: (s) => s === 403,                                                     text: 'Доступ запрещён.' },
        { field: null,          match: (s) => s === 429,                                                     text: 'Слишком много попыток. Подождите немного.' },
        { field: null,          match: (s) => s >= 500,                                                      text: 'Ошибка на сервере. Попробуйте позже.' },
        { field: 'firstname',   match: (s, m) => m.includes('firstname') || m.includes('first name') || m.includes('имя'), text: 'Некорректное имя.' },
        { field: 'lastname',    match: (s, m) => m.includes('lastname') || m.includes('last name') || m.includes('фамили'), text: 'Некорректная фамилия.' },
        { field: 'phone',       match: (s, m) => m.includes('phone') || m.includes('телефон'),                text: 'Неверный формат телефона.' },
        { field: 'email',       match: (s, m) => m.includes('email') || m.includes('e-mail'),                 text: 'Неверный формат email.' },
        { field: 'notify_new_booking', match: (s, m) => m.includes('telegram') || m.includes('tg') || m.includes('not connected') || m.includes('не подключен'), text: 'Подключите Telegram, чтобы включить уведомления.' },
    ];

    function formatError(err) {
        const status = err.status || 0;
        const msg = (err.message || '').toLowerCase();

        for (const rule of ERROR_MAP) {
            if (rule.match(status, msg)) return rule;
        }
        return { field: null, text: err.message || 'Произошла неизвестная ошибка.' };
    }

    // --- Load user ---
    async function loadUser() {
        showSkeleton(true);
        try {
            const user = await authAPI.self();
            populateForm(user);
        } catch (err) {
            showToast(err.message || 'Не удалось загрузить профиль', 'error');
        } finally {
            showSkeleton(false);
            setTimeout(() => updateWrapHeight(), 1050);
        }
    }

    const firstSection = document.querySelector('.profile-section[data-section="personal"]');
    if (firstSection) {
        firstSection.classList.add('is-active');
        firstSection.removeAttribute('hidden');
    }

    // --- Navigation ---
    navItems.forEach((btn) => {
        btn.addEventListener('click', () => {
            const target = btn.dataset.section;
            if (target) switchSection(target);
        });
    });

    // --- Detect changes on any input ---
    form.addEventListener('input', toggleSaveBtn);
    form.addEventListener('change', toggleSaveBtn);

    // --- Submit ---
    form.addEventListener('submit', async (event) => {
        event.preventDefault();

        if (!isDirty()) {
            showToast('Нет изменений для сохранения', '');
            return;
        }

        if (!validateAllFields()) return;

        saveBtn.disabled = true;
        saveBtn.textContent = 'Сохраняем…';

        try {
            const current = captureFormState();
            const payload = {};

            if (current.firstname !== initialState.firstname) payload.firstname = current.firstname || '';
            if (current.lastname !== initialState.lastname) payload.lastname = current.lastname || '';
            if (current.phone !== initialState.phone) payload.phone_number = phoneDigits(current.phone) || '';
            if (current.notify_new_booking !== initialState.notify_new_booking) {
                payload.notification = { telegram_enable: current.notify_new_booking };
            }
            if (current.on_vacation !== initialState.on_vacation) {
                payload.on_vacation = current.on_vacation;
            }

            await authAPI.update(
                payload.notification || {},
                payload.firstname || null,
                payload.lastname || null,
                payload.phone_number || null,
                payload.on_vacation != null ? payload.on_vacation : null
            );

            const user = await authAPI.self();
            populateForm(user);

            showToast('Изменения сохранены', 'success');
        } catch (err) {
            const e = formatError(err);
            if (e.field) {
                const fieldEl = form.elements[e.field];
                if (fieldEl) fieldEl.classList.add('input-error');
            }
            showToast(e.text, 'error');
        } finally {
            saveBtn.disabled = false;
            saveBtn.textContent = 'Сохранить изменения';
            toggleSaveBtn();
        }
    });

    // --- Telegram bind/unbind ---
    if (tgBindBtn) {
        tgBindBtn.addEventListener('click', async () => {
            tgBindBtn.disabled = true;
            try {
                if (tgBound) {
                    await authAPI.unbind_tg();
                    tgBound = false;
                    updateTgUI();
                    showToast('Telegram отвязан', 'success');
                    const user = await authAPI.self();
                    populateForm(user);
                } else {
                    const binding = await authAPI.bind_tg();
                    if (binding && binding.bot_url) {
                        window.open(binding.bot_url, '_blank', 'noopener,noreferrer');
                        // Без уведомления — клиент переходит на другую страницу
                    } else {
                        showToast('Не удалось получить ссылку для привязки Telegram', 'error');
                    }
                }
            } catch (err) {
                showToast('Не удалось выполнить операцию с Telegram', 'error');
            } finally {
                tgBindBtn.disabled = false;
            }
        });
    }

    // --- Change password (with cooldown, persisted in localStorage) ---
    const PASSWORD_COOLDOWN_KEY = 'pwd_cooldown_ts';

    function isPasswordCooldown() {
        const ts = parseInt(localStorage.getItem(PASSWORD_COOLDOWN_KEY), 10);
        return ts && Date.now() < ts;
    }

    function setPasswordCooldown(active) {
        if (active) {
            const expiry = Date.now() + 60000;
            localStorage.setItem(PASSWORD_COOLDOWN_KEY, expiry);
            changePasswordBtn.disabled = true;
            changePasswordBtn.classList.add('btn-disabled');
            changePasswordBtn.textContent = 'Подождите минуту…';
        } else {
            localStorage.removeItem(PASSWORD_COOLDOWN_KEY);
            changePasswordBtn.disabled = false;
            changePasswordBtn.classList.remove('btn-disabled');
            changePasswordBtn.textContent = 'Отправить ссылку для смены пароля';
        }
    }

    // Если кулдаун ещё активен после перезагрузки — применить
    if (isPasswordCooldown()) {
        setPasswordCooldown(true);
        const remaining = parseInt(localStorage.getItem(PASSWORD_COOLDOWN_KEY), 10) - Date.now();
        setTimeout(() => setPasswordCooldown(false), remaining);
    }

    if (changePasswordBtn) {
        changePasswordBtn.addEventListener('click', async () => {
            if (isPasswordCooldown()) {
                showToast('Подождите минуту перед повторной отправкой', 'error');
                return;
            }
            changePasswordBtn.disabled = true;
            changePasswordBtn.textContent = 'Отправляем…';
            try {
                const emailEl = form.querySelector('input[name="email"]');
                const email = emailEl ? emailEl.value : '';
                if (!email) {
                    showToast('Не удалось определить email', 'error');
                    changePasswordBtn.disabled = false;
                    changePasswordBtn.textContent = 'Отправить ссылку для смены пароля';
                    return;
                }
                await authAPI.forget_password(email);
                showToast('Ссылка для смены пароля отправлена на email', 'success');
                setPasswordCooldown(true);
                setTimeout(() => setPasswordCooldown(false), 60000);
            } catch (err) {
                showToast(err.message || 'Ошибка отправки', 'error');
                changePasswordBtn.disabled = false;
                changePasswordBtn.textContent = 'Отправить ссылку для смены пароля';
            }
        });
    }

    // --- Logout ---
    if (logoutBtn) {
        logoutBtn.addEventListener('click', async () => {
            await authAPI.logout();
            window.location.href = '/auth';
        });
    }

    // --- Ctrl+Enter to save ---
    document.addEventListener('keydown', (event) => {
        if ((event.ctrlKey || event.metaKey) && event.key === 'Enter') {
            if (isDirty()) form.requestSubmit();
        }
    });

    // Init
    setupLiveValidation();
    loadUser();
})();
