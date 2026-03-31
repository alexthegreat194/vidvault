// vidvault.go — local video gallery server
// Build:  go build -o vidvault vidvault.go
// Run:    ./vidvault [directory] [--port 8765]
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

// ── supported extensions ─────────────────────────────────────────────────────

var videoExts = map[string]bool{
	".webm": true,
	".mp4":  true,
	".mpv":  true,
	".mkv":  true,
	".mov":  true,
	".avi":  true,
	".m4v":  true,
	".ogv":  true,
}

// ── video metadata ────────────────────────────────────────────────────────────

type Video struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Folder string `json:"folder"`
	Ext    string `json:"ext"`
}

func scanVideos(root string) ([]Video, error) {
	var videos []Video
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		ext := strings.ToLower(filepath.Ext(path))
		if !videoExts[ext] {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		folder := filepath.Dir(rel)
		if folder == "." {
			folder = "/"
		}
		videos = append(videos, Video{
			Name:   filepath.Base(path),
			Path:   filepath.ToSlash(rel),
			Folder: filepath.ToSlash(folder),
			Ext:    ext,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(videos, func(i, j int) bool {
		return videos[i].Path < videos[j].Path
	})
	return videos, nil
}

// ── server ────────────────────────────────────────────────────────────────────

type server struct {
	root string
	mux  *http.ServeMux
}

func newServer(root string) *server {
	s := &server{root: root, mux: http.NewServeMux()}
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/api/videos", s.handleVideos)
	s.mux.HandleFunc("/video", s.handleVideo)
	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, indexHTML)
}

func (s *server) handleVideos(w http.ResponseWriter, r *http.Request) {
	videos, err := scanVideos(s.root)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(videos)
}

func (s *server) handleVideo(w http.ResponseWriter, r *http.Request) {
	relPath, err := url.QueryUnescape(r.URL.Query().Get("path"))
	if err != nil || relPath == "" {
		http.Error(w, "missing path", http.StatusBadRequest)
		return
	}
	// Sanitise: no directory traversal
	clean := filepath.Clean(filepath.FromSlash(relPath))
	if strings.HasPrefix(clean, "..") {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	abs := filepath.Join(s.root, clean)

	// Double-check the resolved path stays inside root
	rootAbs, _ := filepath.Abs(s.root)
	fileAbs, err := filepath.Abs(abs)
	if err != nil || !strings.HasPrefix(fileAbs, rootAbs+string(os.PathSeparator)) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	f, err := os.Open(fileAbs)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(fileAbs))
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Accept-Ranges", "bytes")

	// Serve with range support (needed for video scrubbing)
	fi, _ := f.Stat()
	http.ServeContent(w, r, filepath.Base(fileAbs), fi.ModTime(), f)
}

// ── open browser ──────────────────────────────────────────────────────────────

func openBrowser(url string) {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd, args = "open", []string{url}
	case "windows":
		cmd, args = "cmd", []string{"/c", "start", url}
	default:
		cmd, args = "xdg-open", []string{url}
	}
	exec.Command(cmd, args...).Start()
}

// ── main ──────────────────────────────────────────────────────────────────────

func main() {
	port := flag.Int("port", 8765, "port to listen on")
	flag.IntVar(port, "p", 8765, "port to listen on (shorthand)")
	flag.Parse()

	dir := "."
	if flag.NArg() > 0 {
		dir = flag.Arg(0)
	}

	root, err := filepath.Abs(dir)
	if err != nil {
		log.Fatalf("error: cannot resolve path %q: %v", dir, err)
	}
	if fi, err := os.Stat(root); err != nil || !fi.IsDir() {
		log.Fatalf("error: %q is not a directory", root)
	}

	addr := "127.0.0.1:" + strconv.Itoa(*port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("error: cannot bind to %s: %v", addr, err)
	}

	browserURL := "http://localhost:" + strconv.Itoa(*port)
	fmt.Printf("\n  VIDVAULT  →  %s\n", browserURL)
	fmt.Printf("  scanning   →  %s\n\n", root)
	fmt.Println("  Press Ctrl+C to stop.\n")

	go openBrowser(browserURL)

	log.Fatal(http.Serve(ln, newServer(root)))
}

// ── embedded HTML ─────────────────────────────────────────────────────────────

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>VIDVAULT</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=DM+Mono:ital,wght@0,300;0,400;0,500;1,300&family=Bebas+Neue&display=swap">
<style>
  *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

  :root {
    --bg:       #0a0a0b;
    --surface:  #111114;
    --border:   #222228;
    --muted:    #3a3a44;
    --text:     #c8c8d4;
    --dim:      #5a5a6a;
    --accent:   #e8ff47;
    --accent2:  #ff6b47;
    --radius:   4px;
    --mono:     'DM Mono', monospace;
    --display:  'Bebas Neue', sans-serif;
  }

  html { scroll-behavior: smooth; }

  body {
    background: var(--bg);
    color: var(--text);
    font-family: var(--mono);
    font-size: 13px;
    min-height: 100vh;
    display: grid;
    grid-template-rows: auto 1fr;
  }

  header {
    position: sticky; top: 0; z-index: 100;
    background: var(--bg);
    border-bottom: 1px solid var(--border);
    padding: 14px 24px;
    display: flex; align-items: center; gap: 20px; flex-wrap: wrap;
  }

  .logo { font-family: var(--display); font-size: 32px; letter-spacing: 3px; color: var(--accent); line-height: 1; flex-shrink: 0; }
  .logo span { color: var(--accent2); }

  .stats { color: var(--dim); font-size: 11px; flex-shrink: 0; }
  .stats b { color: var(--text); }

  .controls { display: flex; align-items: center; gap: 10px; margin-left: auto; flex-wrap: wrap; }

  .search-wrap { position: relative; }
  .search-wrap svg { position: absolute; left: 9px; top: 50%; transform: translateY(-50%); color: var(--muted); pointer-events: none; }

  #search {
    background: var(--surface); border: 1px solid var(--border);
    color: var(--text); font-family: var(--mono); font-size: 12px;
    padding: 6px 10px 6px 30px; border-radius: var(--radius); width: 220px;
    outline: none; transition: border-color .15s;
  }
  #search:focus { border-color: var(--accent); }
  #search::placeholder { color: var(--dim); }

  .btn-group { display: flex; gap: 2px; }

  .btn {
    background: var(--surface); border: 1px solid var(--border);
    color: var(--dim); font-family: var(--mono); font-size: 11px;
    padding: 6px 10px; border-radius: var(--radius); cursor: pointer;
    transition: all .12s; letter-spacing: .05em;
  }
  .btn:hover { color: var(--text); border-color: var(--muted); }
  .btn.active { background: var(--accent); color: #000; border-color: var(--accent); font-weight: 500; }

  select.btn {
    appearance: none; padding-right: 22px;
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='10' height='6' viewBox='0 0 10 6'%3E%3Cpath d='M0 0l5 6 5-6z' fill='%235a5a6a'/%3E%3C/svg%3E");
    background-repeat: no-repeat; background-position: right 7px center;
  }

  .shell { display: grid; grid-template-columns: 220px 1fr; min-height: 0; }

  aside {
    border-right: 1px solid var(--border); padding: 16px 0;
    overflow-y: auto; position: sticky; top: 57px; height: calc(100vh - 57px);
  }

  .sidebar-label { padding: 0 16px 8px; font-size: 10px; letter-spacing: .12em; color: var(--dim); text-transform: uppercase; }

  .folder-btn {
    display: flex; align-items: center; gap: 8px; width: 100%;
    padding: 6px 16px; background: none; border: none; color: var(--dim);
    font-family: var(--mono); font-size: 12px; text-align: left; cursor: pointer;
    transition: all .1s; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
  }
  .folder-btn:hover { color: var(--text); background: rgba(255,255,255,.03); }
  .folder-btn.active { color: var(--accent); background: rgba(232,255,71,.05); }
  .folder-btn .count { margin-left: auto; font-size: 10px; color: var(--muted); flex-shrink: 0; }
  .folder-btn.active .count { color: var(--accent); opacity: .6; }

  main { padding: 20px 24px; overflow-y: auto; }

  #gallery { display: grid; grid-template-columns: repeat(auto-fill, minmax(260px, 1fr)); gap: 12px; }
  #gallery.list-view { grid-template-columns: 1fr; gap: 4px; }

  .card {
    background: var(--surface); border: 1px solid var(--border);
    border-radius: var(--radius); overflow: hidden; cursor: pointer;
    transition: border-color .15s, transform .15s;
    animation: fadeIn .25s ease both;
  }
  @keyframes fadeIn { from { opacity: 0; transform: translateY(6px); } to { opacity: 1; transform: translateY(0); } }
  .card:hover { border-color: var(--muted); transform: translateY(-2px); }

  .thumb { width: 100%; aspect-ratio: 16/9; background: #0e0e11; position: relative; overflow: hidden; }
  .thumb video { width: 100%; height: 100%; object-fit: cover; display: block; opacity: 0; transition: opacity .3s; }
  .thumb video.loaded { opacity: 1; }

  .play-icon { position: absolute; inset: 0; display: flex; align-items: center; justify-content: center; opacity: 1; transition: opacity .2s; pointer-events: none; }
  .play-icon svg { width: 36px; height: 36px; filter: drop-shadow(0 2px 8px rgba(0,0,0,.8)); }
  .card:hover .play-icon { opacity: 0; }

  .card-meta { padding: 10px 12px; display: flex; flex-direction: column; gap: 4px; }
  .card-name { font-size: 12px; color: var(--text); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; line-height: 1.4; }
  .card-path { font-size: 10px; color: var(--dim); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .card-ext { display: inline-block; padding: 1px 5px; border: 1px solid var(--border); border-radius: 2px; font-size: 9px; color: var(--muted); letter-spacing: .08em; text-transform: uppercase; align-self: flex-start; }

  #gallery.list-view .card { display: grid; grid-template-columns: 180px 1fr; transform: none !important; }
  #gallery.list-view .thumb { aspect-ratio: unset; height: 60px; }
  #gallery.list-view .card-meta { padding: 8px 14px; flex-direction: row; align-items: center; gap: 12px; }
  #gallery.list-view .card-name { font-size: 13px; flex: 1; }
  #gallery.list-view .card-path { display: none; }
  #gallery.list-view .play-icon svg { width: 22px; height: 22px; }

  #modal { display: none; position: fixed; inset: 0; z-index: 999; background: rgba(0,0,0,.92); align-items: center; justify-content: center; backdrop-filter: blur(4px); }
  #modal.open { display: flex; }

  .modal-inner { width: min(90vw, 1100px); background: var(--surface); border: 1px solid var(--border); border-radius: 6px; overflow: hidden; display: flex; flex-direction: column; max-height: 92vh; animation: modalIn .2s ease; }
  @keyframes modalIn { from { opacity: 0; transform: scale(.97); } to { opacity: 1; transform: scale(1); } }

  .modal-header { display: flex; align-items: center; gap: 12px; padding: 12px 16px; border-bottom: 1px solid var(--border); }
  .modal-title { font-size: 13px; color: var(--text); flex: 1; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .modal-path { font-size: 10px; color: var(--dim); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 400px; }

  .modal-nav { display: flex; gap: 4px; }
  .nav-btn { background: var(--bg); border: 1px solid var(--border); color: var(--text); font-family: var(--mono); font-size: 14px; width: 30px; height: 30px; border-radius: var(--radius); cursor: pointer; display: flex; align-items: center; justify-content: center; transition: all .12s; }
  .nav-btn:hover { border-color: var(--accent); color: var(--accent); }

  .close-btn { background: none; border: none; color: var(--dim); font-size: 20px; cursor: pointer; line-height: 1; padding: 4px; transition: color .12s; }
  .close-btn:hover { color: var(--text); }

  #modal-video { width: 100%; max-height: calc(92vh - 52px); background: #000; display: block; }

  .empty { grid-column: 1/-1; padding: 60px; text-align: center; color: var(--dim); font-size: 13px; }
  .empty strong { color: var(--text); display: block; font-size: 16px; margin-bottom: 6px; }

  ::-webkit-scrollbar { width: 6px; height: 6px; }
  ::-webkit-scrollbar-track { background: transparent; }
  ::-webkit-scrollbar-thumb { background: var(--muted); border-radius: 3px; }
  ::-webkit-scrollbar-thumb:hover { background: var(--dim); }
</style>
</head>
<body>

<header>
  <div class="logo">VID<span>VAULT</span></div>
  <div class="stats" id="stats"></div>
  <div class="controls">
    <div class="search-wrap">
      <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/></svg>
      <input id="search" type="search" placeholder="filter files…" autocomplete="off">
    </div>
    <select class="btn" id="sort-select">
      <option value="name">name ↑</option>
      <option value="name-desc">name ↓</option>
      <option value="folder">folder</option>
      <option value="ext">type</option>
    </select>
    <div class="btn-group">
      <button class="btn active" id="grid-btn" title="Grid view">⊞</button>
      <button class="btn"        id="list-btn" title="List view">☰</button>
    </div>
  </div>
</header>

<div class="shell">
  <aside>
    <div class="sidebar-label">Folders</div>
    <div id="folder-list"></div>
  </aside>
  <main>
    <div id="gallery"></div>
  </main>
</div>

<div id="modal">
  <div class="modal-inner">
    <div class="modal-header">
      <span class="modal-title" id="modal-title"></span>
      <span class="modal-path"  id="modal-path"></span>
      <div class="modal-nav">
        <button class="nav-btn" id="prev-btn" title="Previous">‹</button>
        <button class="nav-btn" id="next-btn" title="Next">›</button>
      </div>
      <button class="close-btn" id="close-btn" title="Close (Esc)">✕</button>
    </div>
    <video id="modal-video" controls autoplay></video>
  </div>
</div>

<script>
  const gallery    = document.getElementById('gallery');
  const folderList = document.getElementById('folder-list');
  const searchEl   = document.getElementById('search');
  const statsEl    = document.getElementById('stats');
  const sortSel    = document.getElementById('sort-select');
  const gridBtn    = document.getElementById('grid-btn');
  const listBtn    = document.getElementById('list-btn');
  const modal      = document.getElementById('modal');
  const modalVid   = document.getElementById('modal-video');
  const modalTitle = document.getElementById('modal-title');
  const modalPath  = document.getElementById('modal-path');
  const prevBtn    = document.getElementById('prev-btn');
  const nextBtn    = document.getElementById('next-btn');
  const closeBtn   = document.getElementById('close-btn');

  let ALL_VIDEOS = [];
  let filtered   = [];
  let activeFolder = '__all__';
  let currentIdx   = -1;

  async function init() {
    const r = await fetch('/api/videos');
    ALL_VIDEOS = await r.json();
    buildFolderNav();
    render();
  }

  function buildFolderNav() {
    const folders = ['__all__', ...new Set(ALL_VIDEOS.map(v => v.folder || '/'))].sort();
    folderList.innerHTML = '';
    for (const f of folders) {
      const count = f === '__all__'
        ? ALL_VIDEOS.length
        : ALL_VIDEOS.filter(v => (v.folder || '/') === f).length;
      const btn = document.createElement('button');
      btn.className = 'folder-btn' + (f === activeFolder ? ' active' : '');
      btn.dataset.folder = f;
      btn.innerHTML = ` + "`" + `
        <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>
        <span>${f === '__all__' ? 'All files' : f}</span>
        <span class="count">${count}</span>` + "`" + `;
      btn.addEventListener('click', () => {
        activeFolder = f;
        document.querySelectorAll('.folder-btn').forEach(b => b.classList.remove('active'));
        btn.classList.add('active');
        render();
      });
      folderList.appendChild(btn);
    }
  }

  function render() {
    const q    = searchEl.value.toLowerCase();
    const sort = sortSel.value;

    filtered = ALL_VIDEOS.filter(v => {
      const inFolder = activeFolder === '__all__' || (v.folder || '/') === activeFolder;
      const inSearch = !q || v.name.toLowerCase().includes(q) || (v.folder||'').toLowerCase().includes(q);
      return inFolder && inSearch;
    });

    filtered.sort((a, b) => {
      if (sort === 'name')      return a.name.localeCompare(b.name);
      if (sort === 'name-desc') return b.name.localeCompare(a.name);
      if (sort === 'folder')    return (a.folder||'').localeCompare(b.folder||'') || a.name.localeCompare(b.name);
      if (sort === 'ext')       return a.ext.localeCompare(b.ext) || a.name.localeCompare(b.name);
      return 0;
    });

    statsEl.innerHTML = ` + "`" + `<b>${filtered.length}</b> / ${ALL_VIDEOS.length} videos` + "`" + `;
    gallery.innerHTML = '';

    if (!filtered.length) {
      gallery.innerHTML = '<div class="empty"><strong>Nothing here</strong>Try a different search or folder</div>';
      return;
    }

    filtered.forEach((v, i) => {
      const card = document.createElement('div');
      card.className = 'card';
      card.style.animationDelay = Math.min(i, 30) * 15 + 'ms';
      card.innerHTML = ` + "`" + `
        <div class="thumb">
          <video preload="metadata" muted></video>
          <div class="play-icon">
            <svg viewBox="0 0 36 36" fill="none" xmlns="http://www.w3.org/2000/svg">
              <circle cx="18" cy="18" r="17" fill="rgba(0,0,0,.55)" stroke="rgba(255,255,255,.25)" stroke-width="1"/>
              <polygon points="14,11 27,18 14,25" fill="white"/>
            </svg>
          </div>
        </div>
        <div class="card-meta">
          <span class="card-name" title="${v.name}">${v.name}</span>
          <span class="card-path" title="${v.folder}">${v.folder || '/'}</span>
          <span class="card-ext">${v.ext.slice(1)}</span>
        </div>` + "`" + `;

      const vid = card.querySelector('video');
      const obs = new IntersectionObserver(entries => {
        if (entries[0].isIntersecting) {
          vid.src = '/video?path=' + encodeURIComponent(v.path) + '#t=2';
          vid.addEventListener('loadeddata', () => vid.classList.add('loaded'), { once: true });
          obs.disconnect();
        }
      }, { threshold: 0.1 });
      obs.observe(card);

      card.addEventListener('click', () => openModal(i));
      gallery.appendChild(card);
    });
  }

  function openModal(idx) {
    currentIdx = idx;
    const v = filtered[idx];
    modalTitle.textContent = v.name;
    modalPath.textContent  = v.folder || '/';
    modalVid.src = '/video?path=' + encodeURIComponent(v.path);
    modal.classList.add('open');
    modalVid.focus();
  }

  function closeModal() {
    modal.classList.remove('open');
    modalVid.pause();
    modalVid.src = '';
  }

  closeBtn.addEventListener('click', closeModal);
  modal.addEventListener('click', e => { if (e.target === modal) closeModal(); });
  prevBtn.addEventListener('click', () => { if (currentIdx > 0) openModal(currentIdx - 1); });
  nextBtn.addEventListener('click', () => { if (currentIdx < filtered.length - 1) openModal(currentIdx + 1); });

  document.addEventListener('keydown', e => {
    if (!modal.classList.contains('open')) return;
    if (e.key === 'Escape')     closeModal();
    if (e.key === 'ArrowLeft')  prevBtn.click();
    if (e.key === 'ArrowRight') nextBtn.click();
  });

  modalVid.addEventListener('ended', () => {
    if (currentIdx < filtered.length - 1) openModal(currentIdx + 1);
  });

  searchEl.addEventListener('input', render);
  sortSel.addEventListener('change', render);

  gridBtn.addEventListener('click', () => {
    gallery.classList.remove('list-view');
    gridBtn.classList.add('active');
    listBtn.classList.remove('active');
  });

  listBtn.addEventListener('click', () => {
    gallery.classList.add('list-view');
    listBtn.classList.add('active');
    gridBtn.classList.remove('active');
  });

  init();
</script>
</body>
</html>`
