window.themeInitialized = true;
document.addEventListener('DOMContentLoaded', () => {
  const themeToggle = document.getElementById('themeToggle');
  if (!themeToggle) {
    console.warn('Кнопка смены темы не найдена');
    return;
  }

  const savedTheme = localStorage.getItem('theme') || 'light';
  document.documentElement.setAttribute('data-theme', savedTheme);

  themeToggle.addEventListener('click', () => {
    const currentTheme = document.documentElement.getAttribute('data-theme');
    const newTheme = currentTheme === 'light' ? 'dark' : 'light';
    
    document.documentElement.setAttribute('data-theme', newTheme);
    localStorage.setItem('theme', newTheme);
  });
});

function initNavbarScroll() {
  const navbar = document.querySelector('.navbar');
  if (!navbar) return;

  window.addEventListener('scroll', () => {
    if (window.scrollY > 20) {
      navbar.style.backdropFilter = 'blur(20px)';
      navbar.style.webkitBackdropFilter = 'blur(20px)';
    } else {
      navbar.style.backdropFilter = 'none';
      navbar.style.webkitBackdropFilter = 'none';
    }
  }, { passive: true });
}

initNavbarScroll();