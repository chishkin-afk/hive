const burgerBtn = document.getElementById('burgerBtn');
const mobileMenuOverflow = document.getElementById('mobileMenuOverflow');

function toggleMobileMenu() {
    burgerBtn.classList.toggle('active');
    mobileMenuOverflow.classList.toggle('open');
    document.body.classList.toggle('block-scroll');
}

burgerBtn.addEventListener('click', toggleMobileMenu);

const menuLinks = document.querySelectorAll('#mobileMenuOverflow .menu-link');
menuLinks.forEach(link => {
    link.addEventListener('click', toggleMobileMenu);
});

const createServiceBtn = document.getElementById('createServiceBtn');
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
        }
    });
    })
});
