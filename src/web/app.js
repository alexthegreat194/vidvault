const gallery = document.getElementById("gallery"),
	folderList = document.getElementById("folder-list"),
	searchEl = document.getElementById("search"),
	statsEl = document.getElementById("stats"),
	sortSel = document.getElementById("sort-select"),
	gridBtn = document.getElementById("grid-btn"),
	listBtn = document.getElementById("list-btn"),
	modal = document.getElementById("modal"),
	modalVid = document.getElementById("modal-video"),
	modalTitle = document.getElementById("modal-title"),
	modalPath = document.getElementById("modal-path"),
	prevBtn = document.getElementById("prev-btn"),
	nextBtn = document.getElementById("next-btn"),
	closeBtn = document.getElementById("close-btn"),
	selectBtn = document.getElementById("select-btn"),
	selectBar = document.getElementById("select-bar"),
	selectCount = document.getElementById("select-count"),
	selectMoveBtn = document.getElementById("select-move-btn"),
	selectAllBtn = document.getElementById("select-all-btn"),
	selectClearBtn = document.getElementById("select-clear-btn"),
	newFolderBtn = document.getElementById("new-folder-btn"),
	sidebarFolderForm = document.getElementById("sidebar-folder-form"),
	sidebarFolderInput = document.getElementById("sidebar-folder-input"),
	sidebarFolderConfirm = document.getElementById("sidebar-folder-confirm"),
	sidebarFolderCancel = document.getElementById("sidebar-folder-cancel");

let ALL_VIDEOS = [],
	ALL_FOLDERS = [],
	FOLDER_META = {},
	filtered = [],
	activeFolder = "__all__",
	currentIdx = -1;
let selectMode = false,
	selectedPaths = new Set();

/**
 * Populates ALL_FOLDERS (string[]) and FOLDER_META (name → folderInfo) from
 * the raw /api/folders response array.
 * @param {{ name: string, has_other_files: boolean }[]} data
 */
function parseFolders(data) {
	ALL_FOLDERS = data.map((f) => f.name);
	FOLDER_META = Object.fromEntries(data.map((f) => [f.name, f]));
}



async function init() {
	const [vr, fr] = await Promise.all([
		fetch("/api/videos"),
		fetch("/api/folders"),
	]);
	ALL_VIDEOS = await vr.json();
	parseFolders(await fr.json());
	buildFolderNav();
	render();
	populateUploadFolders();
}
async function refresh() {
	const [vr, fr] = await Promise.all([
		fetch("/api/videos"),
		fetch("/api/folders"),
	]);
	ALL_VIDEOS = await vr.json();
	parseFolders(await fr.json());
	buildFolderNav();
	render();
	populateUploadFolders();
}

/**
 * 
 */
function buildFolderNav() {
	const merged = new Set([
		...ALL_VIDEOS.map((v) => v.folder || "/"),
		...ALL_FOLDERS,
	]);
	const folders = ["__all__", ...[...merged].sort()];
	folderList.innerHTML = "";
	for (const f of folders) {
		const count =
			f === "__all__"
				? ALL_VIDEOS.length
				: ALL_VIDEOS.filter((v) => (v.folder || "/") == f).length;
		const btn = document.createElement("button");
		btn.className = "folder-btn" + (f === activeFolder ? " active" : "");
		btn.dataset.folder = f;
		const hasOther = f !== "__all__" && FOLDER_META[f]?.has_other_files;
		btn.innerHTML =
			'<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>' +
			"<span>" +
			(f === "__all__" ? "All files" : f) +
			"</span>" +
			(hasOther
				? '<span class="folder-warn" title="Contains non-video files">⚠</span>'
				: "") +
			'<span class="count">' +
			count +
			"</span>";
		btn.addEventListener("click", () => {
			activeFolder = f;
			document
				.querySelectorAll(".folder-btn")
				.forEach((b) => b.classList.remove("active"));
			btn.classList.add("active");
			render();
		});
		if (f !== "__all__") {
			btn.addEventListener("dragover", (e) => {
				e.preventDefault();
				btn.classList.add("drag-over");
			});
			btn.addEventListener("dragleave", () =>
				btn.classList.remove("drag-over"),
			);
			btn.addEventListener("drop", async (e) => {
				e.preventDefault();
				btn.classList.remove("drag-over");
				const src = e.dataTransfer.getData("text/plain");
				if (src) await moveVideo(src, f);
			});
			const del = document.createElement("button");
			del.className = "del-folder-btn";
			del.title = "Delete folder (moves files to root)";
			del.textContent = "✕";
			del.addEventListener("click", async (e) => {
				e.stopPropagation();
				const warnMsg = FOLDER_META[f]?.has_other_files
					? "\n\n⚠ This folder contains non-video files. They will also be moved to root."
					: "";
				if (
					!confirm(
						'Delete folder "' +
						f +
						'"? All files inside will be moved to root.' +
						warnMsg,
					)
				)
					return;
				const res = await fetch("/api/rmdir", {
					method: "POST",
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify({ folder: f }),
				});
				if (res.ok) {
					if (activeFolder === f) activeFolder = "__all__";
					toast("Deleted " + f, "success");
					await refresh();
				} else {
					toast("Delete failed: " + (await res.text()), "error");
				}
			});
			btn.appendChild(del);
		}
		folderList.appendChild(btn);
	}
}

