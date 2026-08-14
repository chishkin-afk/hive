// Глобальное состояние
let currentPage = 1;
const limit = 10;
let totalPages = 1;
let currentOffer = null;
let authorId = null;

// Элементы DOM
const servicesList = document.getElementById('servicesList');
const pagination = document.getElementById('pagination');
const backBtn = document.getElementById('backBtn');
const forwardBtn = document.getElementById('forwardBtn');
const currentPageSpan = document.getElementById('currentPage');
const totalPagesSpan = document.getElementById('totalPages');
const createServiceBtn = document.getElementById('createServiceBtn');
const serviceModalOverlay = document.getElementById('serviceModalOverlay');
const serviceCreateModalOverlay = document.getElementById('serviceCreateModalOverlay');

// Поля редактирования услуги
const editServiceTitle = document.getElementById('editServiceTitle');
const editServiceDescription = document.getElementById('editServiceDescription');
const editServicePrice = document.getElementById('editServicePrice');
const editServiceDuration = document.getElementById('editServiceDuration');
const editServiceInterval = document.getElementById('editServiceInterval');
const saveEditServiceBtn = document.getElementById('saveEditServiceBtn');
const suspendServiceBtn = document.getElementById('suspendServiceBtn');
const resumeServiceBtn = document.getElementById('resumeServiceBtn');
const copyURLServiceBtn = document.getElementById('copyURLServiceBtn');
const deleteServiceBtn = document.getElementById('deleteServiceBtn');
const editBookingsList = document.getElementById('editBookingsList');

// Поля создания услуги
const createServiceTitle = document.getElementById('createServiceTitle');
const createServiceDescription = document.getElementById('createServiceDescription');
const createServicePrice = document.getElementById('createServicePrice');
const createServiceDuration = document.getElementById('createServiceDuration');
const createServiceInterval = document.getElementById('createServiceInterval');

// Получение ID авторизованного пользователя
async function initAuth() {
    try {
        const user = await authAPI.self();
        authorId = user.id;
        return true;
    } catch (error) {
        console.error('Ошибка получения данных пользователя:', error);
        servicesList.innerHTML = '<div class="service-card">Ошибка авторизации. Пожалуйста, войдите в аккаунт.</div>';
        return false;
    }
}

// Загрузка списка офферов с скелетоном
async function loadOffers(page) {
    if (!authorId) {
        const authorized = await initAuth();
        if (!authorized) return;
    }
    
    // Показываем скелетон при загрузке
    showSkeleton();
    
    try {
        const response = await offerAPI.list(authorId, page, limit);
        
        totalPages = response.total_pages || 0;
        currentPage = page;
        
        // Если список пустой и это не первая страница, возвращаемся на первую
        if (response.list.length === 0 && currentPage > 1) {
            currentPage = 1;
            totalPages = 0;
        }
        
        renderOffers(response.list);
        updatePagination();
    } catch (error) {
        console.error('Ошибка загрузки офферов:', error);
        servicesList.innerHTML = '<div class="service-card">Ошибка загрузки услуг</div>';
        currentPage = 1;
        totalPages = 0;
        updatePagination();
    }
}

// Показ скелетона загрузки
function showSkeleton() {
    servicesList.innerHTML = `
        <div class="service-card loading">
            <div class="skeleton skeleton-title"></div>
            <div class="skeleton skeleton-description"></div>
            <div class="skeleton skeleton-meta"></div>
        </div>

    `;
}

