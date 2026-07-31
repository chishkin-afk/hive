const createServiceBtn = document.getElementById('createServiceBtn');
createServiceBtn.addEventListener('click', () => {
    const overlay = document.getElementById('serviceCreateModalOverlay');

    document.body.classList.add('block-scroll');
    overlay.classList.add('open');
    createServiceBtn.classList.add('hide');

    overlay.addEventListener('mousedown', function(e) {
        if (e.target === this) {
            document.body.classList.remove('block-scroll');
            overlay.classList.remove('open');
            createServiceBtn.classList.remove('hide');
            overlay.removeEventListener('mousedown')
        }
    });
});

const services = document.querySelectorAll('.service-card');
services.forEach(el => {
    el.addEventListener('click', () => {
        const overlay = document.getElementById('serviceModalOverlay');

        document.body.classList.add('block-scroll');
        overlay.classList.add('open');
        createServiceBtn.classList.add('hide');


        overlay.addEventListener('mousedown', function(e) {
            if (e.target === this) {
                document.body.classList.remove('block-scroll');
                overlay.classList.remove('open');
                createServiceBtn.classList.remove('hide');
                overlay.removeEventListener('mousedown')
            }
        });
    })
});
