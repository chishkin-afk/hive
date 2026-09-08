(function() {
    'use strict';

    const CONSENT_KEY = 'cookie_consent';
    const CONSENT_EXPIRY_DAYS = 365;

    const banner = document.getElementById('cookieConsentBanner');
    const acceptBtn = document.getElementById('cookieAcceptBtn');
    const rejectBtn = document.getElementById('cookieRejectBtn');

    function getCookieConsent() {
        return localStorage.getItem(CONSENT_KEY);
    }

    function setCookieConsent(status) {
        localStorage.setItem(CONSENT_KEY, status);
        document.cookie = CONSENT_KEY + '=' + status + '; path=/; max-age=' + (CONSENT_EXPIRY_DAYS * 86400);
    }

    function showBanner() {
        if (banner) {
            banner.style.display = 'block';
            requestAnimationFrame(() => {
                banner.classList.add('show');
            });
        }
    }

    function hideBanner() {
        if (banner) {
            banner.classList.remove('show');
            setTimeout(() => {
                banner.style.display = 'none';
            }, 400);
        }
    }

    function loadYandexMetrika() {
        if (window.ym) return;

        (function(m,e,t,r,i,k,a){
            m[i]=m[i]||function(){(m[i].a=m[i].a||[]).push(arguments)};
            m[i].l=1*new Date();
            for (var j = 0; j < document.scripts.length; j++) {if (document.scripts[j].src === r) { return; }}
            k=e.createElement(t),a=e.getElementsByTagName(t)[0],k.async=1,k.src=r,a.parentNode.insertBefore(k,a)
        })(window, document,'script','https://mc.yandex.ru/metrika/tag.js?id=112359025', 'ym');

        ym(112359025, 'init', {ssr:true, webvisor:true, clickmap:true, referrer: document.referrer, url: location.href, accurateTrackBounce:true, trackLinks:true});
    }

    function handleAccept() {
        setCookieConsent('accepted');
        hideBanner();
        loadYandexMetrika();
    }

    function handleReject() {
        setCookieConsent('rejected');
        hideBanner();
    }

    function init() {
        const consent = getCookieConsent();

        if (consent === 'accepted') {
            loadYandexMetrika();
            if (banner) banner.style.display = 'none';
            return;
        }

        if (consent === 'rejected') {
            if (banner) banner.style.display = 'none';
            return;
        }

        showBanner();
    }

    if (acceptBtn) acceptBtn.addEventListener('click', handleAccept);
    if (rejectBtn) rejectBtn.addEventListener('click', handleReject);

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();