// Рендеринг карточек офферов
function renderOffers(offers) {
    if (offers.length === 0) {
        servicesList.innerHTML = '<div class="service-card">Услуги не найдены</div>';
        return;
    }
    
    servicesList.innerHTML = offers.map(offer => {
        // Бэкенд возвращает price в penny, конвертируем в рубли
        let priceInRubles = 0;
        const priceValue = offer.price_penny ?? offer.price; // пробуем оба поля
        if (priceValue !== null && priceValue !== undefined && typeof priceValue === 'number') {
            priceInRubles = priceValue / 100;
        }
        
        return `
        <div class="service-card" data-offer-id="${offer.id}">
            <h3 class="service-title">${escapeHtml(offer.title)}</h3>
            <p class="service-description">${escapeHtml(offer.description || 'Без описания')}</p>
            <div class="service-meta">
                <div class="service-duration-wrapper">
                    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                        <path d="M12 22C17.5228 22 22 17.5228 22 12C22 6.47715 17.5228 2 12 2C6.47715 2 2 6.47715 2 12C2 17.5228 6.47715 22 12 22Z"
                            stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                        <path d="M12 6V12L16 14"
                            stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                    </svg>
                    <span class="service-duration">${offer.slot_duration_minutes} мин</span>
                </div>
                <div class="service-price">
                    <span class="service-amount">${priceInRubles}</span>
                    <span class="service-currency">₽</span>
                </div>
            </div>
        </div>
    `;
    }).join('');
    
    // Навешиваем обработчики кликов на карточки
    document.querySelectorAll('.service-card[data-offer-id]').forEach(card => {
        card.addEventListener('click', () => {
            const offerId = card.dataset.offerId;
            openOfferModal(offerId);
        });
    });
}

// Обновление отдельной карточки после редактирования
function updateOfferCard(offer) {
    const card = document.querySelector(`.service-card[data-offer-id="${offer.id}"]`);
    if (!card) return;
    
    const priceValue = offer.price_penny ?? offer.price ?? 0;
    const priceInRubles = typeof priceValue === 'number' ? priceValue / 100 : 0;
    
    card.querySelector('.service-title').textContent = offer.title;
    card.querySelector('.service-description').textContent = offer.description || 'Без описания';
    card.querySelector('.service-duration').textContent = `${offer.slot_duration_minutes} мин`;
    card.querySelector('.service-amount').textContent = priceInRubles;
}

// Обновление пагинации
function updatePagination() {
    // Если страниц нет вообще, показываем "— из —"
    if (totalPages === 0) {
        currentPageSpan.textContent = '—';
        totalPagesSpan.textContent = '—';
        backBtn.classList.add('inactive');
        forwardBtn.classList.add('inactive');
        return;
    }
    
    currentPageSpan.textContent = currentPage;
    totalPagesSpan.textContent = totalPages;
    
    backBtn.classList.toggle('inactive', currentPage === 1);
    forwardBtn.classList.toggle('inactive', currentPage >= totalPages);
}

// Открытие модального окна с данными оффера
async function openOfferModal(offerId) {
    try {
        const offer = await offerAPI.getById(offerId);
        currentOffer = offer;
        
        // Заполнение полей модального окна
        editServiceTitle.value = offer.title;
        editServiceDescription.value = offer.description || '';
        // Бэкенд возвращает price в penny, конвертируем в рубли
        const priceValue = offer.price_penny ?? offer.price; // пробуем оба поля
        let priceInRubles = 0;
        if (priceValue !== null && priceValue !== undefined && typeof priceValue === 'number') {
            priceInRubles = priceValue / 100;
        }
        editServicePrice.value = priceInRubles;
        editServiceDuration.value = offer.slot_duration_minutes;
        // Интервал: используем именно slot_interval_minutes, а не продолжительность
        editServiceInterval.value = offer.slot_interval_minutes ?? 0;
        
        // Показать/скрыть кнопки в зависимости от статуса
        if (offer.is_active) {
            suspendServiceBtn.style.display = 'block';
            resumeServiceBtn.style.display = 'none';
        } else {
            suspendServiceBtn.style.display = 'none';
            resumeServiceBtn.style.display = 'block';
        }
        
        // Загрузка записей на эту услугу
        await loadBookingsForOffer(offerId);
        
        // Открытие модального окна
        document.body.classList.add('block-scroll');
        serviceModalOverlay.classList.add('open');
        createServiceBtn.classList.add('hide');
        
        // Закрытие по клику вне модального окна
        serviceModalOverlay.addEventListener('mousedown', function handler(e) {
            if (e.target === this) {
                closeOfferModal();
                serviceModalOverlay.removeEventListener('mousedown', handler);
            }
        });
    } catch (error) {
        console.error('Ошибка загрузки оффера:', error);
        showNotification('Не удалось загрузить данные услуги', 'error');
    }
}

