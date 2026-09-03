let currentPage = 1;
const limit = 10;
let totalPages = 1;
let currentOffer = null;
let authorId = null;

let bookingsCurrentPage = 1;
let bookingsTotalPages = 1;
let bookingsOfferId = null;
const bookingsLimit = 10;

const servicesList = document.getElementById('servicesList');
const pagination = document.getElementById('pagination');
const backBtn = document.getElementById('backBtn');
const forwardBtn = document.getElementById('forwardBtn');
const currentPageSpan = document.getElementById('currentPage');
const totalPagesSpan = document.getElementById('totalPages');
const createServiceBtn = document.getElementById('createServiceBtn');
const serviceModalOverlay = document.getElementById('serviceModalOverlay');
const serviceCreateModalOverlay = document.getElementById('serviceCreateModalOverlay');

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

const createServiceTitle = document.getElementById('createServiceTitle');
const createServiceDescription = document.getElementById('createServiceDescription');
const createServicePrice = document.getElementById('createServicePrice');
const createServiceDuration = document.getElementById('createServiceDuration');
const createServiceInterval = document.getElementById('createServiceInterval');

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

async function loadOffers(page) {
    if (!authorId) {
        const authorized = await initAuth();
        if (!authorized) return;
    }
    
    try {
        const response = await offerAPI.list(authorId, page, limit);
        
        totalPages = response.total_pages || 0;
        currentPage = page;
        
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

function renderOffers(offers) {
    const empty = document.getElementById('servicesEmpty');
    if (offers.length === 0) {
        servicesList.innerHTML = '';
        if (empty) empty.style.display = 'flex';
        return;
    }
    if (empty) empty.style.display = 'none';
    
    servicesList.innerHTML = offers.map(offer => {
        let priceInRubles = 0;
        const priceValue = offer.price_penny ?? offer.price;
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
    
    document.querySelectorAll('.service-card[data-offer-id]').forEach(card => {
        card.addEventListener('click', () => {
            const offerId = card.dataset.offerId;
            openOfferModal(offerId);
        });
    });
}

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

function updatePagination() {
    const paginationEl = document.getElementById('pagination');
    if (!paginationEl) return;
    
    if (totalPages === 0 || totalPages === 1) {
        paginationEl.style.display = 'none';
        return;
    }
    
    paginationEl.style.display = 'flex';
    currentPageSpan.textContent = currentPage;
    totalPagesSpan.textContent = totalPages;
    
    backBtn.classList.toggle('inactive', currentPage === 1);
    forwardBtn.classList.toggle('inactive', currentPage >= totalPages);
}

async function openOfferModal(offerId) {
    try {
        const offer = await offerAPI.getById(offerId);
        currentOffer = offer;
        
        editServiceTitle.value = offer.title;
        editServiceDescription.value = offer.description || '';
        const priceValue = offer.price_penny ?? offer.price;
        let priceInRubles = 0;
        if (priceValue !== null && priceValue !== undefined && typeof priceValue === 'number') {
            priceInRubles = priceValue / 100;
        }
        editServicePrice.value = priceInRubles;
        editServiceDuration.value = offer.slot_duration_minutes;
        editServiceInterval.value = offer.slot_interval_minutes ?? 0;
        
        if (offer.is_active) {
            suspendServiceBtn.style.display = 'block';
            resumeServiceBtn.style.display = 'none';
        } else {
            suspendServiceBtn.style.display = 'none';
            resumeServiceBtn.style.display = 'block';
        }
        
        await loadBookingsForOffer(offerId);
        
        document.body.classList.add('block-scroll');
        serviceModalOverlay.classList.add('open');
        createServiceBtn.classList.add('hide');
        
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

async function loadBookingsForOffer(offerId, page = 1) {
    bookingsOfferId = offerId;
    bookingsCurrentPage = page;
    
    editBookingsList.innerHTML = `
        <div class="booking-card loading">
            <div class="skeleton" style="width: 60%; height: 16px;"></div>
            <div class="skeleton" style="width: 40%; height: 14px;"></div>
            <div class="skeleton" style="width: 80%; height: 14px;"></div>
        </div>
    `;
    
    const bookingsPagination = document.getElementById('bookingsPagination');
    if (bookingsPagination) bookingsPagination.style.display = 'none';
    
    try {
        const response = await bookingAPI.list(page, bookingsLimit, offerId);
        let bookings = [];
        
        if (response && response.list) {
            bookings = response.list;
            bookingsTotalPages = response.total_pages || 1;
        } else if (Array.isArray(response)) {
            bookings = response.map(item => ({ booking: item }));
            bookingsTotalPages = 1;
        }
        
        if (bookings.length === 0) {
            editBookingsList.innerHTML = '<div class="booking-card">Нет записей</div>';
            if (bookingsPagination) bookingsPagination.style.display = 'none';
            return;
        }
        
        editBookingsList.innerHTML = bookings.map(item => {
            const booking = item.booking || item;
            const user = item.user || null;
            
            if (!booking || !booking.start_time) {
                return '';
            }
            
            const date = new Date(booking.start_time * 1000);
            const dateStr = date.toLocaleDateString('ru-RU', { 
                day: 'numeric', 
                month: 'long', 
                year: 'numeric' 
            });
            const timeStr = date.toLocaleTimeString('ru-RU', { 
                hour: '2-digit', 
                minute: '2-digit' 
            });
            const isCancelled = booking.is_cancelled || false;
            const userName = user && user.firstname ? escapeHtml(user.firstname) : 'Неизвестный пользователь';
            
            return `
                <div class="booking-card" data-booking-id="${booking.id}">
                    <div class="booking-card-main">
                        <span class="booking-date">${dateStr} в ${timeStr}</span>
                        <span class="booking-user">${userName}</span>
                        ${booking.comment ? `<span class="booking-comment">${escapeHtml(booking.comment)}</span>` : ''}
                        <span class="booking-status ${isCancelled ? 'cancelled' : 'active'}">
                            ${isCancelled ? 'Отменена' : 'Активна'}
                        </span>
                    </div>
                    ${!isCancelled ? `
                        <button class="booking-cancel-btn" data-booking-id="${booking.id}">
                            Отменить
                        </button>
                    ` : ''}
                </div>
            `;
        }).filter(html => html).join('');
        
        if (bookingsPagination) {
            bookingsPagination.style.display = 'flex';
            updateBookingsPagination();
        }
        
        document.querySelectorAll('.booking-cancel-btn').forEach(btn => {
            btn.addEventListener('click', async (e) => {
                e.stopPropagation();
                const bookingId = btn.dataset.bookingId;
                if (confirm('Вы уверены, что хотите отменить эту запись?')) {
                    try {
                        await bookingAPI.cancel(bookingId);
                        showNotification('Запись отменена', 'success');
                        await loadBookingsForOffer(offerId, bookingsCurrentPage);
                    } catch (error) {
                        console.error('Ошибка отмены записи:', error);
                        showNotification('Не удалось отменить запись', 'error');
                    }
                }
            });
        });
        
    } catch (error) {
        console.error('Ошибка загрузки записей:', error);
        editBookingsList.innerHTML = '<div class="booking-card">Ошибка загрузки записей</div>';
        const bookingsPagination = document.getElementById('bookingsPagination');
        if (bookingsPagination) bookingsPagination.style.display = 'none';
    }
}

function updateBookingsPagination() {
    const bookingsPagination = document.getElementById('bookingsPagination');
    if (!bookingsPagination) return;
    
    const backBtn = document.getElementById('bookingsBackBtn');
    const forwardBtn = document.getElementById('bookingsForwardBtn');
    const currentPageSpan = document.getElementById('bookingsCurrentPage');
    const totalPagesSpan = document.getElementById('bookingsTotalPages');
    
    if (currentPageSpan) currentPageSpan.textContent = bookingsCurrentPage;
    if (totalPagesSpan) totalPagesSpan.textContent = bookingsTotalPages;
    
    if (backBtn) backBtn.disabled = bookingsCurrentPage <= 1;
    if (forwardBtn) forwardBtn.disabled = bookingsCurrentPage >= bookingsTotalPages;
}

function closeOfferModal() {
    document.body.classList.remove('block-scroll');
    serviceModalOverlay.classList.remove('open');
    createServiceBtn.classList.remove('hide');
    currentOffer = null;
    
    [editServiceTitle, editServiceDescription, editServicePrice, editServiceDuration, editServiceInterval]
        .forEach(el => el?.classList.remove('input-error', 'input-success'));
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

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

const bookingsBackBtn = document.getElementById('bookingsBackBtn');
const bookingsForwardBtn = document.getElementById('bookingsForwardBtn');

if (bookingsBackBtn) {
    bookingsBackBtn.addEventListener('click', () => {
        if (bookingsCurrentPage > 1 && bookingsOfferId) {
            bookingsCurrentPage--;
            loadBookingsForOffer(bookingsOfferId, bookingsCurrentPage);
        }
    });
}

if (bookingsForwardBtn) {
    bookingsForwardBtn.addEventListener('click', () => {
        if (bookingsCurrentPage < bookingsTotalPages && bookingsOfferId) {
            bookingsCurrentPage++;
            loadBookingsForOffer(bookingsOfferId, bookingsCurrentPage);
        }
    });
}

saveEditServiceBtn.addEventListener('click', async () => {
    if (!currentOffer) return;
    
    clearFieldErrors();
    
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
        
        const currentDescNormalized = currentOffer.description || '';
        const newDescNormalized = newDescription ? newDescription.trim() : '';
        
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
        
        if (updates.title) currentOffer.title = updates.title;
        if (updates.description !== undefined) currentOffer.description = updates.description;
        if (updates.pricePenny !== undefined) currentOffer.price_penny = updates.pricePenny;
        if (updates.slotDurationMinutes !== undefined) currentOffer.slot_duration_minutes = updates.slotDurationMinutes;
        if (updates.slotIntervalMinutes !== undefined) currentOffer.slot_interval_minutes = updates.slotIntervalMinutes;
        
        updateOfferCard(currentOffer);
        
        [editServiceTitle, editServiceDescription, editServicePrice, editServiceDuration, editServiceInterval]
            .forEach(el => el.classList.remove('input-error', 'input-success'));
    } catch (error) {
        handleApiError(error, 'edit');
    }
});

suspendServiceBtn.addEventListener('click', async () => {
    if (!currentOffer) return;
    
    try {
        await offerAPI.update({
            offerId: currentOffer.id,
            isActive: false,
        });
        
        showNotification('Услуга приостановлена', 'success');
        suspendServiceBtn.style.display = 'none';
        resumeServiceBtn.style.display = 'block';
        currentOffer.is_active = false;
    } catch (error) {
        console.error('Ошибка приостановки услуги:', error);
        showNotification('Не удалось приостановить услугу', 'error');
    }
});

resumeServiceBtn.addEventListener('click', async () => {
    if (!currentOffer) return;
    
    try {
        await offerAPI.update({
            offerId: currentOffer.id,
            isActive: true,
        });
        
        showNotification('Услуга возобновлена', 'success');
        suspendServiceBtn.style.display = 'block';
        resumeServiceBtn.style.display = 'none';
        currentOffer.is_active = true;
    } catch (error) {
        console.error('Ошибка возобновления услуги:', error);
        showNotification('Не удалось возобновить услугу', 'error');
    }
});

copyURLServiceBtn.addEventListener('click', () => {
    if (!currentOffer) return;
    
    const url = `${window.location.origin}/appointment/${currentOffer.id}`;
    navigator.clipboard.writeText(url).then(() => {
        showNotification('Ссылка скопирована в буфер обмена', 'success');
    }).catch(() => {
        showNotification('Не удалось скопировать ссылку', 'error');
    });
});

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
    
    for (const rule of FIELD_ERROR_MAP) {
        if (rule.patterns.some(p => msg.includes(p))) {
            highlightFieldError(rule.field);
            showNotification(rule.message, 'error');
            return;
        }
    }
    
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
                if (!v.trim()) return true;
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

const saveCreateServiceBtn = document.getElementById('saveCreateServiceBtn');
saveCreateServiceBtn?.addEventListener('click', async () => {
    clearFieldErrors();
    
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
    
    let interval = parseInt(createServiceInterval.value, 10);
    if (isNaN(interval) || createServiceInterval.value.trim() === '') {
        interval = 0;
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
        
        document.body.classList.remove('block-scroll');
        serviceCreateModalOverlay.classList.remove('open');
        createServiceBtn.classList.remove('hide');
        
        createServiceTitle.value = '';
        createServiceDescription.value = '';
        createServicePrice.value = '';
        createServiceDuration.value = '';
        createServiceInterval.value = '';
        
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

setupLiveValidation();
loadOffers(currentPage);
