(function() {
    var $ = function(id) { return document.getElementById(id); };

    document.querySelectorAll('.password-toggle').forEach(function(el) {
        el.addEventListener('click', function() {
            var input = this.closest('.password-wrapper').querySelector('.form-input');
            var hidden = input.type === 'password';
            input.type = hidden ? 'text' : 'password';
            this.innerHTML = hidden
                ? '<svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"><path d="M16.86 17.1975C15.4614 18.3376 13.7584 18.9693 12 19C6.27273 19 3 12 3 12C4.01773 9.97167 5.28929 8.39703 7 7M10.2818 5.21002C10.845 5.06904 11.4216 4.99857 12 5.00002C17.7273 5.00002 21 12 21 12C20.5033 12.9937 19.911 13.9292 19.2327 14.7913M13.7345 13.855C13.5098 14.1129 13.2389 14.3198 12.9378 14.4632C12.6367 14.6067 12.3116 14.6839 11.9821 14.6901C11.6525 14.6963 11.3251 14.6315 11.0195 14.4994C10.7139 14.3674 10.4362 14.1709 10.2031 13.9216C9.97007 13.6724 9.78632 13.3755 9.66287 13.0486C9.53942 12.7217 9.47879 12.3716 9.48461 12.0192C9.49042 11.6667 9.56256 11.3191 9.69672 10.9971C9.83087 10.6751 9.96199 10.5626 10.2031 10.3222" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/><path d="M3 3L21 21" stroke-width="2" stroke-linecap="round"/></svg>'
                : '<svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"><path d="M3 12C3 12 6.27273 5 12 5C17.7273 5 21 12 21 12C21 12 17.7273 19 12 19C6.27273 19 3 12 3 12Z" stroke-width="2" stroke-linejoin="round"/><path d="M12 15C13.6569 15 15 13.6569 15 12C15 10.3431 13.6569 9 12 9C10.3431 9 9 10.3431 9 12C9 13.6569 10.3431 15 12 15Z" stroke-width="2"/></svg>';
        });
    });

    function showNotification(message, type) {
        type = type || 'info';
        var existing = document.querySelector('.notification-toast');
        if (existing) existing.remove();
        var toast = document.createElement('div');
        toast.className = 'notification-toast ' + type;
        toast.textContent = message;
        document.body.appendChild(toast);
        requestAnimationFrame(function() { toast.classList.add('show'); });
        setTimeout(function() {
            toast.classList.remove('show');
            setTimeout(function() { toast.remove(); }, 300);
        }, 4000);
    }

    async function resetPassword() {
        var tokenInput = $('resetToken');
        var passInput = $('resetPassword');
        var token = tokenInput.value.trim();
        var password = passInput.value;

        passInput.classList.remove('error');

        if (!token)    { showNotification('Токен не найден. Перейдите по ссылке из письма заново.', 'error'); return; }
        if (!password) { passInput.classList.add('error'); showNotification('Введите новый пароль', 'error'); return; }
        if (password.length < 6)  { passInput.classList.add('error'); showNotification('Пароль должен быть не менее 6 символов', 'error'); return; }
        if (password.length > 32) { passInput.classList.add('error'); showNotification('Пароль должен быть не более 32 символов', 'error'); return; }

        try {
            await authAPI.change_password(token, password);
            await authAPI.logout();
            showNotification('Пароль успешно изменён! Сейчас вы будете перенаправлены на страницу входа.', 'success');
            setTimeout(function() { window.location.href = '/auth?action=login'; }, 2000);
        } catch (err) {
            var msg = (err.message || 'Произошла ошибка.');
            if (err.status === 401 || err.status === 404) msg = 'Токен недействителен или истёк. Запросите восстановление пароля заново.';
            showNotification(msg, 'error');
        }
    }

    $('resetBtn').addEventListener('click', resetPassword);
    $('resetForm').addEventListener('keydown', function(e) {
        if (e.key === 'Enter') { e.preventDefault(); resetPassword(); }
    });
})();
