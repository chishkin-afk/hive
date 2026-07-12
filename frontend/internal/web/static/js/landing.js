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

document.addEventListener('DOMContentLoaded', () => {
  const observer = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        entry.target.classList.add('visible');
        observer.unobserve(entry.target);
      }
    });
  }, { threshold: 0.1 });
  
  document.querySelectorAll('.section-title, .section-subtitle, .feature-card, .pricing-card')
    .forEach(el => observer.observe(el));
});