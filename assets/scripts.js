$(function () {
  $('[data-toggle="tooltip"]').tooltip();
});

// ── Dark mode ────────────────────────────────────────────────────────────────
(function () {
  var html = document.documentElement;
  var toggle = document.getElementById('theme-toggle');
  var icon = document.getElementById('theme-icon');

  function applyTheme(theme) {
    html.setAttribute('data-theme', theme);
    if (icon) {
      icon.className = theme === 'dark' ? 'fas fa-sun' : 'fas fa-moon';
    }
  }

  var current = html.getAttribute('data-theme') || 'light';
  applyTheme(current);

  if (toggle) {
    toggle.addEventListener('click', function () {
      var next = html.getAttribute('data-theme') === 'dark' ? 'light' : 'dark';
      applyTheme(next);
      localStorage.setItem('theme', next);
    });
  }

  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', function (e) {
    if (!localStorage.getItem('theme')) {
      applyTheme(e.matches ? 'dark' : 'light');
    }
  });
})();

// ── Reading progress bar ─────────────────────────────────────────────────────
(function () {
  var bar = document.getElementById('reading-progress');
  if (!bar) return;

  function update() {
    var docEl = document.documentElement;
    var scrollTop = docEl.scrollTop || document.body.scrollTop;
    var scrollHeight = docEl.scrollHeight - docEl.clientHeight;
    var pct = scrollHeight > 0 ? (scrollTop / scrollHeight) * 100 : 0;
    bar.style.width = pct + '%';
    bar.setAttribute('aria-valuenow', Math.round(pct));
  }

  window.addEventListener('scroll', update, { passive: true });
})();

// ── Back to top ──────────────────────────────────────────────────────────────
(function () {
  var btn = document.getElementById('back-to-top');
  if (!btn) return;

  window.addEventListener('scroll', function () {
    if (window.scrollY > 300) {
      btn.classList.add('visible');
    } else {
      btn.classList.remove('visible');
    }
  }, { passive: true });

  btn.addEventListener('click', function () {
    window.scrollTo({ top: 0, behavior: 'smooth' });
  });
})();
