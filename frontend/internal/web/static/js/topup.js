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
