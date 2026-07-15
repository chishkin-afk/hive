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

    document.querySelectorAll(`
        .section-title, .section-subtitle, .feature-card,
        .pricing-card, .faq-item
    `)
        .forEach(el => observer.observe(el));
});

const faqItems = document.querySelectorAll('.faq-item');
faqItems.forEach(item => {
    item.addEventListener('click', () => {
        const isOpen = item.classList.contains('open');

        faqItems.forEach(item => {
            item.classList.remove('open');
            item.querySelector('.faq-answer').style.maxHeight = '0px';
        })

        if (!isOpen) {
            item.classList.add('open');
            const answer = item.querySelector('.faq-answer');
            const height = answer.querySelector('p').scrollHeight;
            answer.style.maxHeight = height + 'px';
        }
    });
})