function render() {
	const q = searchEl.value.toLowerCase(),
		sort = sortSel.value;
	filtered = ALL_VIDEOS.filter((v) => {
		const inF = activeFolder === "__all__" || (v.folder || "/") == activeFolder;
		const inS =
			!q ||
			v.name.toLowerCase().includes(q) ||
			(v.folder || "").toLowerCase().includes(q);
		return inF && inS;
	});
	filtered.sort((a, b) => {
		if (sort === "name") return a.name.localeCompare(b.name);
		if (sort === "name-desc") return b.name.localeCompare(a.name);
		if (sort === "folder")
			return (
				(a.folder || "").localeCompare(b.folder || "") ||
				a.name.localeCompare(b.name)
			);
		if (sort === "ext")
			return a.ext.localeCompare(b.ext) || a.name.localeCompare(b.name);
		return 0;
	});
	statsEl.innerHTML =
		"<b>" + filtered.length + "</b> / " + ALL_VIDEOS.length + " videos";
	gallery.innerHTML = "";
	if (!filtered.length) {
		const otherWarn =
			activeFolder !== "__all__" && FOLDER_META[activeFolder]?.has_other_files
				? '<span class="empty-warn">⚠ This folder contains non-video files that aren\'t shown here.</span>'
				: "";
		gallery.innerHTML =
			'<div class="empty"><strong>Nothing here</strong>' +
			(otherWarn || "Try a different search or folder") +
			"</div>";
		return;
	}
	filtered.forEach((v, i) => {
		const card = document.createElement("div");
		card.className = "card";
		card.style.animationDelay = Math.min(i, 30) * 15 + "ms";
		card.draggable = true;
		card.dataset.path = v.path;
		card.innerHTML =
			'<div class="thumb"><video preload="metadata" muted></video>' +
			'<div class="play-icon"><svg viewBox="0 0 36 36" fill="none"><circle cx="18" cy="18" r="17" fill="rgba(0,0,0,.55)" stroke="rgba(255,255,255,.25)" stroke-width="1"/><polygon points="14,11 27,18 14,25" fill="white"/></svg></div>' +
			'<div class="drag-handle" title="Drag to move"><svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="0 0 10 18"><circle cx="3" cy="2" r="1.5" fill="white"/><circle cx="7" cy="2" r="1.5" fill="white"/><circle cx="3" cy="9" r="1.5" fill="white"/><circle cx="7" cy="9" r="1.5" fill="white"/><circle cx="3" cy="16" r="1.5" fill="white"/><circle cx="7" cy="16" r="1.5" fill="white"/></svg></div>' +
			'<div class="check-overlay"><svg viewBox="0 0 10 8" fill="none" width="10" height="8"><polyline points="1,4 3.5,7 9,1" stroke="#000" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/></svg></div></div>' +
			'<div class="card-meta"><span class="card-name" title="' +
			escHtml(v.name) +
			'">' +
			escHtml(v.name) +
			"</span>" +
			'<span class="card-path">' +
			escHtml(v.folder || "/") +
			"</span>" +
			'<span class="card-ext">' +
			escHtml(v.ext.slice(1)) +
			"</span></div>";
		const vid = card.querySelector("video");
		const obs = new IntersectionObserver(
			(entries) => {
				if (entries[0].isIntersecting) {
					vid.src = "/video?path=" + encodeURIComponent(v.path) + "#t=2";
					vid.addEventListener(
						"loadeddata",
						() => vid.classList.add("loaded"),
						{ once: true },
					);
					obs.disconnect();
				}
			},
			{ threshold: 0.1 },
		);
		obs.observe(card);
		card.addEventListener("click", (e) => {
			if (e.target.closest(".drag-handle")) return;
			if (selectMode) {
				toggleSelect(v.path, card);
			} else {
				openModal(i);
			}
		});
		card.addEventListener("contextmenu", (e) => {
			e.preventDefault();
			showCtxMenu(e, v);
		});
		card.addEventListener("dragstart", (e) => {
			card.classList.add("dragging");
			e.dataTransfer.setData("text/plain", v.path);
			e.dataTransfer.effectAllowed = "move";
		});
		card.addEventListener("dragend", () => card.classList.remove("dragging"));
		gallery.appendChild(card);
	});
}

