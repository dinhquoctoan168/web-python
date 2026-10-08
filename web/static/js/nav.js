/**
 * Shared Navigation Script
 * Hỗ trợ chuyển đổi menu mobile và dropdown menu compact trong IDE
 */

function toggleNavMenu(btn) {
    const navLinks = document.getElementById('appNavLinks');
    if (!navLinks) return;
    const isExpanded = btn.getAttribute('aria-expanded') === 'true';
    btn.setAttribute('aria-expanded', String(!isExpanded));
    navLinks.classList.toggle('show');
}

function toggleCompactMenu(e) {
    if (e) e.stopPropagation();
    const dropdown = document.getElementById('compactMenuDropdown');
    const btn = e ? e.currentTarget : document.querySelector('.btn-compact-dropdown');
    if (!dropdown) return;
    const isOpen = dropdown.classList.toggle('show');
    if (btn) {
        btn.setAttribute('aria-expanded', String(isOpen));
    }
}

function closeAllNavMenus() {
    const dropdown = document.getElementById('compactMenuDropdown');
    if (dropdown && dropdown.classList.contains('show')) {
        dropdown.classList.remove('show');
        const btn = document.querySelector('.btn-compact-dropdown');
        if (btn) {
            btn.setAttribute('aria-expanded', 'false');
            btn.focus();
        }
    }

    const navLinks = document.getElementById('appNavLinks');
    if (navLinks && navLinks.classList.contains('show')) {
        navLinks.classList.remove('show');
        const toggleBtn = document.querySelector('.nav-toggle-btn');
        if (toggleBtn) {
            toggleBtn.setAttribute('aria-expanded', 'false');
            toggleBtn.focus();
        }
    }
}

// Đóng dropdown khi click ra ngoài
document.addEventListener('click', function(e) {
    const dropdown = document.getElementById('compactMenuDropdown');
    if (dropdown && dropdown.classList.contains('show')) {
        if (!e.target.closest('.compact-dropdown')) {
            dropdown.classList.remove('show');
            const btn = document.querySelector('.btn-compact-dropdown');
            if (btn) btn.setAttribute('aria-expanded', 'false');
        }
    }

    const navLinks = document.getElementById('appNavLinks');
    if (navLinks && navLinks.classList.contains('show')) {
        if (!e.target.closest('.nav-links') && !e.target.closest('.nav-toggle-btn')) {
            navLinks.classList.remove('show');
            const toggleBtn = document.querySelector('.nav-toggle-btn');
            if (toggleBtn) toggleBtn.setAttribute('aria-expanded', 'false');
        }
    }
});

// Phím tắt Escape để đóng menu đang mở
document.addEventListener('keydown', function(e) {
    if (e.key === 'Escape' || e.key === 'Esc') {
        closeAllNavMenus();
    }
});
