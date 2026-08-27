// Состояние пагинации
let currentPage = 1;
const limit = 10;
let totalPages = 1;
let bookingIdToCancel = null;

// DOM-элементы
const bookingsList = document.getElementById('bookingsList');
const bookingsEmpty = document.getElementById('bookingsEmpty');
const cancelOverlay = document.getElementById('cancelOverlay');
const cancelClose = document.getElementById('cancelClose');
const cancelConfirm = document.getElementById('cancelConfirm');
const cancelText = document.getElementById('cancelText');

// Экранирование HTML
function escapeHtml(text) {
    if (!text) return '';
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Показать скелетон
function showSkeleton() {
    bookingsList.innerHTML = `
        <div class="booking-card booking-card-skeleton">
            <div class="booking-date skeleton"></div>
            <div class="booking-main">
                <div class="booking-top">
                    <h2 class="booking-title skeleton-text"></h2>
                </div>
                <p class="booking-comment skeleton-text"></p>
            </div>
        </div>
        <div class="booking-card booking-card-skeleton">
            <div class="booking-date skeleton"></div>
            <div class="booking-main">
                <div class="booking-top">
                    <h2 class="booking-title skeleton-text"></h2>
                </div>
                <p class="booking-comment skeleton-text"></p>
            </div>
        </div>
        <div class="booking-card booking-card-skeleton">
            <div class="booking-date skeleton"></div>
            <div class="booking-main">
                <div class="booking-top">
                    <h2 class="booking-title skeleton-text"></h2>
                </div>
                <p class="booking-comment skeleton-text"></p>
            </div>
        </div>
    `;
    bookingsEmpty.style.display = 'none';
}

// Загрузка записей
async function loadBookings(page = 1) {
    currentPage = page;
    showSkeleton();

    try {
        const response = await window.bookingAPI.list(page, limit);
        
        if (!response || !response.list || response.list.length === 0) {
            bookingsList.innerHTML = '';
            bookingsEmpty.style.display = 'flex';
            const pagination = document.getElementById('bookingsPagination');
            if (pagination) pagination.style.display = 'none';
            return;
        }

        totalPages = response.total_pages || 1;

        // Собираем уникальные offer_id
        const offerIds = [...new Set(response.list.map(item => item.booking.offer_id).filter(id => id))];

        // Загружаем данные по всем услугам параллельно
        const offersMap = new Map();
        if (offerIds.length > 0 && window.offerAPI) {
            await Promise.all(offerIds.map(async (offerId) => {
                try {
                    const offer = await window.offerAPI.getById(offerId);
                    if (offer) {
                        offersMap.set(offerId, offer);
                    }
                } catch (err) {
                    console.warn(`Не удалось загрузить услугу ${offerId}:`, err);
                }
            }));
        }

        // Рендерим карточки
        bookingsList.innerHTML = response.list.map(item => {
            const booking = item.booking;
            const user = item.user;
            
            if (!booking || !booking.start_time) return '';
            
            const date = new Date(booking.start_time * 1000);
            const mskOffset = 3 * 60; // MSK = UTC+3
            const mskTime = new Date(date.getTime() + (date.getTimezoneOffset() + mskOffset) * 60000);
            const dayNames = ['ВС', 'ПН', 'ВТ', 'СР', 'ЧТ', 'ПТ', 'СБ'];
            const monthNames = ['ЯНВ', 'ФЕВ', 'МАР', 'АПР', 'МАЙ', 'ИЮН', 'ИЮЛ', 'АВГ', 'СЕН', 'ОКТ', 'НОЯ', 'ДЕК'];
            const dayName = dayNames[mskTime.getDay()];
            const monthName = monthNames[mskTime.getMonth()];
            const dayNumber = String(mskTime.getDate()).padStart(2, '0');
            const monthNumber = String(mskTime.getMonth() + 1).padStart(2, '0');
            const dateStr = `${dayNumber}.${monthNumber}`;
            const timeStr = mskTime.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
            
            const comment = booking.comment ? escapeHtml(booking.comment) : '';
            const isCancelled = booking.is_cancelled || false;
            
            const offer = offersMap.get(booking.offer_id);
            const offerTitle = offer ? escapeHtml(offer.title) : 'Услуга';
            const offerPrice = offer ? `${(offer.price / 100).toFixed(2)} ₽` : '';
            
            // Проверяем, можно ли отменить/перенести запись
            let canModify = false;
            if (!isCancelled) {
                const now = Math.floor(Date.now() / 1000);
                // Если есть duration, проверяем end_time
                if (offer && offer.duration) {
                    const endTime = booking.start_time + offer.duration * 60; // duration в минутах -> секунды
                    if (booking.start_time > now) {
                        canModify = true;
                    }
                } else {
                    // Если duration нет, просто проверяем, что запись ещё не началась
                    if (booking.start_time > now) {
                        canModify = true;
                    }
                }
            }
            
            return `
                <div class="booking-card" data-booking-id="${booking.id}">
                    <div class="booking-date">
                        <span class="booking-date-day">${dayName}</span>
                        <span class="booking-date-month">${dateStr}</span>
                        <span class="booking-date-time">${timeStr}</span>
                    </div>

                    <div class="booking-main">
                        <div class="booking-top">
                            <h2 class="booking-title">${offerTitle}</h2>
                            ${offerPrice ? `<span class="booking-price">${offerPrice}</span>` : ''}
                        </div>
                        ${comment ? `<p class="booking-comment">${comment}</p>` : ''}
                    </div>

                    <div class="booking-side">
                        <div class="booking-actions">
                            ${isCancelled ? `
                                <span class="booking-status cancelled">Отменена</span>
                            ` : canModify ? `
                                <a href="/appointment/${booking.offer_id}?booking_id=${booking.id}" class="booking-btn booking-btn-reschedule">Перенести</a>
                                <button type="button" class="booking-btn booking-btn-cancel danger" data-booking-id="${booking.id}">Отменить</button>
                            ` : `
                                <span class="booking-status completed">Завершена</span>
                            `}
                        </div>
                    </div>
                </div>
            `;
        }).filter(html => html).join('');

        // Обновляем пагинацию
        updatePagination();
        bookingsEmpty.style.display = 'none';

        // Навешиваем обработчики на кнопки отмены
        document.querySelectorAll('.booking-btn-cancel').forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.stopPropagation();
                const id = btn.dataset.bookingId;
                if (id) {
                    bookingIdToCancel = id;
                    const card = btn.closest('.booking-card');
                    const titleEl = card?.querySelector('.booking-title');
                    const name = titleEl ? titleEl.textContent : 'клиента';
                    cancelText.textContent = `Вы действительно хотите отменить запись на "${name}"?`;
                    cancelOverlay.classList.add('open');
                    cancelOverlay.removeAttribute('aria-hidden');
                    document.body.classList.add('block-scroll');
                    // Фокус на модалку
                    const firstFocusable = cancelOverlay.querySelector('button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])');
                    if (firstFocusable) {
                        setTimeout(() => firstFocusable.focus(), 100);
                    }
                }
            });
        });

    } catch (error) {
        console.error('Ошибка загрузки записей:', error);
        bookingsList.innerHTML = '<div class="booking-card" style="color: var(--color-error);">Ошибка загрузки записей</div>';
        bookingsEmpty.style.display = 'none';
    }
}

