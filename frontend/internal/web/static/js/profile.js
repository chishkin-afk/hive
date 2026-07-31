const services = document.querySelectorAll('.service-card');
services.forEach(el => {
    el.addEventListener('click', () => {
        const overlay = document.getElementById('serviceModalOverlay');

        document.body.classList.add('block-scroll');
        overlay.classList.add('open');


        overlay.addEventListener('mousedown', function(e) {
        if (e.target === this) {
            document.body.classList.remove('block-scroll');
            overlay.classList.remove('open');
        }
    });
    })
});

(function () {
    'use strict';

    const form = document.getElementById('profileForm');
    const navItems = document.querySelectorAll('.profile-nav-item');
    const sections = document.querySelectorAll('.profile-section');
    const messageEl = document.getElementById('profileFormMessage');
    const saveBtn = document.getElementById('profileSaveBtn');
    const bioTextarea = form ? form.querySelector('textarea[name="bio"]') : null;
    const bioCounter = document.getElementById('bioCounter');

    if (!form) return;

    const eyeOpenSVG = `<svg class="profile-eye-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M3 12C3 12 6.27273 5 12 5C17.7273 5 21 12 21 12C21 12 17.7273 19 12 19C6.27273 19 3 12 3 12Z" stroke="currentColor" stroke-width="2" stroke-linejoin="round"/>
        <path d="M12 15C13.6569 15 15 13.6569 15 12C15 10.3431 13.6569 9 12 9C10.3431 9 9 10.3431 9 12C9 13.6569 10.3431 15 12 15Z" stroke="currentColor" stroke-width="2"/>
    </svg>`;

    const eyeClosedSVG = `<svg class="profile-eye-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M16.86 17.1975C15.4614 18.3376 13.7584 18.9693 12 19C6.27273 19 3 12 3 12C4.01773 9.97167 5.28929 8.39703 7 7M10.2818 5.21002C10.845 5.06904 11.4216 4.99857 12 5.00002C17.7273 5.00002 21 12 21 12C20.5033 12.9937 19.911 13.9292 19.2327 14.7913M13.7345 13.855C13.5098 14.1129 13.2389 14.3198 12.9378 14.4632C12.6367 14.6067 12.3116 14.6839 11.9821 14.6901C11.6525 14.6963 11.3251 14.6315 11.0195 14.4994C10.7139 14.3674 10.4362 14.1709 10.2031 13.9216C9.97007 13.6724 9.78632 13.3755 9.66287 13.0486C9.53942 12.7217 9.47879 12.3716 9.48461 12.0192C9.49042 11.6667 9.56256 11.3191 9.69672 10.9971C9.83087 10.6751 9.96199 10.5626 10.2031 10.3222" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
        <path d="M3 3L21 21" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
    </svg>`;

    const initialState = captureFormState();

    // Activate the first section on load
    sections.forEach((section) => {
        if (section.dataset.section === 'personal') {
            section.classList.add('is-active');
            section.removeAttribute('hidden');
        }
    });

    function captureFormState() {
        const data = new FormData(form);
        const snapshot = {};

        for (const [key, value] of data.entries()) {
            snapshot[key] = value;
        }

        form.querySelectorAll('input[type="checkbox"]').forEach((input) => {
            snapshot[input.name] = input.checked ? 'on' : '';
        });

        return snapshot;
    }

    function restoreFormState(snapshot) {
        Object.entries(snapshot).forEach(([name, value]) => {
            const el = form.elements[name];
            if (!el) return;

            if (el.type === 'checkbox') {
                el.checked = value === 'on';
            } else {
                el.value = value;
            }
        });

        updateBioCounter();
    }

    function setMessage(text, type) {
        messageEl.textContent = text;
        messageEl.classList.remove('success', 'error');

        if (text) {
            messageEl.classList.add('is-visible');
            if (type) messageEl.classList.add(type);

            clearTimeout(setMessage._timer);
            setMessage._timer = setTimeout(() => {
                messageEl.classList.remove('is-visible', 'success', 'error');
            }, 3000);
        } else {
            messageEl.classList.remove('is-visible');
        }
    }

    function updateBioCounter() {
        if (!bioTextarea || !bioCounter) return;
        bioCounter.textContent = bioTextarea.value.length;
    }

    function switchSection(sectionName) {
        const currentActive = document.querySelector('.profile-section.is-active');
        const nextActive = document.querySelector(`.profile-section[data-section="${sectionName}"]`);

        if (!nextActive || currentActive === nextActive) return;

        // Update nav state
        navItems.forEach((btn) => {
            btn.classList.toggle('active', btn.dataset.section === sectionName);
        });

        // Animate out current
        if (currentActive) {
            currentActive.classList.remove('is-active');
            currentActive.classList.add('is-leaving');
            currentActive.removeAttribute('hidden');

            setTimeout(() => {
                currentActive.classList.remove('is-leaving');
                currentActive.setAttribute('hidden', '');
            }, 250);
        }

        // Animate in next
        nextActive.removeAttribute('hidden');
        // Force reflow to restart transition
        void nextActive.offsetHeight;
        nextActive.classList.add('is-active');
    }

    navItems.forEach((btn) => {
        btn.addEventListener('click', () => {
            const target = btn.dataset.section;
            if (target) switchSection(target);
        });
    });

    if (bioTextarea) {
        bioTextarea.addEventListener('input', updateBioCounter);
        updateBioCounter();
    }

    form.querySelectorAll('.profile-password-toggle').forEach((btn) => {
        btn.addEventListener('click', () => {
            const targetName = btn.dataset.target;
            const input = form.querySelector(`input[name="${targetName}"]`);
            if (!input) return;

            const isPassword = input.type === 'password';
            input.type = isPassword ? 'text' : 'password';
            btn.innerHTML = isPassword ? eyeClosedSVG : eyeOpenSVG;
            btn.setAttribute('aria-label', isPassword ? 'Скрыть пароль' : 'Показать пароль');
        });
    });

    form.addEventListener('submit', (event) => {
        event.preventDefault();

        saveBtn.disabled = true;
        saveBtn.textContent = 'Сохраняем…';
        setMessage('');

        // Visual-only saving. Backend: POST /api/profile/update
        setTimeout(() => {
            saveBtn.disabled = false;
            saveBtn.textContent = 'Сохранить';
            setMessage('Изменения сохранены', 'success');

            const firstname = form.elements.firstname ? form.elements.firstname.value.trim() : '';
            const lastname = form.elements.lastname ? form.elements.lastname.value.trim() : '';
            const heroName = document.querySelector('.profile-hero-name');

            if (heroName && (firstname || lastname)) {
                heroName.textContent = [firstname, lastname].filter(Boolean).join(' ');
            }

            const initialsEl = document.querySelector('.profile-avatar-initials');
            if (initialsEl) {
                const first = firstname.charAt(0).toUpperCase();
                const last = lastname.charAt(0).toUpperCase();
                initialsEl.textContent = (first + last) || '?';
            }

            Object.assign(initialState, captureFormState());
        }, 700);
    });

    document.addEventListener('keydown', (event) => {
        if ((event.ctrlKey || event.metaKey) && event.key === 'Enter') {
            form.requestSubmit();
        }
    });
})();
