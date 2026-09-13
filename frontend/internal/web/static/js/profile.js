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
        if (collectScheduleDiffs().length > 0) return true;
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
        tgBound = !!(user.tg_chat_id && user.tg_chat_id > 0);
        updateTgUI();

        initialState = captureFormState();
        toggleSaveBtn();
        updateHero(user.firstname || '', user.lastname || '');
    }

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

    const PHONE_RE = /^[\+\d\s\-]*$/;

    function phoneDigits(val) {
        return (val || '').replace(/[^\d]/g, '');
    }

    function formatPhone(val) {
        let d = phoneDigits(val);
        if (d.length === 0) return '';
        if (d.startsWith('8')) d = '7' + d.slice(1);
        if (d.length > 11) d = d.slice(0, 11);
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

        const phoneEl = form.elements['phone'];
        if (phoneEl) {
            phoneEl.addEventListener('input', function () {
                const cursor = this.selectionStart;
                const oldLen = (this.value || '').length;
                const digits = phoneDigits(this.value);
                if (digits.length > 11) return;
                this.value = formatPhone(digits);
                const newLen = this.value.length;
                const shift = newLen - oldLen;
                this.setSelectionRange(cursor + shift, cursor + shift);
            });
            phoneEl.addEventListener('keydown', function (e) {
                if (e.key === 'Backspace' || e.key === 'Delete') return;
                if (phoneDigits(this.value).length >= 11) {
                    e.preventDefault();
                }
            });
        }
    }

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
        if (notifyCb && notifyCb.checked && !tgBound) {
            notifyCb.checked = false;
            showToast('Подключите Telegram, чтобы включить уведомления', 'error');
            toggleSaveBtn();
            valid = false;
        }
        return valid;
    }

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

    navItems.forEach((btn) => {
        btn.addEventListener('click', () => {
            const target = btn.dataset.section;
            if (target) switchSection(target);
        });
    });

    form.addEventListener('input', toggleSaveBtn);
    form.addEventListener('change', toggleSaveBtn);

    document.addEventListener('input', (e) => {
        if (e.target.closest('.schedule-inline-form')) toggleSaveBtn();
    });
    document.addEventListener('change', (e) => {
        if (e.target.closest('.schedule-inline-form')) toggleSaveBtn();
    });

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

            const profileChanged = Object.keys(payload).length > 0;

            if (profileChanged) {
                await authAPI.update(
                    payload.notification || {},
                    'firstname' in payload ? payload.firstname : null,
                    'lastname' in payload ? payload.lastname : null,
                    'phone_number' in payload ? payload.phone_number : null,
                    payload.on_vacation != null ? payload.on_vacation : null
                );
                const user = await authAPI.self();
                populateForm(user);
                tgBound = !!(user.tg_chat_id && user.tg_chat_id > 0);
                updateTgUI();
            }

            const scheduleDiffs = collectScheduleDiffs();
            for (const diff of scheduleDiffs) {
                await scheduleAPI.update({
                    scheduleId: diff.scheduleId,
                    startTime: diff.startTime,
                    endTime: diff.endTime,
                    doUpdateBreak: true,
                    breaks: diff.breaks
                });
            }

            if (scheduleDiffs.length > 0) {
                await loadAllSchedules();
            }

            if (profileChanged || scheduleDiffs.length > 0) {
                showToast('Изменения сохранены', 'success');
            }
        } catch (err) {
            const e = formatError(err);
            if (e.field) {
                const fieldEl = form.elements[e.field];
                if (fieldEl) fieldEl.classList.add('input-error');
            }
            showToast(mapScheduleError(err) !== err.message ? mapScheduleError(err) : e.text, 'error');
        } finally {
            saveBtn.disabled = false;
            saveBtn.textContent = 'Сохранить изменения';
            toggleSaveBtn();
        }
    });

    if (tgBindBtn) {
        tgBindBtn.addEventListener('click', async () => {
            tgBindBtn.disabled = true;
            try {
                if (tgBound) {
                    await authAPI.unbind_tg();
                    showToast('Telegram отвязан', 'success');
                    const user = await authAPI.self();
                    tgBound = !!(user.tg_chat_id && user.tg_chat_id > 0);
                    updateTgUI();
                    populateForm(user);
                } else {
                    const binding = await authAPI.bind_tg();
                    if (binding && binding.bot_url) {
                        window.open(binding.bot_url, '_blank', 'noopener,noreferrer');
                        const user = await authAPI.self();
                        tgBound = !!(user.tg_chat_id && user.tg_chat_id > 0);
                        updateTgUI();
                        populateForm(user);
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

    if (logoutBtn) {
        logoutBtn.addEventListener('click', async () => {
            await authAPI.logout();
            window.location.href = '/auth';
        });
    }

    document.addEventListener('keydown', (event) => {
        if ((event.ctrlKey || event.metaKey) && event.key === 'Enter') {
            if (isDirty()) form.requestSubmit();
        }
    });

    const scheduleDaysCarousel = document.getElementById('scheduleDaysCarousel');
    const scheduleDayContent = document.getElementById('scheduleDayContent');
    const scheduleEmpty = document.getElementById('scheduleEmpty');
    const scheduleModalOverlay = document.getElementById('scheduleModalOverlay');
    const scheduleModalTitle = document.getElementById('scheduleModalTitle');
    const scheduleStartTime = document.getElementById('scheduleStartTime');
    const scheduleEndTime = document.getElementById('scheduleEndTime');
    const scheduleBreaksList = document.getElementById('scheduleBreaksList');
    const scheduleAddBreakBtn = document.getElementById('scheduleAddBreakBtn');
    const scheduleModalSave = document.getElementById('scheduleModalSave');
    const scheduleModalCancel = document.getElementById('scheduleModalCancel');
    const scheduleModalClose = document.getElementById('scheduleModalClose');

    let scheduleCache = {};
    let selectedDay = null;
    let editingScheduleId = null;

    const DAY_NAMES_FULL = ['Воскресенье', 'Понедельник', 'Вторник', 'Среда', 'Четверг', 'Пятница', 'Суббота'];

    let scheduleUserId = null;

    function toDayTime(timeStr) {
        if (!timeStr) return { hour: 0, minute: 0 };
        const [h, m] = timeStr.split(':').map(Number);
        return { hour: h || 0, minute: m || 0 };
    }

    async function loadAllSchedules() {
        if (!scheduleUserId) {
            try {
                const user = await authAPI.self();
                scheduleUserId = user.id;
            } catch (e) {
                showToast('Не удалось определить пользователя', 'error');
                return;
            }
        }

        try {
            const result = await scheduleAPI.listByUserId(scheduleUserId);
            const allSchedules = result.list || [];

            for (let d = 0; d < 7; d++) {
                scheduleCache[d] = [];
            }

            allSchedules.forEach(sch => {
                const day = sch.day_of_week;
                if (day >= 0 && day <= 6) {
                    if (!scheduleCache[day]) scheduleCache[day] = [];
                    scheduleCache[day].push(sch);
                }
            });
        } catch (err) {
            showToast('Не удалось загрузить расписание', 'error');
            for (let d = 0; d < 7; d++) {
                scheduleCache[d] = [];
            }
        }

        updateDayButtonsIndicators();

        if (selectedDay != null) {
            renderDayContent(selectedDay);
        }
    }

    function renderDayContent(dayIndex) {
        if (!scheduleDayContent || !scheduleEmpty) return;

        const schedules = scheduleCache[dayIndex] || [];

        const existingCards = scheduleDayContent.querySelectorAll('.schedule-card, .schedule-add-btn, .schedule-inline-form');
        existingCards.forEach(el => el.remove());

        if (schedules.length === 0) {
            scheduleEmpty.style.display = '';
            scheduleEmpty.textContent = 'Нет расписаний на этот день';

            const addBtn = document.createElement('button');
            addBtn.type = 'button';
            addBtn.className = 'profile-btn profile-btn-ghost schedule-add-btn';
            addBtn.textContent = '+ Добавить расписание';
            addBtn.addEventListener('click', () => openScheduleModal(null));
            scheduleDayContent.appendChild(addBtn);
        } else {
            scheduleEmpty.style.display = 'none';

            schedules.sort((a, b) => compareDayTime(a.start_time, b.start_time));

            schedules.forEach(sch => {
                const form = createScheduleInlineForm(sch);
                scheduleDayContent.appendChild(form);
            });
        }
    }

    function createScheduleInlineForm(sch) {
        const form = document.createElement('div');
        form.className = 'schedule-inline-form';
        form.dataset.scheduleId = sch.id;

        const timeRow = document.createElement('div');
        timeRow.className = 'schedule-inline-row';

        const startInput = document.createElement('input');
        startInput.type = 'time';
        startInput.className = 'profile-input schedule-inline-time';
        startInput.value = formatTime(sch.start_time);

        const sep = document.createElement('span');
        sep.className = 'schedule-inline-sep';
        sep.textContent = '—';

        const endInput = document.createElement('input');
        endInput.type = 'time';
        endInput.className = 'profile-input schedule-inline-time';
        endInput.value = formatTime(sch.end_time);

        timeRow.appendChild(startInput);
        timeRow.appendChild(sep);
        timeRow.appendChild(endInput);
        form.appendChild(timeRow);

        const breaksContainer = document.createElement('div');
        breaksContainer.className = 'schedule-inline-breaks';

        const breaksLabel = document.createElement('span');
        breaksLabel.className = 'schedule-inline-breaks-label';
        breaksLabel.textContent = 'Перерывы:';
        breaksContainer.appendChild(breaksLabel);

        const breaksList = document.createElement('div');
        breaksList.className = 'schedule-inline-breaks-list';

        const addBreakRowInline = (bStart, bEnd) => {
            const row = document.createElement('div');
            row.className = 'schedule-inline-break-row';

            const bs = document.createElement('input');
            bs.type = 'time';
            bs.className = 'profile-input schedule-inline-time';
            bs.value = formatTime(bStart);

            const bSep = document.createElement('span');
            bSep.textContent = '–';

            const be = document.createElement('input');
            be.type = 'time';
            be.className = 'profile-input schedule-inline-time';
            be.value = formatTime(bEnd);

            const rm = document.createElement('button');
            rm.type = 'button';
            rm.className = 'schedule-inline-break-remove';
            rm.setAttribute('aria-label', 'Удалить перерыв');
            rm.innerHTML = '<svg width="14" height="14" viewBox="0 0 16 16" fill="none"><path d="M12 4L4 12M4 4l8 8" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>';
            rm.addEventListener('click', () => row.remove());

            row.appendChild(bs);
            row.appendChild(bSep);
            row.appendChild(be);
            row.appendChild(rm);
            breaksList.appendChild(row);
        };

        if (sch.breaks && sch.breaks.length > 0) {
            sch.breaks.forEach(b => addBreakRowInline(b.start_time, b.end_time));
        }

        const addBreakBtn = document.createElement('button');
        addBreakBtn.type = 'button';
        addBreakBtn.className = 'profile-btn profile-btn-ghost schedule-inline-add-break';
        addBreakBtn.textContent = '+ Перерыв';
        addBreakBtn.addEventListener('click', () => {
            addBreakRowInline({ hour: 12, minute: 0 }, { hour: 13, minute: 0 });
            toggleSaveBtn();
        });

        breaksContainer.appendChild(breaksList);
        breaksContainer.appendChild(addBreakBtn);
        form.appendChild(breaksContainer);

        form.addEventListener('input', (e) => {
            if (e.target.matches('input[type="time"]')) {
                toggleSaveBtn();
            }
        });
        form.addEventListener('click', (e) => {
            if (e.target.closest('.schedule-inline-break-remove')) {
                setTimeout(toggleSaveBtn, 0);
            }
        });

        const actions = document.createElement('div');
        actions.className = 'schedule-inline-actions';

        const deleteBtn = document.createElement('button');
        deleteBtn.type = 'button';
        deleteBtn.className = 'profile-btn profile-btn-ghost schedule-inline-delete';
        deleteBtn.textContent = 'Удалить';
        deleteBtn.addEventListener('click', () => deleteSchedule(sch.id));

        actions.appendChild(deleteBtn);
        form.appendChild(actions);

        return form;
    }

    function collectScheduleDiffs() {
        if (!scheduleDayContent) return [];
        const forms = scheduleDayContent.querySelectorAll('.schedule-inline-form');
        const diffs = [];

        forms.forEach(form => {
            const scheduleId = form.dataset.scheduleId;
            if (!scheduleId) return;

            let original = null;
            for (let d = 0; d < 7; d++) {
                const found = (scheduleCache[d] || []).find(s => s.id === scheduleId);
                if (found) { original = found; break; }
            }
            if (!original) return;

            const startInput = form.querySelector('.schedule-inline-row .schedule-inline-time:first-child');
            const endInput = form.querySelector('.schedule-inline-row .schedule-inline-time:last-child');
            if (!startInput || !endInput) return;

            const newStart = startInput.value;
            const newEnd = endInput.value;
            const origStart = formatTime(original.start_time);
            const origEnd = formatTime(original.end_time);

            const breakRows = form.querySelectorAll('.schedule-inline-break-row');
            const newBreaks = [];
            breakRows.forEach(row => {
                const inputs = row.querySelectorAll('input[type="time"]');
                if (inputs.length === 2) {
                    const bs = inputs[0].value;
                    const be = inputs[1].value;
                    if (bs && be && bs < be) {
                        newBreaks.push({ start: bs, end: be });
                    }
                }
            });

            const origBreaks = (original.breaks || []).map(b => ({
                start: formatTime(b.start_time),
                end: formatTime(b.end_time)
            }));

            const timeChanged = newStart !== origStart || newEnd !== origEnd;
            let breaksChanged = newBreaks.length !== origBreaks.length;
            if (!breaksChanged) {
                for (let i = 0; i < newBreaks.length; i++) {
                    if (newBreaks[i].start !== origBreaks[i].start || newBreaks[i].end !== origBreaks[i].end) {
                        breaksChanged = true;
                        break;
                    }
                }
            }

            if (timeChanged || breaksChanged) {
                diffs.push({
                    scheduleId: scheduleId,
                    startTime: toDayTime(newStart),
                    endTime: toDayTime(newEnd),
                    breaks: newBreaks.map(b => ({ start_time: toDayTime(b.start), end_time: toDayTime(b.end) }))
                });
            }
        });

        return diffs;
    }

    function formatTime(dt) {
        if (!dt) return '--:--';
        if (typeof dt === 'object' && 'hour' in dt && 'minute' in dt) {
            return String(dt.hour).padStart(2, '0') + ':' + String(dt.minute).padStart(2, '0');
        }
        if (typeof dt === 'string') {
            return dt.slice(0, 5);
        }
        return '--:--';
    }

    function compareDayTime(a, b) {
        const ha = (typeof a === 'object' && 'hour' in a) ? a.hour : 0;
        const ma = (typeof a === 'object' && 'minute' in a) ? a.minute : 0;
        const hb = (typeof b === 'object' && 'hour' in b) ? b.hour : 0;
        const mb = (typeof b === 'object' && 'minute' in b) ? b.minute : 0;
        if (ha !== hb) return ha - hb;
        return ma - mb;
    }

    function updateDayButtonsIndicators() {
        if (!scheduleDaysCarousel) return;
        const buttons = scheduleDaysCarousel.querySelectorAll('.schedule-day-btn');
        buttons.forEach(btn => {
            const day = parseInt(btn.dataset.day, 10);
            const hasSchedule = (scheduleCache[day] && scheduleCache[day].length > 0);
            btn.classList.toggle('has-schedule', hasSchedule);
        });
    }

    function openScheduleModal(schedule) {
        if (!scheduleModalOverlay) return;

        editingScheduleId = schedule ? schedule.id : null;

        if (scheduleModalTitle) {
            scheduleModalTitle.textContent = schedule ? 'Редактировать расписание' : 'Добавить расписание';
        }

        if (scheduleStartTime) scheduleStartTime.value = schedule ? formatTime(schedule.start_time) : '09:00';
        if (scheduleEndTime) scheduleEndTime.value = schedule ? formatTime(schedule.end_time) : '18:00';

        if (scheduleBreaksList) {
            scheduleBreaksList.innerHTML = '';

            const breaks = (schedule && schedule.breaks) ? schedule.breaks : [];
            breaks.forEach(b => addBreakRow(b.start_time, b.end_time));
        }

        scheduleModalOverlay.removeAttribute('hidden');
    }

    function closeScheduleModal() {
        if (!scheduleModalOverlay) return;
        scheduleModalOverlay.setAttribute('hidden', '');
        editingScheduleId = null;
    }

    function addBreakRow(startTime, endTime) {
        if (!scheduleBreaksList) return;

        const row = document.createElement('div');
        row.className = 'schedule-break-row';

        const startInput = document.createElement('input');
        startInput.type = 'time';
        startInput.className = 'profile-input';
        startInput.value = startTime ? formatTime(startTime) : '12:00';

        const endInput = document.createElement('input');
        endInput.type = 'time';
        endInput.className = 'profile-input';
        endInput.value = endTime ? formatTime(endTime) : '13:00';

        const removeBtn = document.createElement('button');
        removeBtn.type = 'button';
        removeBtn.className = 'schedule-break-remove';
        removeBtn.setAttribute('aria-label', 'Удалить перерыв');
        removeBtn.innerHTML = '<svg width="16" height="16" viewBox="0 0 16 16" fill="none"><path d="M12 4L4 12M4 4l8 8" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/></svg>';
        removeBtn.addEventListener('click', () => row.remove());

        row.appendChild(startInput);
        row.appendChild(endInput);
        row.appendChild(removeBtn);
        scheduleBreaksList.appendChild(row);
    }

    function validateBreaks(breaks, workStart, workEnd) {
        if (breaks.length === 0) return null;

        const sorted = [...breaks].sort((a, b) => {
            if (a.start_time.hour !== b.start_time.hour) return a.start_time.hour - b.start_time.hour;
            return a.start_time.minute - b.start_time.minute;
        });

        const workStartMin = workStart.hour * 60 + workStart.minute;
        const workEndMin = workEnd.hour * 60 + workEnd.minute;

        for (let i = 0; i < sorted.length; i++) {
            const bs = sorted[i].start_time.hour * 60 + sorted[i].start_time.minute;
            const be = sorted[i].end_time.hour * 60 + sorted[i].end_time.minute;

            if (bs < workStartMin || be > workEndMin) {
                return 'Перерыв должен быть внутри рабочего времени';
            }

            if (be - bs < 5) {
                return 'Перерыв должен быть не менее 5 минут';
            }

            if (i > 0) {
                const prevEnd = sorted[i - 1].end_time.hour * 60 + sorted[i - 1].end_time.minute;
                if (bs < prevEnd) {
                    return 'Перерывы не должны пересекаться';
                }
            }
        }

        return null;
    }

    function mapScheduleError(err) {
        const msg = (err && err.message) ? err.message : '';
        if (!msg) return 'Ошибка при сохранении расписания';

        const lower = msg.toLowerCase();
        if (lower.includes('times of schedule must be more than 5 minutes') || lower.includes('errinvalidschedule')) {
            return 'Рабочее время должно быть не менее 5 минут';
        }
        if (lower.includes('invalid day of week') || lower.includes('errinvaliddayofweek')) {
            return 'Некорректный день недели';
        }
        if (lower.includes('invalid break') || lower.includes('errinvalidbreak')) {
            return 'Некорректный перерыв (проверьте время)';
        }
        if (lower.includes('some breaks is overlap') || lower.includes('errbreakoverlap')) {
            return 'Перерывы не должны пересекаться';
        }
        if (lower.includes('invalid user') || lower.includes('errinvaliduserid')) {
            return 'Ошибка авторизации';
        }

        return msg;
    }

    async function saveSchedule() {
        if (!scheduleUserId || selectedDay == null) {
            showToast('Выберите день недели', 'error');
            return;
        }

        const startTime = scheduleStartTime ? scheduleStartTime.value : '';
        const endTime = scheduleEndTime ? scheduleEndTime.value : '';

        if (!startTime || !endTime) {
            showToast('Укажите время начала и конца работы', 'error');
            return;
        }

        if (startTime >= endTime) {
            showToast('Время начала должно быть раньше времени конца', 'error');
            return;
        }

        const [sh, sm] = startTime.split(':').map(Number);
        const [eh, em] = endTime.split(':').map(Number);
        const workDuration = (eh * 60 + em) - (sh * 60 + sm);
        if (workDuration < 5) {
            showToast('Рабочее время должно быть не менее 5 минут', 'error');
            return;
        }

        const breakRows = scheduleBreaksList ? scheduleBreaksList.querySelectorAll('.schedule-break-row') : [];
        const breaks = [];
        for (const row of breakRows) {
            const inputs = row.querySelectorAll('input[type="time"]');
            if (inputs.length === 2) {
                const bStart = inputs[0].value;
                const bEnd = inputs[1].value;
                if (bStart && bEnd && bStart < bEnd) {
                    breaks.push({ start_time: toDayTime(bStart), end_time: toDayTime(bEnd) });
                }
            }
        }

        const breakError = validateBreaks(breaks, { hour: sh, minute: sm }, { hour: eh, minute: em });
        if (breakError) {
            showToast(breakError, 'error');
            return;
        }

        scheduleModalSave.disabled = true;
        scheduleModalSave.textContent = 'Сохраняем…';

        try {
            if (editingScheduleId) {
                const updated = await scheduleAPI.update({
                    scheduleId: editingScheduleId,
                    startTime: toDayTime(startTime),
                    endTime: toDayTime(endTime),
                    doUpdateBreak: true,
                    breaks: breaks
                });
                const day = updated.day_of_week;
                if (scheduleCache[day]) {
                    const idx = scheduleCache[day].findIndex(s => s.id === editingScheduleId);
                    if (idx !== -1) scheduleCache[day][idx] = updated;
                }
            } else {
                const created = await scheduleAPI.create({
                    dayOfWeek: selectedDay,
                    startTime: toDayTime(startTime),
                    endTime: toDayTime(endTime),
                    breaks: breaks
                });
                const day = created.day_of_week;
                if (!scheduleCache[day]) scheduleCache[day] = [];
                scheduleCache[day].push(created);
            }

            closeScheduleModal();
            if (selectedDay != null) {
                updateDayButtonsIndicators();
                renderDayContent(selectedDay);
            }
            showToast('Расписание сохранено', 'success');
        } catch (err) {
            showToast(mapScheduleError(err), 'error');
        } finally {
            scheduleModalSave.disabled = false;
            scheduleModalSave.textContent = 'Сохранить';
        }
    }

    async function deleteSchedule(scheduleId) {
        if (!scheduleUserId || selectedDay == null) return;

        try {
            await scheduleAPI.delete(scheduleId);
            for (let d = 0; d < 7; d++) {
                if (scheduleCache[d]) {
                    scheduleCache[d] = scheduleCache[d].filter(s => s.id !== scheduleId);
                }
            }
            updateDayButtonsIndicators();
            if (selectedDay != null) {
                renderDayContent(selectedDay);
            }
            showToast('Расписание удалено', 'success');
        } catch (err) {
            showToast(mapScheduleError(err), 'error');
        }
    }

    if (scheduleDaysCarousel) {
        scheduleDaysCarousel.addEventListener('click', (e) => {
            const btn = e.target.closest('.schedule-day-btn');
            if (!btn) return;

            const dayIndex = parseInt(btn.dataset.day, 10);
            if (isNaN(dayIndex)) return;

            scheduleDaysCarousel.querySelectorAll('.schedule-day-btn').forEach(b => b.classList.remove('active'));
            btn.classList.add('active');

            selectedDay = dayIndex;
            renderDayContent(dayIndex);
        });
    }

    if (scheduleAddBreakBtn) {
        scheduleAddBreakBtn.addEventListener('click', () => addBreakRow(null, null));
    }

    if (scheduleModalSave) {
        scheduleModalSave.addEventListener('click', saveSchedule);
    }

    if (scheduleModalCancel) {
        scheduleModalCancel.addEventListener('click', closeScheduleModal);
    }

    if (scheduleModalClose) {
        scheduleModalClose.addEventListener('click', closeScheduleModal);
    }

    if (scheduleModalOverlay) {
        scheduleModalOverlay.addEventListener('click', (e) => {
            if (e.target === scheduleModalOverlay) closeScheduleModal();
        });
    }

    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape' && scheduleModalOverlay && !scheduleModalOverlay.hasAttribute('hidden')) {
            closeScheduleModal();
        }
    });

    const originalPopulateForm = populateForm;
    populateForm = function(user) {
        originalPopulateForm(user);
        if (user && user.id) {
            scheduleUserId = user.id;
            if (selectedDay == null) {
                selectedDay = 1;
                if (scheduleDaysCarousel) {
                    scheduleDaysCarousel.querySelectorAll('.schedule-day-btn').forEach(b => b.classList.remove('active'));
                    const activeBtn = scheduleDaysCarousel.querySelector('.schedule-day-btn[data-day="1"]');
                    if (activeBtn) activeBtn.classList.add('active');
                }
            }
            loadAllSchedules();
        }
    };

    const originalSwitchSection = switchSection;
    switchSection = function(sectionName) {
        originalSwitchSection(sectionName);
        if (sectionName === 'schedule' && selectedDay == null) {
            const dayIndex = 0;
            selectedDay = dayIndex;

            if (scheduleDaysCarousel) {
                scheduleDaysCarousel.querySelectorAll('.schedule-day-btn').forEach(b => b.classList.remove('active'));
                const activeBtn = scheduleDaysCarousel.querySelector(`.schedule-day-btn[data-day="${dayIndex}"]`);
                if (activeBtn) activeBtn.classList.add('active');
            }

            if (scheduleUserId) {
                loadAllSchedules().then(() => {
                    if (selectedDay != null) renderDayContent(selectedDay);
                });
            }
        }
    };

    setupLiveValidation();
    loadUser();
})();
