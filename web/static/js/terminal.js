(function () {
  'use strict';

  const WS_PATH       = '/ws';
  const RECONNECT_MAX = 5;
  const RECONNECT_MS  = 2000;

  const term = new Terminal({
    cursorBlink:      true,
    cursorStyle:      'block',
    fontFamily:       '"Cascadia Code", "JetBrains Mono", "Fira Code", "Consolas", monospace',
    fontSize:         14,
    lineHeight:       1.25,
    theme: {
      background:    '#0d1117',
      foreground:    '#e6edf3',
      cursor:        '#58a6ff',
      cursorAccent:  '#0d1117',
      selectionBackground: 'rgba(88,166,255,0.25)',
      black:         '#484f58',
      red:           '#f85149',
      green:         '#3fb950',
      yellow:        '#d29922',
      blue:          '#58a6ff',
      magenta:       '#bc8cff',
      cyan:          '#39c5cf',
      white:         '#b1bac4',
      brightBlack:   '#6e7681',
      brightRed:     '#ff7b72',
      brightGreen:   '#56d364',
      brightYellow:  '#e3b341',
      brightBlue:    '#79c0ff',
      brightMagenta: '#d2a8ff',
      brightCyan:    '#56d4dd',
      brightWhite:   '#f0f6fc',
    },
    allowProposedApi: true,
    scrollback:       5000,
    rightClickSelectsWord: false,
  });

  const fitAddon      = new FitAddon.FitAddon();
  const webLinksAddon = new WebLinksAddon.WebLinksAddon();
  const searchAddon   = new SearchAddon.SearchAddon();

  term.loadAddon(fitAddon);
  term.loadAddon(webLinksAddon);
  term.loadAddon(searchAddon);

  term.open(document.getElementById('terminal-container'));
  fitAddon.fit();
  term.focus();

  const resizeObserver = new ResizeObserver(() => {
    fitAddon.fit();
    sendResize();
  });
  resizeObserver.observe(document.getElementById('terminal-container'));

  let ws              = null;
  let reconnectCount  = 0;
  let intentionalClose = false;

  function buildWsUrl() {
    const proto = location.protocol === 'https:' ? 'wss' : 'ws';
    return `${proto}://${location.host}${WS_PATH}`;
  }

  function connect() {
    ws = new WebSocket(buildWsUrl());
    ws.binaryType = 'arraybuffer';

    ws.onopen = () => {
      reconnectCount = 0;
      sendResize();
    };

    ws.onmessage = (event) => {
      if (event.data instanceof ArrayBuffer) {
        term.write(new Uint8Array(event.data));
      } else {
        term.write(event.data);
      }
    };

    ws.onerror = (err) => {
      console.error('WebSocket error:', err);
    };

    ws.onclose = () => {
      if (intentionalClose) return;
      if (reconnectCount < RECONNECT_MAX) {
        reconnectCount++;
        term.writeln(`\r\n\x1b[33m[Disconnected — reconnecting in ${RECONNECT_MS / 1000}s (attempt ${reconnectCount}/${RECONNECT_MAX})]\x1b[0m`);
        setTimeout(connect, RECONNECT_MS);
      } else {
        term.writeln('\r\n\x1b[31m[Connection lost. Refresh the page to reconnect.]\x1b[0m');
      }
    };
  }

  connect();

  term.onData((data) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      const encoded = new TextEncoder().encode(data);
      ws.send(encoded);
    }
  });

  term.onBinary((data) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      const bytes = Uint8Array.from(data, c => c.charCodeAt(0));
      ws.send(bytes);
    }
  });

  function sendResize() {
    if (!ws || ws.readyState !== WebSocket.OPEN) return;
    const msg = JSON.stringify({ type: 'resize', cols: term.cols, rows: term.rows });
    ws.send(msg);
  }

  term.onResize(({ cols, rows }) => {
    sendResize();
  });

  function copyText(text) {
    if (!text) return;
    if (navigator.clipboard && window.isSecureContext) {
      navigator.clipboard.writeText(text).catch((err) => console.warn('Copy failed:', err));
    } else {
      const ta = document.createElement('textarea');
      ta.value = text;
      ta.style.cssText = 'position:fixed;opacity:0;top:0;left:0;';
      document.body.appendChild(ta);
      ta.focus();
      ta.select();
      try { document.execCommand('copy'); } catch (e) { console.warn('execCommand copy failed:', e); }
      document.body.removeChild(ta);
    }
  }

  async function pasteText() {
    if (navigator.clipboard && window.isSecureContext) {
      try {
        const text = await navigator.clipboard.readText();
        if (ws && ws.readyState === WebSocket.OPEN && text)
          ws.send(new TextEncoder().encode(text));
      } catch (err) {
        console.warn('Paste failed:', err);
      }
    } else {
      console.warn('Paste via API unavailable on non-secure origin. Use browser Ctrl+V.');
    }
  }

  const container = document.getElementById('terminal-container');

  container.addEventListener('contextmenu', async (e) => {
    e.preventDefault();

    const selectedText = term.getSelection();

    removeContextMenu();
    const menu = document.createElement('div');
    menu.id = 'ctx-menu';
    menu.style.cssText = `
      position: fixed;
      left: ${e.clientX}px;
      top: ${e.clientY}px;
      background: #161b22;
      border: 1px solid #30363d;
      border-radius: 6px;
      padding: 4px 0;
      z-index: 9999;
      box-shadow: 0 4px 16px rgba(0,0,0,0.5);
      min-width: 130px;
      font-family: 'Segoe UI', sans-serif;
      font-size: 13px;
    `;

    function menuItem(label, shortcut, action, disabled) {
      const item = document.createElement('div');
      item.style.cssText = `
        display: flex;
        justify-content: space-between;
        gap: 16px;
        padding: 6px 14px;
        color: ${disabled ? '#484f58' : '#e6edf3'};
        cursor: ${disabled ? 'default' : 'pointer'};
      `;
      const labelEl = document.createElement('span');
      labelEl.textContent = label;
      const shortEl = document.createElement('span');
      shortEl.textContent = shortcut;
      shortEl.style.cssText = 'color:#484f58;font-size:11px;';
      item.appendChild(labelEl);
      item.appendChild(shortEl);
      if (!disabled) {
        item.addEventListener('mouseenter', () => item.style.background = '#21262d');
        item.addEventListener('mouseleave', () => item.style.background = '');
        item.addEventListener('mousedown', (ev) => {
          ev.preventDefault();
          removeContextMenu();
          try { action(); } catch (e) { console.warn('menu action error:', e); }
        });
      }
      return item;
    }

    menu.appendChild(menuItem('Copy', 'Ctrl+Shift+C', () => {
      copyText(selectedText);
    }, !selectedText));

    menu.appendChild(menuItem('Paste', 'Ctrl+Shift+V', () => {
      pasteText();
    }, false));

    menu.appendChild(menuItem('Select All', 'Ctrl+Shift+A', () => {
      term.selectAll();
    }, false));

    menu.appendChild(menuItem('Clear', '', () => {
      term.clear();
    }, false));

    document.body.appendChild(menu);

    requestAnimationFrame(() => {
      const rect = menu.getBoundingClientRect();
      if (rect.right > window.innerWidth)  menu.style.left = `${e.clientX - rect.width}px`;
      if (rect.bottom > window.innerHeight) menu.style.top = `${e.clientY - rect.height}px`;
    });

    document.addEventListener('mousedown', removeContextMenuOnClickOutside, { once: true, capture: true });
  });

  function removeContextMenu() {
    const existing = document.getElementById('ctx-menu');
    if (existing) existing.remove();
  }

  function removeContextMenuOnClickOutside(e) {
    const menu = document.getElementById('ctx-menu');
    if (menu && !menu.contains(e.target)) removeContextMenu();
  }

  document.addEventListener('keydown', async (e) => {
    if (e.ctrlKey && e.shiftKey && e.key === 'C') {
      e.preventDefault();
      copyText(term.getSelection());
      return;
    }

    if (e.ctrlKey && e.shiftKey && e.key === 'V') {
      e.preventDefault();
      pasteText();
      return;
    }

    if (e.ctrlKey && e.shiftKey && e.key === 'A') {
      e.preventDefault();
      term.selectAll();
      return;
    }

    if (e.ctrlKey && e.shiftKey && e.key === 'F') {
      e.preventDefault();
      toggleSearch();
      return;
    }
  });

  const searchBar   = document.getElementById('search-bar');
  const searchInput = document.getElementById('search-input');
  const btnSearch   = document.getElementById('btn-search');
  const btnFindPrev = document.getElementById('btn-find-prev');
  const btnFindNext = document.getElementById('btn-find-next');
  const btnSearchClose = document.getElementById('btn-search-close');

  function toggleSearch() {
    const visible = !searchBar.classList.contains('hidden');
    if (visible) {
      closeSearch();
    } else {
      openSearch();
    }
  }

  function openSearch() {
    searchBar.classList.remove('hidden');
    document.body.classList.add('search-visible');
    fitAddon.fit();
    searchInput.value = '';
    searchInput.focus();
  }

  function closeSearch() {
    searchBar.classList.add('hidden');
    document.body.classList.remove('search-visible');
    searchAddon.clearDecorations();
    fitAddon.fit();
    term.focus();
  }

  btnSearch.addEventListener('click', toggleSearch);
  btnSearchClose.addEventListener('click', closeSearch);

  searchInput.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') { e.shiftKey ? findPrev() : findNext(); }
    if (e.key === 'Escape') closeSearch();
  });

  searchInput.addEventListener('input', () => {
    searchAddon.clearDecorations();
    const query = searchInput.value;
    if (query) searchAddon.findNext(query, { incremental: true });
  });

  function findNext() {
    const query = searchInput.value;
    if (query) searchAddon.findNext(query);
  }

  function findPrev() {
    const query = searchInput.value;
    if (query) searchAddon.findPrevious(query);
  }

  btnFindNext.addEventListener('click', findNext);
  btnFindPrev.addEventListener('click', findPrev);

  document.addEventListener('visibilitychange', () => {
    if (document.visibilityState === 'visible') term.focus();
  });

  window.addEventListener('beforeunload', () => {
    intentionalClose = true;
    if (ws) ws.close();
  });
})();
