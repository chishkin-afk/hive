// ===== МОДУЛЬ ЗАПИСИ =====
(function() {
    'use strict';
    
    // Защита от повторной инициализации
    if (window.__appointmentModuleInitialized) return;
    window.__appointmentModuleInitialized = true;
    
    // Получаем API из глобальной области видимости
    const offerAPI = window.offerAPI;
    const bookingAPI = window.bookingAPI;
    const scheduleAPI = window.scheduleAPI;
    const authAPI = window.authAPI;
    
    // ===== КОНФИГУРАЦИЯ =====
    const pathParts = window.location.pathname.split('/').filter(Boolean);
    const APP_CONFIG = {
        offerId: pathParts[1] || '',
        bookingId: new URLSearchParams(window.location.search).get('booking_id'),
        mode: 'create'
    };
    APP_CONFIG.mode = APP_CONFIG.bookingId ? 'edit' : 'create';
    
    const OFFER_ID = APP_CONFIG.offerId;
    const BOOKING_ID = APP_CONFIG.bookingId;
    const MODE = APP_CONFIG.mode;
    
    // ===== СОСТОЯНИЕ =====
    let currentOffer = null;
    let selectedDate = null;
    let selectedTime = null;
    let existingBooking = null;
    
    // ===== DOM ЭЛЕМЕНТЫ =====
    let daysSelector, backBtn, forwardBtn, monthTitle;
    let confirmationOverlay, confirmationCloseBtn, confirmationSubmitBtn;
    let offerDateEl, offerTimeEl, groupInput;
    
    // ===== ИНИЦИАЛИЗАЦИЯ DOM =====
    function initDOM() {
        daysSelector = document.getElementById('daysSelector');
        backBtn = document.getElementById('backBtn');
        forwardBtn = document.getElementById('forwardBtn');
        monthTitle = document.getElementById('monthTitle');
        confirmationOverlay = document.getElementById('confirmationOverlay');
        confirmationCloseBtn = document.getElementById('confirmationCloseBtn');
        confirmationSubmitBtn = document.getElementById('confirmationSubmitBtn');
        offerDateEl = document.querySelector('.offer-date');
        offerTimeEl = document.querySelector('.offer-time');
        groupInput = document.getElementById('groupInput');
    }
    
    // ===== ПРОВЕРКА АВТОРИЗАЦИИ =====
    async function checkAuth() {
        try {
            await window.authAPI.self();
            return true;
        } catch (error) {
            // 400 или 401 — пользователь не авторизован, это нормально
            if (error.status === 400 || error.status === 401) {
                // Если токен невалидный, очищаем куки
                if (error.status === 401) {
                    await fetch('/api/v1/auth/logout', {
                        method: 'POST',
                        credentials: 'include'
                    }).catch(() => {});
                }
                return false;
            }
            // Другие ошибки — логируем
            console.error('Auth check failed:', error);
            return false;
        }
    }
    
    // ===== ГЕНЕРАЦИЯ КАЛЕНДАРЯ =====
    async function generateCalendar() {
        if (!daysSelector) return;
        daysSelector.innerHTML = '';
        
        const today = new Date();
        today.setHours(0, 0, 0, 0);
        
        const startDate = new Date(today);
        const endDate = new Date(startDate);
        endDate.setDate(endDate.getDate() + 14);
        
        const monthNames = ['Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь', 
                            'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь'];
        monthTitle.textContent = `${monthNames[startDate.getMonth()]} ${startDate.getFullYear()}`;
        
        const activeDaysMap = {};
        
        try {
            if (currentOffer?.author_id) {
                const allSchedules = await window.scheduleAPI.listByUserId(currentOffer.author_id);
                
                if (allSchedules.list && allSchedules.list.length > 0) {
                    allSchedules.list.forEach(schedule => {
                        const dayOfWeek = schedule.day_of_week ?? schedule.dayOfWeek;
                        let startTime = schedule.start_time;
                        let endTime = schedule.end_time;
                        
                        if (typeof schedule.start_time === 'object' && schedule.start_time.hour !== undefined) {
                            const h = String(schedule.start_time.hour).padStart(2, '0');
                            const m = String(schedule.start_time.minute || 0).padStart(2, '0');
                            startTime = `${h}:${m}`;
                        }
                        if (typeof schedule.end_time === 'object' && schedule.end_time.hour !== undefined) {
                            const h = String(schedule.end_time.hour).padStart(2, '0');
                            const m = String(schedule.end_time.minute || 0).padStart(2, '0');
                            endTime = `${h}:${m}`;
                        }
                        
                        if (dayOfWeek !== undefined && startTime && endTime) {
                            activeDaysMap[dayOfWeek] = {
                                ...schedule,
                                start_time: startTime,
                                end_time: endTime
                            };
                        }
                    });
                }
            }
        } catch (error) {
            console.error('Failed to load schedule:', error);
        }
        
        window.__activeDaysMap = activeDaysMap;
        
        const activeDayList = [];
        for (let date = new Date(startDate); date <= endDate; date.setDate(date.getDate() + 1)) {
            const dbDayOfWeek = date.getDay();
            if (activeDaysMap[dbDayOfWeek]) {
                const year = date.getFullYear();
                const month = String(date.getMonth() + 1).padStart(2, '0');
                const day = String(date.getDate()).padStart(2, '0');
                activeDayList.push(`${year}-${month}-${day}`);
            }
        }
        
        for (let date = new Date(startDate); date <= endDate; date.setDate(date.getDate() + 1)) {
            const year = date.getFullYear();
            const month = String(date.getMonth() + 1).padStart(2, '0');
            const day = String(date.getDate()).padStart(2, '0');
            const dateStr = `${year}-${month}-${day}`;
            
            const dayCard = document.createElement('button');
            dayCard.className = 'day-card';
            dayCard.dataset.date = dateStr;
            
            const dayNumber = date.getDate();
            const dayName = date.toLocaleDateString('ru-RU', { weekday: 'short' }).toUpperCase();
            const dbDayOfWeek = date.getDay();
            const isWeekend = dbDayOfWeek === 0 || dbDayOfWeek === 6;
            const hasSchedule = activeDayList.includes(dateStr);
            
            if (!hasSchedule) {
                dayCard.classList.add('inactive');
            } else if (isWeekend) {
                dayCard.classList.add('weekend');
            }
            
            if (selectedDate && dateStr === selectedDate) {
                dayCard.classList.add('active');
            }
            
            dayCard.innerHTML = `<span class="day-number">${dayNumber}</span><span class="day-name">${dayName}</span>`;
            dayCard.addEventListener('click', () => {
                if (!dayCard.classList.contains('inactive')) {
                    selectDay(dayCard);
                }
            });
            daysSelector.appendChild(dayCard);
        }
        
        const todayStr = today.toISOString().split('T')[0];
        const todayCard = daysSelector.querySelector(`[data-date="${todayStr}"]:not(.inactive)`) ||
                          daysSelector.querySelector('.day-card:not(.inactive)');
        if (todayCard && !selectedDate) {
            selectDay(todayCard);
        }
    }
    
    function selectDay(dayCard) {
        if (dayCard.classList.contains('active')) return;
        
        document.querySelectorAll('.day-card').forEach(card => card.classList.remove('active'));
        dayCard.classList.add('active');
        selectedDate = dayCard.dataset.date;
        loadTimeSlots(selectedDate);
    }
    
    async function loadTimeSlots(date) {
        try {
            const dateStr = date.split('T')[0];
            const slotsData = await window.bookingAPI.slots(OFFER_ID, dateStr);
            
            let times = [];
            let rawTimes = [];
            
            if (Array.isArray(slotsData)) {
                rawTimes = slotsData;
            } else if (slotsData.times && Array.isArray(slotsData.times)) {
                rawTimes = slotsData.times;
            } else if (slotsData.list && Array.isArray(slotsData.list)) {
                rawTimes = slotsData.list;
            }
            
            if (rawTimes && rawTimes.length > 0) {
                rawTimes.forEach(slot => {
                    if (typeof slot === 'string') {
                        times.push(slot);
                    } else if (typeof slot === 'object' && slot.hour !== undefined) {
                        const h = String(slot.hour).padStart(2, '0');
                        const m = String(slot.minute || 0).padStart(2, '0');
                        times.push(`${h}:${m}`);
                    }
                });
            }
            
            renderTimeSlots(times);
        } catch (error) {
            console.error('Failed to load time slots:', error);
            renderTimeSlots([]);
        }
    }
    
    function generateTimeSlots(startTime, endTime, duration, interval, breaks) {
        const times = [];
        const [startHour, startMin] = startTime.split(':').map(Number);
        const [endHour, endMin] = endTime.split(':').map(Number);
        
        let currentTime = startHour * 60 + startMin;
        const endTimeMin = endHour * 60 + endMin;
        
        const sortedBreaks = [...breaks].sort((a, b) => {
            const [aHour, aMin] = a.start_time.split(':').map(Number);
            const [bHour, bMin] = b.start_time.split(':').map(Number);
            return (aHour * 60 + aMin) - (bHour * 60 + bMin);
        });
        
        while (currentTime + duration <= endTimeMin) {
            let inBreak = false;
            for (const brk of sortedBreaks) {
                const [brkStartHour, brkStartMin] = brk.start_time.split(':').map(Number);
                const [brkEndHour, brkEndMin] = brk.end_time.split(':').map(Number);
                const breakStart = brkStartHour * 60 + brkStartMin;
                const breakEnd = brkEndHour * 60 + brkEndMin;
                
                if (currentTime < breakEnd && currentTime + duration > breakStart) {
                    inBreak = true;
                    currentTime = breakEnd;
                    break;
                }
            }
            
            if (!inBreak) {
                const hours = Math.floor(currentTime / 60);
                const minutes = currentTime % 60;
                const timeStr = `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}`;
                times.push(timeStr);
                currentTime += duration + interval;
            }
        }
        
        return times;
    }
    
    function renderTimeSlots(times) {
        const container = document.querySelector('.time-sections');
        if (!container) return;
        
        container.innerHTML = '';
        
        if (!times || times.length === 0) {
            container.innerHTML = `
                <div class="time-section">
                    <span class="time-section-title">Нет доступных слотов</span>
                </div>`;
            return;
        }
        
        const morning = [];
        const day = [];
        const evening = [];
        const night = [];
        
        times.forEach(timeStr => {
            const [hours] = timeStr.split(':').map(Number);
            if (hours >= 6 && hours < 12) {
                morning.push(timeStr);
            } else if (hours >= 12 && hours < 18) {
                day.push(timeStr);
            } else if (hours >= 18 && hours < 24) {
                evening.push(timeStr);
            } else {
                night.push(timeStr);
            }
        });
        
        const sections = [
            { title: 'Утро', slots: morning },
            { title: 'День', slots: day },
            { title: 'Вечер', slots: evening },
            { title: 'Ночь', slots: night }
        ];
        
        sections.forEach(({ title, slots }) => {
            if (slots.length === 0) return;
            
            const section = document.createElement('div');
            section.className = 'time-section';
            
            const titleEl = document.createElement('span');
            titleEl.className = 'time-section-title';
            titleEl.textContent = title;
            section.appendChild(titleEl);
            
            const slotsContainer = document.createElement('div');
            slotsContainer.className = 'time-slots';
            
            slots.forEach(timeStr => {
                const slot = document.createElement('button');
                slot.className = 'time-slot';
                slot.dataset.time = timeStr;
                slot.textContent = timeStr;
                if (selectedTime === timeStr) slot.classList.add('active');
                slot.addEventListener('click', () => {
                    document.querySelectorAll('.time-slot').forEach(s => s.classList.remove('active'));
                    slot.classList.add('active');
                    selectedTime = timeStr;
                });
                slotsContainer.appendChild(slot);
            });
            
            section.appendChild(slotsContainer);
            container.appendChild(section);
        });
    }
    
    function setupCalendarScrollBtns() {
        function getScrollStep() { return window.innerWidth < 768 ? 256 : 512; }
        function checkButtons() {
            const maxScroll = daysSelector.scrollWidth - daysSelector.clientWidth;
            backBtn.classList.toggle('inactive', daysSelector.scrollLeft <= 0);
            forwardBtn.classList.toggle('inactive', daysSelector.scrollLeft >= maxScroll - 1);
        }
        backBtn.addEventListener('click', () => daysSelector.scrollBy({ left: -getScrollStep(), behavior: 'smooth' }));
        forwardBtn.addEventListener('click', () => daysSelector.scrollBy({ left: getScrollStep(), behavior: 'smooth' }));
        daysSelector.addEventListener('scroll', checkButtons);
        window.addEventListener('resize', checkButtons);
        checkButtons();
    }
    
    function showStep(stepNumber) {
        document.querySelectorAll('.confirmation-step').forEach(step => step.classList.remove('active'));
        const activeStep = document.getElementById(`step${stepNumber}`);
        if (activeStep) activeStep.classList.add('active');
    }
    
    function openConfirmation() {
        if (!selectedDate || !selectedTime) {
            showNotification('Выберите дату и время', 'error');
            return;
        }
        
        // Обновляем данные в модалке
        const serviceTitleInModal = document.querySelector('#step1 .service-title');
        const servicePriceInModal = document.querySelector('#step1 .service-amount');
        if (serviceTitleInModal && currentOffer) {
            serviceTitleInModal.textContent = currentOffer.title;
        }
        if (servicePriceInModal && currentOffer) {
            const price = (currentOffer.price_penny ?? currentOffer.price) / 100;
            servicePriceInModal.textContent = price.toFixed(0);
        }
        
        const dateObj = new Date(selectedDate);
        const dateStr = dateObj.toLocaleDateString('ru-RU', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' });
        offerDateEl.textContent = dateStr;
        offerTimeEl.textContent = selectedTime;
        if (MODE === 'edit' && existingBooking) groupInput.value = existingBooking.comment || '';
        else groupInput.value = '';
        
        document.body.classList.add('block-scroll');
        confirmationOverlay.classList.add('open');
        showStep(1);
    }
    
    function closeConfirmation() {
        // Сначала скрываем оверлей
        confirmationOverlay.classList.remove('open');
        document.body.classList.remove('block-scroll');
        
        // Сбрасываем всё состояние модалки после завершения анимации закрытия
        setTimeout(() => {
            // Скрываем все шаги, чтобы не было видно переключения
            document.querySelectorAll('.confirmation-step').forEach(step => step.classList.remove('active'));
            
            // Очищаем поля ввода
            document.querySelectorAll('.confirmation-step input').forEach(input => input.value = '');
            
            // Сбрасываем режим входа (если был переключен)
            const passwordGroup = document.getElementById('passwordGroup');
            const firstnameGroup = document.getElementById('firstnameGroup');
            const forgotBtn = document.getElementById('forgotPasswordBtn');
            const title = document.getElementById('step2Title');
            const submitBtn = document.getElementById('emailSubmitBtn');
            
            passwordGroup.style.display = 'none';
            firstnameGroup.style.display = 'flex';
            forgotBtn.style.display = 'none';
            title.textContent = 'Введите данные';
            submitBtn.textContent = 'Продолжить';
            
            // Показываем шаг 1 уже после того, как оверлей полностью скрыт
            showStep(1);
            
            // Сбрасываем выбранное время
            document.querySelectorAll('.time-slot').forEach(s => s.classList.remove('active'));
            selectedTime = null;
        }, 300);
    }
    
    function parseTime(timeStr) {
        const [hour, minute] = timeStr.split(':').map(Number);
        return { hour: hour || 0, minute: minute || 0 };
    }

    async function createBooking() {
        const comment = groupInput.value.trim();
        const startTime = {
            date: selectedDate,
            start_time: parseTime(selectedTime)
        };
        
        try {
            if (MODE === 'create') {
                await window.bookingAPI.create({ offerId: OFFER_ID, comment: comment, startTime: startTime });
                showNotification('Запись создана!', 'success');
                
                // Заполняем данные в финальном окне
                const dateObj = new Date(selectedDate);
                const dateStr = dateObj.toLocaleDateString('ru-RU', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' });
                document.getElementById('offerDateConfirmed').textContent = dateStr;
                document.getElementById('offerTimeConfirmed').textContent = selectedTime;
                document.querySelector('#step5 .service-name').textContent = currentOffer.title;
                
                showStep(5);
                
                // Обновляем слоты после успешной записи (с задержкой, чтобы модалка успела отрисоваться)
                if (selectedDate) {
                    setTimeout(() => loadTimeSlots(selectedDate), 300);
                }
            } else {
                await window.bookingAPI.update({ bookingId: BOOKING_ID, isCancelled: existingBooking.is_cancelled, comment: comment, startTime: startTime });
                showNotification('Запись обновлена!', 'success');
                // Закрываем модалку после успешного обновления
                setTimeout(() => {
                    closeConfirmation();
                    // Обновляем слоты после изменения записи
                    if (selectedDate) {
                        loadTimeSlots(selectedDate);
                    }
                }, 500);
            }
        } catch (error) {
            console.error('Failed to save booking:', error);
            showNotification(mapErrorMessage(error), 'error');
        }
    }
    
    // ===== ОБРАБОТКА ШАГА 2 (EMAIL/ПАРОЛЬ/ИМЯ) =====
    async function handleStep2() {
        const emailInput = document.getElementById('emailInput');
        const email = emailInput ? emailInput.value.trim() : '';
        const firstnameInput = document.getElementById('firstnameInput');
        const firstname = firstnameInput ? firstnameInput.value.trim() : '';
        const passwordInput = document.getElementById('passwordInput');
        const password = passwordInput ? passwordInput.value.trim() : '';
        const passwordGroup = document.getElementById('passwordGroup');
        const firstnameGroup = document.getElementById('firstnameGroup');
        
        if (!email || !isValidEmail(email)) {
            showNotification('Введите корректный email', 'error');
            // Скрываем поле имени, если оно видно
            const firstnameGroup = document.getElementById('firstnameGroup');
            if (firstnameGroup) firstnameGroup.style.display = 'none';
            return;
        }
        
        window.registrationEmail = email;
        const isLoginMode = passwordGroup.style.display !== 'none';
        
        // === РЕЖИМ ВХОДА ===
        if (isLoginMode) {
            // ПРИНУДИТЕЛЬНО СКРЫВАЕМ ПОЛЕ ИМЕНИ
            firstnameGroup.style.display = 'none';
            
            if (!password || password.length < 6) {
                showNotification('Введите пароль (минимум 6 символов)', 'error');
                passwordInput.focus();
                return; // ВЫХОДИМ ИЗ ФУНКЦИИ
            }
            
            try {
                await window.authAPI.login(email, password);
                showNotification('Вход выполнен успешно!', 'success');
                await createBooking();
                return;
            } catch (error) {
                console.error('Login error:', error);
                if (error.status === 401) {
                    showNotification('Неверный пароль. Попробуйте снова или восстановите пароль.', 'error');
                    passwordInput.value = '';
                    passwordInput.focus();
                } else {
                    showNotification(mapErrorMessage(error), 'error');
                }
                return; // ВЫХОДИМ ИЗ ФУНКЦИИ, ЧТОБЫ НЕ ПОКАЗЫВАТЬ ПОЛЕ ИМЕНИ
            }
        }
        
        // === РЕЖИМ РЕГИСТРАЦИИ ===
        // Проверяем, существует ли пользователь
        let userExists = false;
        try {
            await window.authAPI.login(email, 'dummy_password_for_check');
            showNotification('Введите пароль для входа', 'info');
            return;
        } catch (error) {
            userExists = error.status === 401;
        }
        
        if (userExists) {
            // Переключаем на режим входа
            const title = document.getElementById('step2Title');
            const submitBtn = document.getElementById('emailSubmitBtn');
            const forgotBtn = document.getElementById('forgotPasswordBtn');
            
            passwordGroup.style.display = 'flex';
            firstnameGroup.style.display = 'none';
            title.textContent = 'Введите пароль';
            submitBtn.textContent = 'Войти';
            forgotBtn.textContent = 'Восстановить пароль?';
            forgotBtn.style.display = 'block';
            forgotBtn.onclick = async function() {
                const email = document.getElementById('emailInput').value.trim();
                if (!email || !isValidEmail(email)) {
                    showNotification('Введите корректный email', 'error');
                    return;
                }
                try {
                    await window.authAPI.forget_password(email);
                    showNotification('Ссылка для восстановления отправлена на почту', 'success');
                } catch (error) {
                    showNotification(mapErrorMessage(error), 'error');
                }
            };
            showNotification('Пользователь с таким email уже существует. Введите пароль.', 'info');
            return;
        }
        
        // === РЕГИСТРАЦИЯ: ВВОД ИМЕНИ ===
        if (!firstname || firstname.length === 0) {
            showNotification('Введите имя для регистрации', 'info');
            return;
        }
        
        if (firstname.length < 3) {
            showNotification('Имя должно содержать минимум 3 символа', 'error');
            return;
        }
        
        window.registrationFirstname = firstname;
        const tempPassword = 'temp_' + Math.random().toString(36).slice(-8);
        
        try {
            await window.authAPI.prepare(firstname, '', email, '', tempPassword);
            showNotification('Код подтверждения отправлен на email', 'success');
            showStep(3);
        } catch (error) {
            console.error('Prepare error:', error);
            if (error.status === 409) {
                const title = document.getElementById('step2Title');
                const submitBtn = document.getElementById('emailSubmitBtn');
                const forgotBtn = document.getElementById('forgotPasswordBtn');
                
                passwordGroup.style.display = 'flex';
                firstnameGroup.style.display = 'none';
                title.textContent = 'Введите пароль';
                submitBtn.textContent = 'Войти';
                forgotBtn.style.display = 'block';
                forgotBtn.onclick = function() {
                    window.location.href = '/reset';
                };
                showNotification('Пользователь с таким email уже существует. Введите пароль.', 'info');
            } else {
                showNotification(mapErrorMessage(error), 'error');
            }
        }
    }
    
    // ===== ОБРАБОТКА ШАГА 3 (КОД ПОДТВЕРЖДЕНИЯ) =====
    async function handleStep3() {
        const authCode = document.getElementById('authCode');
        const code = authCode ? authCode.value.trim() : '';
        
        if (!code) {
            showNotification('Введите код из письма', 'error');
            return;
        }
        
        try {
            await window.authAPI.confirm(window.registrationEmail, parseInt(code));
            showNotification('Успешная авторизация!', 'success');
            await createBooking();
        } catch (error) {
            console.error('Confirmation error:', error);
            showNotification(mapErrorMessage(error), 'error');
        }
    }
    
    // ===== ОСНОВНАЯ ЛОГИКА SUBMIT =====
    async function submitBooking() {
        if (!selectedDate || !selectedTime) {
            showNotification('Выберите дату и время', 'error');
            return;
        }
        
        const activeStep = document.querySelector('.confirmation-step.active');
        if (!activeStep) {
            console.error('No active step found');
            return;
        }
        
        switch (activeStep.id) {
            case 'step1':
                const authorized = await checkAuth();
                if (authorized) {
                    await createBooking();
                } else {
                    showStep(2);
                    // Сбрасываем состояние шага 2
                    const passwordGroup = document.getElementById('passwordGroup');
                    const firstnameGroup = document.getElementById('firstnameGroup');
                    passwordGroup.style.display = 'none';
                    firstnameGroup.style.display = 'none';
                    document.getElementById('step2Title').textContent = 'Введите почту';
                    document.getElementById('emailSubmitBtn').textContent = 'Продолжить';
                }
                break;
                
            case 'step2':
                // Показываем поле имени при входе в шаг 2
                document.getElementById('firstnameGroup').style.display = 'flex';
                await handleStep2();
                break;
                
            case 'step3':
                await handleStep3();
                break;
                
            case 'step4':
                await handleStep4();
                break;
                
            default:
                console.warn('Unknown step:', activeStep.id);
        }
    }
    
    // ===== ОБРАБОТКА ШАГА 4 (регистрация с кодом) =====
    async function handleStep4() {
        const firstnameInput = document.getElementById('authFirstname');
        const authCodeNew = document.getElementById('authCodeNew');
        const firstname = firstnameInput ? firstnameInput.value.trim() : '';
        const code = authCodeNew ? authCodeNew.value.trim() : '';
        
        if (!firstname || firstname.length < 3) {
            showNotification('Имя должно содержать минимум 3 символа', 'error');
            return;
        }
        
        if (!code) {
            showNotification('Введите код из письма', 'error');
            return;
        }
        
        try {
            await window.authAPI.prepare(firstname, '', window.registrationEmail, '', '');
            await window.authAPI.confirm(window.registrationEmail, parseInt(code));
            showNotification('Регистрация успешна!', 'success');
            await createBooking();
        } catch (error) {
            console.error('Registration error:', error);
            showNotification(mapErrorMessage(error), 'error');
        }
    }
    
    function isValidEmail(email) {
        return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
    }
    
    function showNotification(message, type = 'info') {
        const existing = document.querySelector('.appointment-toast');
        if (existing) existing.remove();
        const toast = document.createElement('div');
        toast.className = `appointment-toast ${type}`;
        toast.textContent = message;
        document.body.appendChild(toast);
        requestAnimationFrame(() => toast.classList.add('show'));
        setTimeout(() => {
            toast.classList.remove('show');
            setTimeout(() => toast.remove(), 300);
        }, 3000);
    }

    // ===== МАППИНГ ОШИБОК БЭКЕНДА =====
    function mapErrorMessage(error) {
        if (!error) return 'Неизвестная ошибка';
        
        // Если ошибка уже содержит понятное сообщение
        if (error.message && !error.message.includes('unknown error')) {
            const msg = error.message.toLowerCase();
            if (msg.includes('too much bookings') || msg.includes('too many bookings')) {
                return 'Вы не можете создать еще одну запись.';
            }
            if (msg.includes('already booked')) return 'Это время уже занято.';
            if (msg.includes('not found')) return 'Услуга не найдена.';
            if (msg.includes('invalid')) return 'Некорректные данные.';
            if (msg.includes('unauthorized')) return 'Необходимо авторизоваться.';
            if (msg.includes('forbidden')) return 'Нет доступа.';
            if (msg.includes('conflict')) return 'Конфликт данных.';
            if (msg.includes('too many requests')) return 'Слишком много запросов.';
            return error.message;
        }
        
        if (error.status === 400) return 'Некорректный запрос.';
        if (error.status === 401) return 'Необходимо авторизоваться.';
        if (error.status === 403) return 'Нет доступа.';
        if (error.status === 404) return 'Ресурс не найден.';
        if (error.status === 409) return 'Конфликт данных.';
        if (error.status === 422) return 'Ошибка валидации.';
        if (error.status === 429) return 'Слишком много запросов.';
        if (error.status >= 500) return 'Ошибка сервера.';
        
        return 'Произошла ошибка.';
    }
    
    function setupEvents() {
        if (!confirmationCloseBtn || !confirmationOverlay || !confirmationSubmitBtn) return;
        
        confirmationCloseBtn.addEventListener('click', closeConfirmation);
        confirmationOverlay.addEventListener('mousedown', (e) => {
            if (e.target === confirmationOverlay) closeConfirmation();
        });
        confirmationSubmitBtn.addEventListener('click', submitBooking);
        
        const emailSubmitBtn = document.getElementById('emailSubmitBtn');
        if (emailSubmitBtn) {
            emailSubmitBtn.addEventListener('click', submitBooking);
        }
        
        const codeSubmitBtn = document.getElementById('codeSubmitBtn');
        if (codeSubmitBtn) {
            codeSubmitBtn.addEventListener('click', submitBooking);
        }
        
        const codeSubmitBtnNew = document.getElementById('codeSubmitBtnNew');
        if (codeSubmitBtnNew) {
            codeSubmitBtnNew.addEventListener('click', submitBooking);
        }
        
        const connectTgBtn = document.getElementById('connectTgBtn');
        if (connectTgBtn) {
            connectTgBtn.addEventListener('click', () => {
                showNotification('Функция подключения Telegram будет доступна позже', 'info');
            });
        }
        
        const timeSections = document.querySelector('.time-sections');
        if (timeSections) {
            timeSections.addEventListener('click', (e) => {
                if (e.target.classList.contains('time-slot')) openConfirmation();
            });
        }
    }
    
    async function init() {
        initDOM();
        
        try {
            currentOffer = await window.offerAPI.getById(OFFER_ID);
            
            document.querySelector('.service-title').textContent = currentOffer.title;
            document.getElementById('serviceDescription').textContent = currentOffer.description || '';
            document.querySelector('.service-amount').textContent = (currentOffer.price_penny ?? currentOffer.price) / 100;
            
            if (MODE === 'edit' && BOOKING_ID) {
                existingBooking = await window.bookingAPI.getById(BOOKING_ID);
                const bookingDate = new Date(existingBooking.start_time * 1000);
                selectedDate = bookingDate.toISOString().split('T')[0];
                selectedTime = bookingDate.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
                
                // Предзаполняем комментарий
                const groupInput = document.getElementById('groupInput');
                if (groupInput) {
                    groupInput.value = existingBooking.comment || '';
                }
            }
            
            await generateCalendar();
            setupCalendarScrollBtns();
            setupEvents();
            
            // Если есть выбранная дата (например, при редактировании), загружаем слоты
            if (selectedDate) {
                await loadTimeSlots(selectedDate);
            }
            
        } catch (error) {
            console.error('Initialization error:', error);
            showNotification('Не удалось загрузить данные услуги', 'error');
        }
    }
    
    // Запуск
    init();
})();
