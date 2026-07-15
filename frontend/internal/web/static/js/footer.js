const email = document.getElementById('email');
email.addEventListener('click', () => {
    const text = email.textContent
    if (text === 'Скопировано!') {
        return;
    }

    navigator.clipboard.writeText(text)
        .then(() => {
            email.textContent = 'Скопировано!';

            timeoutId = setTimeout(() => {
                email.textContent = text;
            }, 2000);
        });
});