// Загрузка записей для услуги со скелетоном
async function loadBookingsForOffer(offerId) {
    // Показываем скелетон
    editBookingsList.innerHTML = `
        <div class="booking-card loading">
            <div class="skeleton" style="width: 60%; height: 16px;"></div>
            <div class="skeleton" style="width: 40%; height: 14px;"></div>
            <div class="skeleton" style="width: 80%; height: 14px;"></div>
        </div>
    `;
    
    try {
        // TODO: реализовать API для загрузки записей
        // const bookings = await bookingAPI.listByOffer(offerId);
        setTimeout(() => {
            editBookingsList.innerHTML = '<div class="booking-card">Пока что записей нет</div>';
        }, 500);
    } catch (error) {
        editBookingsList.innerHTML = '<div class="booking-card">Ошибка загрузки записей</div>';
    }
}

// Закрытие модального окна
function closeOfferModal() {
    document.body.classList.remove('block-scroll');
    serviceModalOverlay.classList.remove('open');
    createServiceBtn.classList.remove('hide');
    currentOffer = null;
    
    // Сбрасываем подсветку полей
    [editServiceTitle, editServiceDescription, editServicePrice, editServiceDuration, editServiceInterval]
        .forEach(el => el?.classList.remove('input-error', 'input-success'));
}

// Экранирование HTML
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Обработчики пагинации
backBtn.addEventListener('click', () => {
    if (currentPage > 1) {
        currentPage--;
        loadOffers(currentPage);
    }
});

forwardBtn.addEventListener('click', () => {
    if (currentPage < totalPages) {
        currentPage++;
        loadOffers(currentPage);
    }
});

// Сохранение изменений в услуге - отправляем только изменённые поля с валидацией
saveEditServiceBtn.addEventListener('click', async () => {
    if (!currentOffer) return;
    
    // Сброс ошибок валидации
    clearFieldErrors();
    
    // Валидация полей - по одному, как при создании
    const newTitle = editServiceTitle.value.trim();
    if (!newTitle || newTitle.length <= 3 || newTitle.length >= 128) {
        editServiceTitle.classList.add('input-error');
        editServiceTitle.focus();
        showNotification('Название должно быть больше 3 и меньше 128 символов', 'error');
        return;
    }
    
    const newDescription = editServiceDescription.value.trim();
    if (newDescription && newDescription.length <= 3) {
        editServiceDescription.classList.add('input-error');
        editServiceDescription.focus();
        showNotification('Описание должно быть больше 3 символов или пустым', 'error');
        return;
    }
    
    const newPrice = parseFloat(editServicePrice.value);
    if (isNaN(newPrice) || newPrice < 0) {
        editServicePrice.classList.add('input-error');
        editServicePrice.focus();
        showNotification('Укажите корректную цену (минимум 0)', 'error');
        return;
    }
    
    const newDuration = parseInt(editServiceDuration.value, 10);
    if (isNaN(newDuration) || newDuration <= 0) {
        editServiceDuration.classList.add('input-error');
        editServiceDuration.focus();
        showNotification('Укажите корректную продолжительность больше 0', 'error');
        return;
    }
    
    let newInterval = parseInt(editServiceInterval.value, 10);
    if (isNaN(newInterval) || editServiceInterval.value.trim() === '') {
        newInterval = 0;
    } else if (newInterval < 0) {
        editServiceInterval.classList.add('input-error');
        editServiceInterval.focus();
        showNotification('Интервал не может быть отрицательным', 'error');
        return;
    }
    
    try {
        const updates = { offerId: currentOffer.id };
        let hasChanges = false;
        
        if (newTitle !== currentOffer.title) {
            updates.title = newTitle;
            hasChanges = true;
        }
        
        // Нормализуем описание: пустая строка = null
        const newDescNormalized = newDescription ? newDescription : null;
        const currentDescNormalized = currentOffer.description || null;
        
        // Сравниваем и отправляем, если изменилось (включая случай когда очистили описание)
        if (newDescNormalized !== currentDescNormalized) {
            updates.description = newDescNormalized;
            hasChanges = true;
        }
        
        const newPricePenny = Math.floor(newPrice * 100);
        const currentPriceValue = currentOffer.price_penny ?? currentOffer.price ?? 0;
        if (newPricePenny !== currentPriceValue) {
            updates.pricePenny = newPricePenny;
            hasChanges = true;
        }
        
        if (newDuration !== currentOffer.slot_duration_minutes) {
            updates.slotDurationMinutes = newDuration;
            hasChanges = true;
        }
        
        const currentInterval = currentOffer.slot_interval_minutes ?? currentOffer.slot_duration_minutes ?? 0;
        // Сравниваем как числа, приводя к одному типу
        if (Number(newInterval) !== Number(currentInterval)) {
            updates.slotIntervalMinutes = newInterval;
            hasChanges = true;
        }
        
        if (!hasChanges) {
            showNotification('Нет изменений для сохранения', 'info');
            return;
        }
        
        await offerAPI.update(updates);
        
        showNotification('Услуга обновлена', 'success');
        
        // Обновляем локально поля currentOffer вручную
        if (updates.title) currentOffer.title = updates.title;
        if (updates.description !== undefined) currentOffer.description = updates.description;
        if (updates.pricePenny !== undefined) currentOffer.price_penny = updates.pricePenny;
        if (updates.slotDurationMinutes !== undefined) currentOffer.slot_duration_minutes = updates.slotDurationMinutes;
        if (updates.slotIntervalMinutes !== undefined) currentOffer.slot_interval_minutes = updates.slotIntervalMinutes;
        
        // Обновляем карточку на странице
        updateOfferCard(currentOffer);
        
        // Убираем подсветку полей после успешного обновления
        [editServiceTitle, editServiceDescription, editServicePrice, editServiceDuration, editServiceInterval]
            .forEach(el => el.classList.remove('input-error', 'input-success'));
    } catch (error) {
        handleApiError(error, 'edit');
    }
});