/**
 * Escapes a string for safe insertion into HTML attribute values or text nodes.
 * @param {string} s
 * @returns {string}
 */
function escHtml(s) {
	return s
		.replace(/&/g, "&amp;")
		.replace(/</g, "&lt;")
		.replace(/>/g, "&gt;")
		.replace(/"/g, "&quot;");
}

/**
 * Opens the video lightbox for the item at index idx within the current
 * filtered list.
 * @param {number} idx
 */
function openModal(idx) {
	currentIdx = idx;
	const v = filtered[idx];
	modalTitle.textContent = v.name;
	modalPath.textContent = v.folder || "/";
	modalVid.src = "/video?path=" + encodeURIComponent(v.path);
	modal.classList.add("open");
}
/** Closes the video lightbox and stops playback. */
function closeModal() {
	modal.classList.remove("open");
	modalVid.pause();
	modalVid.src = "";
}
closeBtn.addEventListener("click", closeModal);
modal.addEventListener("click", (e) => {
	if (e.target === modal) closeModal();
});
prevBtn.addEventListener("click", () => {
	if (currentIdx > 0) openModal(currentIdx - 1);
});
nextBtn.addEventListener("click", () => {
	if (currentIdx < filtered.length - 1) openModal(currentIdx + 1);
});
modalVid.addEventListener("ended", () => {
	if (currentIdx < filtered.length - 1) openModal(currentIdx + 1);
});
document.addEventListener("keydown", (e) => {
	if (!modal.classList.contains("open")) return;
	if (e.key === "Escape") closeModal();
	if (e.key === "ArrowLeft") prevBtn.click();
	if (e.key === "ArrowRight") nextBtn.click();
});

// context menu
const ctxMenu = document.getElementById("ctx-menu"),
	ctxMoveEl = document.getElementById("ctx-move");
let ctxVideo = null;
/**
 * Positions and shows the card context menu at the pointer location.
 * @param {MouseEvent} e
 * @param {{ name: string, path: string, folder: string }} v - The video the menu was opened for.
 */
function showCtxMenu(e, v) {
	ctxVideo = v;
	ctxMenu.style.left = Math.min(e.clientX, window.innerWidth - 180) + "px";
	ctxMenu.style.top = Math.min(e.clientY, window.innerHeight - 60) + "px";
	ctxMenu.classList.add("open");
}
document.addEventListener("click", () => ctxMenu.classList.remove("open"));
ctxMoveEl.addEventListener("click", () => {
	if (ctxVideo) openMoveModal(ctxVideo);
});

// move modal
const moveModal = document.getElementById("move-modal"),
	moveSubtitle = document.getElementById("move-subtitle"),
	folderOptions = document.getElementById("folder-options"),
	newFolderName = document.getElementById("new-folder-name"),
	addFolderBtn = document.getElementById("add-folder-btn"),
	moveCancelBtn = document.getElementById("move-cancel-btn"),
	moveConfirmBtn = document.getElementById("move-confirm-btn"),
	moveClose = document.getElementById("move-close");
let moveTarget = null,
	moveSelectedFolder = null;

/**
 * Opens the move-to-folder modal for a single video. When selectedPaths has
 * more than one entry the subtitle reflects the multi-selection count instead.
 * @param {{ name: string, path: string, folder: string }} v
 */
function openMoveModal(v) {
	moveTarget = v;
	moveSubtitle.textContent =
		selectedPaths.size > 1 ? selectedPaths.size + " items selected" : v.name;
	moveSelectedFolder = v.folder || "/";
	newFolderName.value = "";
	renderFolderOptions();
	moveModal.classList.add("open");
}
/** Rebuilds the folder option list inside the move modal, reflecting the
 *  current ALL_FOLDERS / ALL_VIDEOS state and highlighting moveSelectedFolder. */
function renderFolderOptions() {
	const folders = [
		...new Set([...ALL_FOLDERS, ...ALL_VIDEOS.map((v) => v.folder || "/")]),
	].sort();
	folderOptions.innerHTML = "";
	for (const f of folders) {
		const opt = document.createElement("div");
		opt.className =
			"folder-option" + (f === moveSelectedFolder ? " selected" : "");
		opt.innerHTML =
			'<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>' +
			(f === "/" ? "Root (/)" : escHtml(f));
		opt.addEventListener("click", () => {
			moveSelectedFolder = f;
			document
				.querySelectorAll(".folder-option")
				.forEach((o) => o.classList.remove("selected"));
			opt.classList.add("selected");
		});
		folderOptions.appendChild(opt);
	}
}
addFolderBtn.addEventListener("click", () => {
	const n = newFolderName.value.trim();
	if (!n) return;
	if (!ALL_FOLDERS.includes(n)) ALL_FOLDERS.push(n);
	newFolderName.value = "";
	moveSelectedFolder = n;
	renderFolderOptions();
});
newFolderName.addEventListener("keydown", (e) => {
	if (e.key === "Enter") addFolderBtn.click();
});
/** Closes the move-to-folder modal without taking any action. */
function closeMoveModal() {
	moveModal.classList.remove("open");
}
moveClose.addEventListener("click", closeMoveModal);
moveCancelBtn.addEventListener("click", closeMoveModal);
moveModal.addEventListener("click", (e) => {
	if (e.target === moveModal) closeMoveModal();
});
moveConfirmBtn.addEventListener("click", async () => {
	if (!moveSelectedFolder) return;
	const paths =
		selectedPaths.size > 0
			? [...selectedPaths]
			: moveTarget
				? [moveTarget.path]
				: [];
	if (!paths.length) return;
	let errs = 0;
	for (const p of paths) {
		const res = await fetch("/api/move", {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ path: p, dest_folder: moveSelectedFolder }),
		});
		if (!res.ok) errs++;
	}
	await refresh();
	if (errs) toast(errs + " move(s) failed", "error");
	else
		toast(
			"Moved " +
			(paths.length > 1 ? paths.length + " items" : "") +
			(paths.length > 1 ? " to " : "") +
			moveSelectedFolder,
			"success",
		);
	if (selectMode) exitSelectMode();
	closeMoveModal();
});
/**
 * Moves a single video file to destFolder via POST /api/move, then refreshes
 * the gallery. Shows a success or error toast on completion.
 * @param {string} srcPath - Relative path of the file to move.
 * @param {string} destFolder - Destination folder name relative to root.
 */
