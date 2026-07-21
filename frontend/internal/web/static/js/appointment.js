function setupCalendarScrollBtns() {
    const daysSelector = document.getElementById('daysSelector');
    const backBtn = document.getElementById('backBtn');
    const forwardBtn = document.getElementById('forwardBtn');

    function getScrollStep() {
        if (window.innerWidth < 768) {
            return 256;
        } else {
            return 512;
        }
    }

    function checkButtons() {
        const maxScroll = daysSelector.scrollWidth - daysSelector.clientWidth;

        backBtn.classList.toggle('inactive', daysSelector.scrollLeft <= 0);
        forwardBtn.classList.toggle('inactive', daysSelector.scrollLeft >= maxScroll-1);
    }

    backBtn.addEventListener('click', function() {
        daysSelector.scrollBy({ left: -getScrollStep(), behavior: 'smooth' });
    });

    forwardBtn.addEventListener('click', function() {
        daysSelector.scrollBy({ left: getScrollStep(), behavior: 'smooth' });
    });

    daysSelector.addEventListener('scroll', checkButtons);
    window.addEventListener('resize', checkButtons);
    checkButtons();
}
setupCalendarScrollBtns();

function setupCalendarCardClick() {
    const allDays = document.querySelectorAll('.day-card:not(.inactive)');
    
    allDays.forEach(function(day) {
        day.addEventListener('click', function() {
            document.querySelector('.day-card.active').classList.remove('active');
            this.classList.add('active');
        });
    });
}
setupCalendarCardClick();

function setupConfirmation() {
    const timeSlots = document.querySelectorAll('.time-slot');
    const overlay = document.getElementById('confirmationOverlay');
    const closeBtn = document.getElementById('confirmationCloseBtn');
    const dates = document.querySelectorAll('.offer-date');
    const times = document.querySelectorAll('.offer-time');
    const activeDay = document.querySelector('.day-card.active');

    timeSlots.forEach(function(slot) {
        slot.addEventListener('click', function() {
            const activeDay = document.querySelector('.day-card.active');

            const date = activeDay.dataset.date;
            const time = this.dataset.time;

            const formatted = new Date(date).toLocaleDateString('ru', {
                weekday: 'short',
                day: 'numeric',
                month: 'long',
                year: 'numeric'
            });

            dates.forEach(el => {
                el.textContent = formatted;
            });
            times.forEach(el => {
                el.textContent = time;
            });

            document.body.classList.add('block-scroll');
            overlay.classList.add('open');
        });
    });

    closeBtn.addEventListener('click', function() {
        document.body.classList.remove('block-scroll');
        overlay.classList.remove('open');
    });

    overlay.addEventListener('mousedown', function(e) {
        if (e.target === this) {
            document.body.classList.remove('block-scroll');
            overlay.classList.remove('open');
        }
    });
}
setupConfirmation();

function showStep(stepNumber) {
    document.querySelectorAll('.confirmation-step').forEach(el => {
        el.classList.remove('active');
    });
    
    const target = document.getElementById(`step${stepNumber}`);
    if (target) {
        target.classList.add('active');
    }
}

let currentStep = 1;

function nextStep() {
    currentStep++;
    showStep(currentStep);
}

showStep(1);