// Приостановка услуги
suspendServiceBtn.addEventListener('click', async () => {
    if (!currentOffer) return;
    
    try {
        await offerAPI.update({
            offerId: currentOffer.id,
            isActive: false,
        });
        
        showNotification('Услуга приостановлена', 'success');
        // Меняем кнопки, не закрывая модалку
        suspendServiceBtn.style.display = 'none';
        resumeServiceBtn.style.display = 'block';
        currentOffer.is_active = false;
    } catch (error) {
        console.error('Ошибка приостановки услуги:', error);
        showNotification('Не удалось приостановить услугу', 'error');
    }
});

// Возобновление услуги
resumeServiceBtn.addEventListener('click', async () => {
    if (!currentOffer) return;
    
    try {
        await offerAPI.update({
            offerId: currentOffer.id,
            isActive: true,
        });
        
        showNotification('Услуга возобновлена', 'success');
        // Меняем кнопки, не закрывая модалку
        suspendServiceBtn.style.display = 'block';
        resumeServiceBtn.style.display = 'none';
        currentOffer.is_active = true;
    } catch (error) {
        console.error('Ошибка возобновления услуги:', error);
        showNotification('Не удалось возобновить услугу', 'error');
    }
});

// Копирование ссылки на услугу
copyURLServiceBtn.addEventListener('click', () => {
    if (!currentOffer) return;
    
    const url = `${window.location.origin}/appointment/${currentOffer.id}`;
    navigator.clipboard.writeText(url).then(() => {
        showNotification('Ссылка скопирована в буфер обмена', 'success');
    }).catch(() => {
        showNotification('Не удалось скопировать ссылку', 'error');
    });
});

// Удаление услуги
deleteServiceBtn.addEventListener('click', async () => {
    if (!currentOffer) return;
    
    if (!confirm('Вы уверены, что хотите удалить эту услугу? Это действие нельзя отменить.')) return;
    
    try {
        await offerAPI.delete(currentOffer.id);
        
        showNotification('Услуга удалена', 'success');
        closeOfferModal();
        loadOffers(currentPage);
    } catch (error) {
        console.error('Ошибка удаления услуги:', error);
        showNotification('Не удалось удалить услугу', 'error');
    }
});