async function moveVideo(srcPath, destFolder) {
	const res = await fetch("/api/move", {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify({ path: srcPath, dest_folder: destFolder }),
	});
	if (res.ok) {
		toast("Moved to " + destFolder, "success");
		await refresh();
	} else {
		toast("Move failed: " + (await res.text()), "error");
	}
}

// upload modal
const uploadModal = document.getElementById("upload-modal"),
	dropZone = document.getElementById("drop-zone"),
	uploadInput = document.getElementById("upload-input"),
	uploadFolderSel = document.getElementById("upload-folder-select"),
	newFolderInput = document.getElementById("new-folder-input"),
	fileQueue = document.getElementById("file-queue"),
	uploadSummary = document.getElementById("upload-summary"),
	uploadStartBtn = document.getElementById("upload-start-btn"),
	uploadCancelBtn = document.getElementById("upload-cancel-btn"),
	uploadClose = document.getElementById("upload-close"),
	openUploadBtn = document.getElementById("open-upload-btn");
let uploadFiles = [];

/** Rebuilds the destination folder <select> in the upload modal from the
 *  current ALL_FOLDERS / ALL_VIDEOS state, preserving the previous selection. */
function populateUploadFolders() {
	const folders = [
		...new Set([
			"/",
			...ALL_FOLDERS,
			...ALL_VIDEOS.map((v) => v.folder || "/"),
		]),
	].sort();
	const prev = uploadFolderSel.value;
	uploadFolderSel.innerHTML = "";
	for (const f of folders) {
		const o = document.createElement("option");
		o.value = f;
		o.textContent = f === "/" ? "Root (/)" : f;
		uploadFolderSel.appendChild(o);
	}
	const no = document.createElement("option");
	no.value = "__new__";
	no.textContent = "+ new folder…";
	uploadFolderSel.appendChild(no);
	if (prev && [...uploadFolderSel.options].some((o) => o.value === prev))
		uploadFolderSel.value = prev;
}
uploadFolderSel.addEventListener("change", () => {
	if (uploadFolderSel.value === "__new__") {
		newFolderInput.classList.add("visible");
		newFolderInput.focus();
	} else {
		newFolderInput.classList.remove("visible");
		newFolderInput.value = "";
	}
});
openUploadBtn.addEventListener("click", () => {
	uploadFiles = [];
	renderFileQueue();
	uploadModal.classList.add("open");
});
function closeUploadModal() {
	uploadModal.classList.remove("open");
}
uploadClose.addEventListener("click", closeUploadModal);
uploadCancelBtn.addEventListener("click", closeUploadModal);
uploadModal.addEventListener("click", (e) => {
	if (e.target === uploadModal) closeUploadModal();
});
dropZone.addEventListener("click", () => uploadInput.click());
dropZone.addEventListener("dragover", (e) => {
	e.preventDefault();
	dropZone.classList.add("drag-active");
});
dropZone.addEventListener("dragleave", () =>
	dropZone.classList.remove("drag-active"),
);
dropZone.addEventListener("drop", (e) => {
	e.preventDefault();
	dropZone.classList.remove("drag-active");
	addFiles([...e.dataTransfer.files]);
});
uploadInput.addEventListener("change", () => {
	addFiles([...uploadInput.files]);
	uploadInput.value = "";
});

