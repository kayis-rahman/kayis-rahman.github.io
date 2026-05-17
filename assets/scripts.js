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
    // Sync utterances iframe if it is already loaded
    var frame = document.querySelector('.utterances-frame');
    if (frame) {
      frame.contentWindow.postMessage(
        { type: 'set-theme', theme: theme === 'dark' ? 'github-dark' : 'github-light' },
        'https://utteranc.es'
      );
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
    var scrollTop = window.pageYOffset || docEl.scrollTop;
    var scrollHeight = docEl.scrollHeight - docEl.clientHeight;
    var pct = scrollHeight > 0 ? (scrollTop / scrollHeight) * 100 : 0;
    bar.style.width = pct + '%';
    bar.setAttribute('aria-valuenow', Math.round(pct));
  }

  window.addEventListener('scroll', update, { passive: true });
  update();
})();

// ── Back to top ──────────────────────────────────────────────────────────────
(function () {
  var btn = document.getElementById('back-to-top');
  if (!btn) return;

  window.addEventListener('scroll', function () {
    if ((window.pageYOffset || document.documentElement.scrollTop) > 300) {
      btn.classList.add('visible');
    } else {
      btn.classList.remove('visible');
    }
  }, { passive: true });

  btn.addEventListener('click', function () {
    window.scrollTo({ top: 0, behavior: 'smooth' });
  });
})();

// ── Copy-to-clipboard for code blocks ───────────────────────────────────────
(function () {
  if (!navigator.clipboard) return;

  document.querySelectorAll('pre').forEach(function (pre) {
    var wrapper = document.createElement('div');
    wrapper.className = 'code-block-wrapper';
    pre.parentNode.insertBefore(wrapper, pre);
    wrapper.appendChild(pre);

    var btn = document.createElement('button');
    btn.className = 'copy-btn';
    btn.setAttribute('aria-label', 'Copy code');
    btn.innerHTML = '<i class="fas fa-copy"></i>';
    wrapper.appendChild(btn);

    btn.addEventListener('click', function () {
      var code = pre.querySelector('code') || pre;
      navigator.clipboard.writeText(code.innerText).then(function () {
        btn.innerHTML = '<i class="fas fa-check"></i>';
        setTimeout(function () {
          btn.innerHTML = '<i class="fas fa-copy"></i>';
        }, 2000);
      });
    });
  });
})();

// ── Utterances comments ──────────────────────────────────────────────────────
(function () {
  var container = document.getElementById('comments');
  if (!container) return;

  var theme = document.documentElement.getAttribute('data-theme') === 'dark'
    ? 'github-dark' : 'github-light';

  var script = document.createElement('script');
  script.src = 'https://utteranc.es/client.js';
  script.setAttribute('repo', 'kayis-rahman/kayis-rahman.github.io');
  script.setAttribute('issue-term', 'pathname');
  script.setAttribute('theme', theme);
  script.setAttribute('crossorigin', 'anonymous');
  script.async = true;
  container.appendChild(script);
})();

// ── Service worker ───────────────────────────────────────────────────────────
if ('serviceWorker' in navigator) {
  window.addEventListener('load', function () {
    navigator.serviceWorker.register('/sw.js');
  });
}
