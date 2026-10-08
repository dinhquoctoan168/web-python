/**
 * Shared Navigation Script
 * Hỗ trợ chuyển đổi menu mobile và dropdown menu compact trong IDE
 */

function toggleNavMenu(btn) {
    const navLinks = document.getElementById('appNavLinks');
    if (!navLinks) return;
    const isExpanded = btn.getAttribute('aria-expanded') === 'true';
    btn.setAttribute('aria-expanded', !isExpanded);
    navLinks.classList.toggle('show');
}

function toggleCompactMenu(e) {
    if (e) e.stopPropagation();
    const dropdown = document.getElementById('compactMenuDropdown');
    const btn = e ? e.currentTarget : null;
    if (!dropdown) return;
    const isOpen = dropdown.classList.toggle('show');
    if (btn) {
        btn.setAttribute('aria-expanded', isOpen);
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
});