// Обновление пагинации
function updatePagination() {
    const backBtn = document.getElementById('bookingsBackBtn');
    const forwardBtn = document.getElementById('bookingsForwardBtn');
    const currentPageSpan = document.getElementById('bookingsCurrentPage');
    const totalPagesSpan = document.getElementById('bookingsTotalPages');
    const pagination = document.getElementById('bookingsPagination');
    
    if (!pagination) return;
    
    // Показываем пагинацию только если есть больше одной страницы
    if (totalPages <= 1) {
        pagination.style.display = 'none';
        return;
    }
    
    pagination.style.display = 'flex';
    if (currentPageSpan) currentPageSpan.textContent = currentPage;
    if (totalPagesSpan) totalPagesSpan.textContent = totalPages;
    if (backBtn) {
        backBtn.disabled = currentPage <= 1;
        backBtn.classList.toggle('inactive', currentPage <= 1);
    }
    if (forwardBtn) {
        forwardBtn.disabled = currentPage >= totalPages;
        forwardBtn.classList.toggle('inactive', currentPage >= totalPages);
    }
}

// Закрытие модалки отмены
function closeCancelModal() {
    cancelOverlay.classList.remove('open');
    cancelOverlay.setAttribute('aria-hidden', 'true');
    document.body.classList.remove('block-scroll');
    bookingIdToCancel = null;
}

// События модалки отмены
if (cancelClose) {
    cancelClose.addEventListener('click', closeCancelModal);
}
if (cancelOverlay) {
    cancelOverlay.addEventListener('mousedown', (e) => {
        if (e.target === cancelOverlay) closeCancelModal();
    });
}

// Подтверждение отмены
if (cancelConfirm) {
    cancelConfirm.addEventListener('click', async () => {
        if (!bookingIdToCancel) return;
        
        try {
            await window.bookingAPI.cancel(bookingIdToCancel);
            closeCancelModal();
            await loadBookings(currentPage);
        } catch (error) {
            console.error('Ошибка отмены записи:', error);
            alert('Не удалось отменить запись. Попробуйте позже.');
            closeCancelModal();
        }
    });
}

// Пагинация — кнопки
const backBtn = document.getElementById('bookingsBackBtn');
const forwardBtn = document.getElementById('bookingsForwardBtn');

if (backBtn) {
    backBtn.addEventListener('click', () => {
        if (currentPage > 1) {
            loadBookings(currentPage - 1);
        }
    });
}

if (forwardBtn) {
    forwardBtn.addEventListener('click', () => {
        if (currentPage < totalPages) {
            loadBookings(currentPage + 1);
        }
    });
}

// Загружаем записи при загрузке страницы
document.addEventListener('DOMContentLoaded', () => {
    loadBookings(1);
});