// Маппинг ошибок бэкенда к полям формы
const FIELD_ERROR_MAP = [
    { field: 'createServiceTitle', patterns: ['title', 'название'], message: 'Название должно быть больше 3 и меньше 128 символов' },
    { field: 'createServiceDescription', patterns: ['description', 'описание'], message: 'Описание не должно превышать 512 символов' },
    { field: 'createServicePrice', patterns: ['price', 'цена'], message: 'Укажите корректную цену больше 0' },
    { field: 'createServiceDuration', patterns: ['duration', 'продолжительность'], message: 'Укажите корректную продолжительность больше 0' },
    { field: 'createServiceInterval', patterns: ['interval', 'интервал'], message: 'Укажите корректный интервал (минимум 0)' },
];

function clearFieldErrors() {
    [createServiceTitle, createServiceDescription, createServicePrice, createServiceDuration, createServiceInterval]
        .forEach(el => el?.classList.remove('input-error'));
}

function highlightFieldError(fieldName) {
    const mapping = FIELD_ERROR_MAP.find(m => m.field === fieldName);
    if (mapping) {
        const element = document.getElementById(fieldName);
        if (element) {
            element.classList.add('input-error');
            element.focus();
        }
    }
}

function handleApiError(error, context = 'create') {
    const msg = (error.message || '').toLowerCase();
    const status = error.status;
    
    // Пробуем найти подходящее поле по сообщению об ошибке
    for (const rule of FIELD_ERROR_MAP) {
        if (rule.patterns.some(p => msg.includes(p))) {
            highlightFieldError(rule.field);
            showNotification(rule.message, 'error');
            return;
        }
    }
    
    // Если не нашли специфичное поле, показываем общую ошибку
    if (status === 400) {
        showNotification('Проверьте корректность заполненных данных', 'error');
    } else if (status === 401) {
        showNotification('Требуется авторизация', 'error');
    } else if (status >= 500) {
        showNotification('Ошибка сервера. Попробуйте позже', 'error');
    } else {
        showNotification(error.message || 'Произошла ошибка', 'error');
    }
}

// Валидация при вводе - убираем ошибку при корректном вводе
function setupLiveValidation() {
    const validators = [
        { 
            element: createServiceTitle, 
            test: (v) => v.trim().length > 3 && v.trim().length < 128,
            emptyIsOk: false 
        },
        { 
            element: createServiceDescription, 
            test: (v) => {
                const trimmed = v.trim();
                // Либо пусто, либо больше 3 символов
                return !trimmed || trimmed.length > 3;
            },
            emptyIsOk: true 
        },
        { 
            element: createServicePrice, 
            test: (v) => {
                const num = parseFloat(v);
                return !isNaN(num) && num >= 0;
            },
            emptyIsOk: false 
        },
        { 
            element: createServiceDuration, 
            test: (v) => {
                const num = parseInt(v, 10);
                return !isNaN(num) && num > 0;
            },
            emptyIsOk: false 
        },
        { 
            element: createServiceInterval, 
            test: (v) => {
                if (!v.trim()) return true; // пустое поле ок (будет 0)
                const num = parseInt(v, 10);
                return !isNaN(num) && num >= 0;
            },
            emptyIsOk: true 
        },
        { 
            element: editServiceTitle, 
            test: (v) => v.trim().length > 3 && v.trim().length < 128,
            emptyIsOk: false 
        },
        { 
            element: editServiceDescription, 
            test: (v) => {
                const trimmed = v.trim();
                // Либо пусто, либо больше 3 символов
                return !trimmed || trimmed.length > 3;
            },
            emptyIsOk: true 
        },
        { 
            element: editServicePrice, 
            test: (v) => {
                const num = parseFloat(v);
                return !isNaN(num) && num >= 0;
            },
            emptyIsOk: false 
        },
        { 
            element: editServiceDuration, 
            test: (v) => {
                const num = parseInt(v, 10);
                return !isNaN(num) && num > 0;
            },
            emptyIsOk: false 
        },
        { 
            element: editServiceInterval, 
            test: (v) => {
                if (!v.trim()) return true;
                const num = parseInt(v, 10);
                return !isNaN(num) && num >= 0;
            },
            emptyIsOk: true 
        },
    ];

    validators.forEach(({ element, test, emptyIsOk }) => {
        if (!element) return;
        element.addEventListener('input', function () {
            const val = this.value;
            if (!val && emptyIsOk) {
                this.classList.remove('input-error', 'input-success');
                return;
            }
            const ok = val && test(val);
            this.classList.toggle('input-error', !ok);
            this.classList.toggle('input-success', ok);
        });
    });
}

