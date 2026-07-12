const burgerBtn = document.getElementById('burgerBtn');
const mobileMenu = document.getElementById('mobileMenu');

function toggleMobileMenu() {
    burgerBtn.classList.toggle('active');
    mobileMenu.classList.toggle('open');
    document.body.classList.toggle('block-scroll');
}

burgerBtn.addEventListener('click', toggleMobileMenu);

const menuLinks = document.querySelectorAll('#mobileMenu .menu-link');
menuLinks.forEach(link => {
    link.addEventListener('click', toggleMobileMenu);
});