const ALLOWED_EXTS = new Set([
	"webm",
	"mp4",
	"mpv",
	"mkv",
	"mov",
	"avi",
	"m4v",
	"ogv",
]);
function addFiles(files) {
	for (const f of files) {
		const ext = f.name.split(".").pop().toLowerCase();
		const ok = ALLOWED_EXTS.has(ext);
		uploadFiles.push({
			file: f,
			id: Math.random().toString(36).slice(2),
			status: ok ? "pending" : "error",
			error: ok ? "" : "Unsupported type",
		});
	}
	renderFileQueue();
}
/** Re-renders the file queue list inside the upload modal from uploadFiles. */
function renderFileQueue() {
	fileQueue.innerHTML = "";
	for (const item of uploadFiles) {
		const el = document.createElement("div");
		el.className = "file-item";
		el.dataset.id = item.id;
		const mb = (item.file.size / 1048576).toFixed(1);
		el.innerHTML =
			'<span class="file-item-name">' +
			escHtml(item.file.name) +
			"</span>" +
			'<span class="file-item-size">' +
			mb +
			" MB</span>" +
			'<span class="file-item-status ' +
			item.status +
			'">' +
			(item.status === "error" ? item.error : item.status) +
			"</span>" +
			'<button class="file-remove" data-id="' +
			item.id +
			'">×</button>';
		fileQueue.appendChild(el);
	}
	fileQueue.querySelectorAll(".file-remove").forEach((btn) => {
		btn.addEventListener("click", () => {
			uploadFiles = uploadFiles.filter((f) => f.id !== btn.dataset.id);
			renderFileQueue();
			updateSummary();
		});
	});
	updateSummary();
}
/** Updates the upload summary line and enables/disables the Upload button
 *  based on the number of valid (non-error) files in uploadFiles. */
function updateSummary() {
	const valid = uploadFiles.filter((f) => f.status !== "error").length;
	uploadSummary.textContent = uploadFiles.length
		? valid +
		" of " +
		uploadFiles.length +
		" file" +
		(uploadFiles.length > 1 ? "s" : "") +
		" ready"
		: "";
	uploadStartBtn.disabled = valid === 0;
}
/**
 * Updates the status badge of a queued upload item.
 * @param {string} id - The item's random id assigned in uploadFiles.
 * @param {'pending'|'uploading'|'done'|'error'} cls - CSS modifier class.
 * @param {string} text - Display text for the badge.
 */
function setItemStatus(id, cls, text) {
	const el = fileQueue.querySelector(
		'[data-id="' + id + '"] .file-item-status',
	);
	if (el) {
		el.className = "file-item-status " + cls;
		el.textContent = text;
	}
}
uploadStartBtn.addEventListener("click", async () => {
	let dest = uploadFolderSel.value;
	if (dest === "__new__") dest = newFolderInput.value.trim() || "/";
	const todo = uploadFiles.filter((f) => f.status === "pending");
	uploadStartBtn.disabled = true;
	for (const item of todo) {
		setItemStatus(item.id, "uploading", "…");
		const fd = new FormData();
		fd.append("file", item.file);
		fd.append("folder", dest);
		try {
			const res = await fetch("/api/upload", { method: "POST", body: fd });
			const results = await res.json();
			const r = results[0];
			if (r && r.error) {
				setItemStatus(item.id, "error", r.error);
			} else {
				setItemStatus(item.id, "done", "✓ done");
				item.status = "done";
			}
		} catch (e) {
			setItemStatus(item.id, "error", "failed");
		}
	}
	await refresh();
	uploadStartBtn.disabled = false;
	toast("Upload complete", "success");
});

