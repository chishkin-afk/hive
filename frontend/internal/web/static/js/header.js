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