// Создание новой услуги с валидацией
const saveCreateServiceBtn = document.getElementById('saveCreateServiceBtn');
saveCreateServiceBtn?.addEventListener('click', async () => {
    // Сброс ошибок валидации
    clearFieldErrors();
    
    // Валидация полей - проверяем по одному, останавливаемся на первой ошибке
    const title = createServiceTitle.value.trim();
    if (!title || title.length <= 3 || title.length >= 128) {
        createServiceTitle.classList.add('input-error');
        createServiceTitle.focus();
        showNotification('Название должно быть больше 3 и меньше 128 символов', 'error');
        return;
    }
    
    const description = createServiceDescription.value.trim();
    if (description && description.length <= 3) {
        createServiceDescription.classList.add('input-error');
        createServiceDescription.focus();
        showNotification('Описание должно быть больше 3 символов или пустым', 'error');
        return;
    }
    
    const price = parseFloat(createServicePrice.value);
    if (isNaN(price) || price < 0) {
        createServicePrice.classList.add('input-error');
        createServicePrice.focus();
        showNotification('Укажите корректную цену (минимум 0)', 'error');
        return;
    }
    
    const duration = parseInt(createServiceDuration.value, 10);
    if (isNaN(duration) || duration <= 0) {
        createServiceDuration.classList.add('input-error');
        createServiceDuration.focus();
        showNotification('Укажите корректную продолжительность больше 0', 'error');
        return;
    }
    
    // Интервал: если не заполнен - ставим 0, иначе проверяем на отрицательное число
    let interval = parseInt(createServiceInterval.value, 10);
    if (isNaN(interval) || createServiceInterval.value.trim() === '') {
        interval = 0; // По умолчанию 0, если не указано
    } else if (interval < 0) {
        createServiceInterval.classList.add('input-error');
        createServiceInterval.focus();
        showNotification('Интервал не может быть отрицательным', 'error');
        return;
    }
    
    try {
        const pricePenny = Math.floor(price * 100);
        
        await offerAPI.create({
            title: title,
            description: description || null,
            pricePenny: pricePenny,
            slotDurationMinutes: duration,
            slotIntervalMinutes: interval,
        });
        
        // Закрытие модального окна
        document.body.classList.remove('block-scroll');
        serviceCreateModalOverlay.classList.remove('open');
        createServiceBtn.classList.remove('hide');
        
        // Очистка полей
        createServiceTitle.value = '';
        createServiceDescription.value = '';
        createServicePrice.value = '';
        createServiceDuration.value = '';
        createServiceInterval.value = '';
        
        // Перезагрузка списка
        loadOffers(1);
    } catch (error) {
        handleApiError(error, 'create');
    }
});

createServiceBtn.addEventListener('click', () => {
    document.body.classList.add('block-scroll');
    serviceCreateModalOverlay.classList.add('open');
    createServiceBtn.classList.add('hide');

    serviceCreateModalOverlay.addEventListener('mousedown', function handler(e) {
        if (e.target === this) {
            document.body.classList.remove('block-scroll');
            serviceCreateModalOverlay.classList.remove('open');
            createServiceBtn.classList.remove('hide');
            serviceCreateModalOverlay.removeEventListener('mousedown', handler);
        }
    });
});

// Уведомления
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
    }, 3000);
}

// Инициализация при загрузке страницы
setupLiveValidation();
loadOffers(currentPage);