searchEl.addEventListener("input", render);
sortSel.addEventListener("change", render);
gridBtn.addEventListener("click", () => {
	gallery.classList.remove("list-view");
	gridBtn.classList.add("active");
	listBtn.classList.remove("active");
});
listBtn.addEventListener("click", () => {
	gallery.classList.add("list-view");
	listBtn.classList.add("active");
	gridBtn.classList.remove("active");
});

// select mode
/**
 * Toggles the selection state of a gallery card in select mode.
 * @param {string} path - Relative path of the video (used as the unique key).
 * @param {HTMLElement} card - The card DOM element to update visually.
 */
function toggleSelect(path, card) {
	if (selectedPaths.has(path)) {
		selectedPaths.delete(path);
		card.classList.remove("selected");
	} else {
		selectedPaths.add(path);
		card.classList.add("selected");
	}
	updateSelectBar();
}
/** Syncs the floating select bar's count label and visibility with selectedPaths. */
function updateSelectBar() {
	const n = selectedPaths.size;
	selectCount.textContent = n + " selected";
	selectBar.classList.toggle("visible", n > 0);
}
/** Exits select mode, clears all selections, and resets the select bar and button. */
function exitSelectMode() {
	selectMode = false;
	document.body.classList.remove("select-mode");
	selectBtn.classList.remove("active");
	selectedPaths.clear();
	document
		.querySelectorAll(".card.selected")
		.forEach((c) => c.classList.remove("selected"));
	selectBar.classList.remove("visible");
}
selectBtn.addEventListener("click", () => {
	if (selectMode) {
		exitSelectMode();
	} else {
		selectMode = true;
		document.body.classList.add("select-mode");
		selectBtn.classList.add("active");
	}
});
selectMoveBtn.addEventListener("click", () => {
	if (!selectedPaths.size) return;
	const first = filtered.find((v) => selectedPaths.has(v.path)) || filtered[0];
	openMoveModal(first);
});
selectAllBtn.addEventListener("click", () => {
	filtered.forEach((v) => selectedPaths.add(v.path));
	document.querySelectorAll(".card").forEach((c) => {
		if (filtered.some((v) => v.path === c.dataset.path))
			c.classList.add("selected");
	});
	updateSelectBar();
});
selectClearBtn.addEventListener("click", () => {
	selectedPaths.clear();
	document
		.querySelectorAll(".card.selected")
		.forEach((c) => c.classList.remove("selected"));
	updateSelectBar();
});

// new folder (sidebar)
newFolderBtn.addEventListener("click", () => {
	sidebarFolderForm.classList.add("visible");
	sidebarFolderInput.value = "";
	sidebarFolderInput.focus();
});
/** Hides the inline new-folder input in the sidebar and clears its value. */
function hideSidebarFolderForm() {
	sidebarFolderForm.classList.remove("visible");
	sidebarFolderInput.value = "";
}
sidebarFolderCancel.addEventListener("click", hideSidebarFolderForm);
/** Reads the sidebar folder input, POSTs to /api/mkdir, then refreshes the
 *  gallery and hides the form. Shows a toast on success or failure. */
async function createFolder() {
	const name = sidebarFolderInput.value.trim();
	if (!name) return;
	const res = await fetch("/api/mkdir", {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify({ folder: name }),
	});
	if (res.ok) {
		hideSidebarFolderForm();
		toast("Created " + name, "success");
		await refresh();
	} else {
		toast("Failed: " + (await res.text()), "error");
	}
}
sidebarFolderConfirm.addEventListener("click", createFolder);
sidebarFolderInput.addEventListener("keydown", (e) => {
	if (e.key === "Enter") createFolder();
	if (e.key === "Escape") hideSidebarFolderForm();
});

/**
 * Displays a transient notification at the bottom-right of the screen.
 * @param {string} msg - Message text to display.
 * @param {'success'|'error'|''} type - CSS modifier applied to the toast element.
 */
function toast(msg, type = "") {
	const el = document.createElement("div");
	el.className = "toast " + type;
	el.textContent = msg;
	document.body.appendChild(el);
	setTimeout(() => el.remove(), 3000);
}

init();
