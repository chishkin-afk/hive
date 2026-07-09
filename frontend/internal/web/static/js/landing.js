class LandingApp {
  constructor() {
    this.init();
  }

  init() {
    this.initTheme();
    this.initFAQ();
    this.loadPricing();
    this.initScrollAnimations();
    this.initSmoothScroll();
  }

  initTheme() {
    const themeToggle = document.getElementById('themeToggle');
    if (!themeToggle) return;

    const savedTheme = localStorage.getItem('theme') || 'light';
    document.documentElement.setAttribute('data-theme', savedTheme);

    themeToggle.addEventListener('click', () => {
      const currentTheme = document.documentElement.getAttribute('data-theme');
      const newTheme = currentTheme === 'light' ? 'dark' : 'light';
      
      document.documentElement.setAttribute('data-theme', newTheme);
      localStorage.setItem('theme', newTheme);
    });
  }

  initFAQ() {
  const faqCards = document.querySelectorAll('.faq-flip-card');
  
  faqCards.forEach(card => {
    card.addEventListener('click', () => {
      faqCards.forEach(c => {
        if (c !== card) c.classList.remove('flipped');
      });
      card.classList.toggle('flipped');
    });
  });
}

  async loadPricing() {
    const pricingGrid = document.getElementById('pricingGrid');
    
    try {
      const data = await paymentsService.getRates();
      this.renderPricing(data.list, pricingGrid);
    } catch (error) {
      console.error('Failed to load pricing:', error);
      pricingGrid.innerHTML = '<p class="error-message">Не удалось загрузить тарифы</p>';
    }
  }

  renderPricing(rates, container) {
    container.innerHTML = rates.map((rate, index) => {
      const isPopular = index === 1;
      
      return `
        <div class="pricing-card ${isPopular ? 'popular' : ''}">
          ${isPopular ? '<span class="pricing-badge">Популярный</span>' : ''}
          <h3 class="pricing-name">${rate.name}</h3>
          <div class="pricing-price">
            <span class="price-value">${rate.amount_penny/100}</span>
            <span class="price-currency">₽</span>
          </div>
          <p class="pricing-period">${rate.paid_days} дней</p>
          <a href="/register" class="btn btn-primary" style="width: 100%;">Выбрать тариф</a>
        </div>
      `;
    }).join('');
  }

  initScrollAnimations() {
    const observer = new IntersectionObserver((entries) => {
      entries.forEach(entry => {
        if (entry.isIntersecting) {
          entry.target.classList.add('visible');
        }
      });
    }, {
      threshold: 0.1,
      rootMargin: '0px 0px -50px 0px'
    });

    document.querySelectorAll('.feature-card, .pricing-card, .faq-item').forEach(el => {
      el.classList.add('animate-on-scroll');
      observer.observe(el);
    });
  }

  initSmoothScroll() {
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
      anchor.addEventListener('click', function(e) {
        e.preventDefault();
        const target = document.querySelector(this.getAttribute('href'));
        
        if (target) {
          const offsetTop = target.offsetTop - 80;
          window.scrollTo({
            top: offsetTop,
            behavior: 'smooth'
          });
        }
      });
    });
  }
}

document.addEventListener('DOMContentLoaded', () => {
  window.landingApp = new LandingApp();
});