$(function () {
  $('[data-toggle="tooltip"]').tooltip();
});

// ── Dark mode ────────────────────────────────────────────────────────────────
(function () {
  var html   = document.documentElement;
  var toggle = document.getElementById('theme-toggle');
  var icon   = document.getElementById('theme-icon');

  function applyTheme(theme) {
    html.setAttribute('data-theme', theme);
    if (icon) icon.className = theme === 'dark' ? 'fas fa-sun' : 'fas fa-moon';
    var frame = document.querySelector('.utterances-frame');
    if (frame) {
      frame.contentWindow.postMessage(
        { type: 'set-theme', theme: theme === 'dark' ? 'github-dark' : 'github-light' },
        'https://utteranc.es'
      );
    }
  }

  applyTheme(html.getAttribute('data-theme') || 'light');

  if (toggle) {
    toggle.addEventListener('click', function () {
      var next = html.getAttribute('data-theme') === 'dark' ? 'light' : 'dark';
      applyTheme(next);
      localStorage.setItem('theme', next);
    });
  }

  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', function (e) {
    if (!localStorage.getItem('theme')) applyTheme(e.matches ? 'dark' : 'light');
  });
})();

// ── Reading progress bar ─────────────────────────────────────────────────────
(function () {
  var bar = document.getElementById('reading-progress');
  if (!bar) return;
  function update() {
    var el  = document.documentElement;
    var pct = el.scrollHeight - el.clientHeight > 0
      ? ((window.pageYOffset || el.scrollTop) / (el.scrollHeight - el.clientHeight)) * 100 : 0;
    bar.style.transform = 'scaleX(' + (pct / 100) + ')';
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
    btn.classList.toggle('visible', (window.pageYOffset || document.documentElement.scrollTop) > 300);
  }, { passive: true });
  btn.addEventListener('click', function () { window.scrollTo({ top: 0, behavior: 'smooth' }); });
})();

// ── Table of contents ─────────────────────────────────────────────────────────
(function () {
  var tocEl   = document.getElementById('toc');
  var body    = document.querySelector('.post-body');
  if (!tocEl || !body) return;

  var headings = body.querySelectorAll('h2, h3');
  if (headings.length < 3) return;

  // Assign IDs
  headings.forEach(function (h, i) {
    if (!h.id) {
      h.id = (h.textContent.trim().toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '')) || ('h-' + i);
    }
  });

  var nav = document.createElement('nav');
  nav.className = 'toc';

  var btn = document.createElement('button');
  btn.className = 'toc-toggle';
  btn.setAttribute('aria-expanded', 'true');
  btn.appendChild(document.createTextNode('Contents '));
  var chevron = document.createElement('i');
  chevron.className = 'fas fa-chevron-up toc-chevron';
  btn.appendChild(chevron);
  nav.appendChild(btn);

  var list = document.createElement('ol');
  list.className = 'toc-list';
  headings.forEach(function (h) {
    var li = document.createElement('li');
    li.className = h.tagName === 'H3' ? 'toc-h3' : 'toc-h2';
    var link = document.createElement('a');
    link.setAttribute('href', '#' + h.id);
    link.textContent = h.textContent;
    li.appendChild(link);
    list.appendChild(li);
  });
  nav.appendChild(list);
  tocEl.appendChild(nav);

  btn.addEventListener('click', function () {
    var open = btn.getAttribute('aria-expanded') === 'true';
    btn.setAttribute('aria-expanded', String(!open));
    list.style.display = open ? 'none' : '';
    nav.querySelector('.toc-chevron').style.transform = open ? 'rotate(180deg)' : '';
  });

  window.addEventListener('scroll', function () {
    var pos = (window.pageYOffset || document.documentElement.scrollTop) + 90;
    var active = null;
    headings.forEach(function (h) { if (h.offsetTop <= pos) active = h.id; });
    list.querySelectorAll('a').forEach(function (a) {
      a.classList.toggle('toc-active', a.getAttribute('href') === '#' + active);
    });
  }, { passive: true });
})();

// ── Heading anchor links ──────────────────────────────────────────────────────
(function () {
  document.querySelectorAll('.post-body h2, .post-body h3, .post-body h4').forEach(function (h) {
    if (!h.id) return;
    var a = document.createElement('a');
    a.className = 'heading-anchor';
    a.href = '#' + h.id;
    a.setAttribute('aria-label', 'Link to this section');
    a.textContent = '#';
    h.appendChild(a);
  });
})();

// ── Code language labels ──────────────────────────────────────────────────────
(function () {
  document.querySelectorAll('div[class*="language-"]').forEach(function (block) {
    var m = block.className.match(/language-(\w+)/);
    if (!m || m[1] === 'plaintext' || m[1] === 'text') return;
    var label = document.createElement('span');
    label.className = 'code-lang-label';
    label.textContent = m[1];
    block.style.position = 'relative';
    block.appendChild(label);
  });
})();

// ── Copy-to-clipboard ────────────────────────────────────────────────────────
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
        setTimeout(function () { btn.innerHTML = '<i class="fas fa-copy"></i>'; }, 2000);
      });
    });
  });
})();

// ── Lazy-load images ─────────────────────────────────────────────────────────
(function () {
  document.querySelectorAll('img:not([loading])').forEach(function (img) {
    img.setAttribute('loading', 'lazy');
  });
})();

// ── Utterances comments ──────────────────────────────────────────────────────
(function () {
  var container = document.getElementById('comments');
  if (!container) return;
  var theme  = document.documentElement.getAttribute('data-theme') === 'dark' ? 'github-dark' : 'github-light';
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
  window.addEventListener('load', function () { navigator.serviceWorker.register('/sw.js'); });
}
