const gallery = document.getElementById("gallery"),
	folderList = document.getElementById("folder-list"),
	tagList = document.getElementById("tag-list"),
	searchEl = document.getElementById("search"),
	statsEl = document.getElementById("stats"),
	sortSel = document.getElementById("sort-select"),
	gridBtn = document.getElementById("grid-btn"),
	listBtn = document.getElementById("list-btn"),
	favoritesOnlyToggle = document.getElementById("favorites-only-toggle"),
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
	selectTagBtn = document.getElementById("select-tag-btn"),
	selectMoveBtn = document.getElementById("select-move-btn"),
	selectAllBtn = document.getElementById("select-all-btn"),
	selectClearBtn = document.getElementById("select-clear-btn"),
	newFolderBtn = document.getElementById("new-folder-btn"),
	sidebarFolderForm = document.getElementById("sidebar-folder-form"),
	sidebarFolderInput = document.getElementById("sidebar-folder-input"),
	sidebarFolderConfirm = document.getElementById("sidebar-folder-confirm"),
	sidebarFolderCancel = document.getElementById("sidebar-folder-cancel"),
	newTagBtn = document.getElementById("new-tag-btn"),
	sidebarTagForm = document.getElementById("sidebar-tag-form"),
	sidebarTagInput = document.getElementById("sidebar-tag-input"),
	sidebarTagConfirm = document.getElementById("sidebar-tag-confirm"),
	sidebarTagCancel = document.getElementById("sidebar-tag-cancel"),
	newCollectionBtn = document.getElementById("new-collection-btn"),
	sidebarCollectionForm = document.getElementById("sidebar-collection-form"),
	sidebarCollectionInput = document.getElementById("sidebar-collection-input"),
	sidebarCollectionConfirm = document.getElementById("sidebar-collection-confirm"),
	sidebarCollectionCancel = document.getElementById("sidebar-collection-cancel"),
	collectionList = document.getElementById("collection-list"),
	collectionsCollapseBtn = document.getElementById("collections-collapse-btn"),
	tagsCollapseBtn = document.getElementById("tags-collapse-btn"),
	foldersCollapseBtn = document.getElementById("folders-collapse-btn"),
	tagsSection = tagsCollapseBtn.closest(".sidebar-section"),
	foldersSection = foldersCollapseBtn.closest(".sidebar-section"),
	collectionsSection = collectionsCollapseBtn.closest(".sidebar-section"),
	dupBanner = document.getElementById("dup-banner"),
	dupBannerText = document.getElementById("dup-banner-text"),
	dupReviewBtn = document.getElementById("dup-review-btn"),
	dupDismissBtn = document.getElementById("dup-dismiss-btn"),
	dupModal = document.getElementById("dup-modal"),
	dupSubtitle = document.getElementById("dup-subtitle"),
	dupGroupsEl = document.getElementById("dup-groups"),
	dupClose = document.getElementById("dup-close"),
	dupCancelBtn = document.getElementById("dup-cancel-btn"),
	dupResolveAllBtn = document.getElementById("dup-resolve-all-btn"),
	tagModal = document.getElementById("tag-modal"),
	tagSubtitle = document.getElementById("tag-subtitle"),
	tagOptions = document.getElementById("tag-options"),
	tagClose = document.getElementById("tag-close"),
	tagCancelBtn = document.getElementById("tag-cancel-btn"),
	tagNewInput = document.getElementById("tag-new-input"),
	tagNewBtn = document.getElementById("tag-new-btn"),
	selectCollectionBtn = document.getElementById("select-collection-btn"),
	collectionPickerModal = document.getElementById("collection-picker-modal"),
	collectionPickerSubtitle = document.getElementById("collection-picker-subtitle"),
	collectionPickerOptions = document.getElementById("collection-picker-options"),
	collectionPickerClose = document.getElementById("collection-picker-close"),
	collectionPickerCancelBtn = document.getElementById("collection-picker-cancel-btn"),
	collectionPickerConfirmBtn = document.getElementById("collection-picker-confirm-btn"),
	collectionPickerNewInput = document.getElementById("collection-picker-new-input"),
	collectionPickerNewBtn = document.getElementById("collection-picker-new-btn"),
	watchCollectionModal = document.getElementById("watch-collection-modal"),
	watchCollectionTitle = document.getElementById("watch-collection-title"),
	watchCollectionSubtitle = document.getElementById("watch-collection-subtitle"),
	watchCollectionBody = document.getElementById("watch-collection-body"),
	watchCollectionClose = document.getElementById("watch-collection-close"),
	watchCollectionDeleteBtn = document.getElementById("watch-collection-delete-btn"),
	watchCollectionRenameBtn = document.getElementById("watch-collection-rename-btn");

let ALL_VIDEOS = [],
	ALL_FOLDERS = [],
	FOLDER_META = {},
	ALL_TAGS = [],
	TAG_ASSIGNMENTS = {},
	TAG_MAP = {},
	ALL_COLLECTIONS = [],
	COLLECTION_MAP = {},
	activeTagFilters = new Set(),
	filtered = [],
	activeFolder = "__all__",
	currentIdx = -1;
let selectMode = false,
	selectedPaths = new Set();
let DUPLICATE_GROUPS = [],
	duplicateBannerDismissed = false,
	autoOpenedDuplicates = false;
let showFavoritesOnly = false;
let tagTargetPaths = [];
let activeCollectionID = "";
let collectionPickerTargetHashes = [];
const DEFAULT_LOOP_ALL_VIDEOS = true;

function applyDefaultVideoFlags(videoEl) {
	if (!videoEl) return;
	videoEl.loop = DEFAULT_LOOP_ALL_VIDEOS;
}

/**
 * Populates ALL_FOLDERS (string[]) and FOLDER_META (name → folderInfo) from
 * the raw /api/folders response array.
 * @param {{ name: string, has_other_files: boolean }[]} data
 */
function parseFolders(data) {
	ALL_FOLDERS = data.map((f) => f.name);
	FOLDER_META = Object.fromEntries(data.map((f) => [f.name, f]));
}

function parseTags(data) {
	ALL_TAGS = Array.isArray(data.tags) ? data.tags : [];
	TAG_ASSIGNMENTS =
		data && typeof data.assignments === "object" && data.assignments
			? data.assignments
			: {};
	TAG_MAP = Object.fromEntries(ALL_TAGS.map((t) => [t.id, t]));
}

function parseCollections(data) {
	const raw =
		data && Array.isArray(data.collections) ? data.collections : [];
	ALL_COLLECTIONS = raw.map((c) => ({
		id: c.id,
		name: c.name,
		video_hashes: Array.isArray(c.video_hashes) ? c.video_hashes : [],
		created_at: c.created_at || "",
		updated_at: c.updated_at || "",
	}));
	COLLECTION_MAP = Object.fromEntries(ALL_COLLECTIONS.map((c) => [c.id, c]));
	if (activeCollectionID && !COLLECTION_MAP[activeCollectionID]) {
		activeCollectionID = "";
	}
}

function tagCountByID(tagID) {
	return ALL_VIDEOS.filter((v) => Array.isArray(v.tags) && v.tags.includes(tagID)).length;
}

function renderTagChip(tag) {
	return (
		'<span class="tag-chip" style="--tag-color:' +
		escHtml(tag.color || "#888888") +
		'">' +
		escHtml(tag.name) +
		"</span>"
	);
}

function renderVideoTagChips(tagIDs) {
	if (!Array.isArray(tagIDs) || !tagIDs.length) return "";
	const chips = tagIDs
		.map((id) => TAG_MAP[id])
		.filter(Boolean)
		.map((tag) => renderTagChip(tag))
		.join("");
	if (!chips) return "";
	return '<div class="card-tags">' + chips + "</div>";
}



async function init() {
	const [vr, fr, tr, cr] = await Promise.all([
		fetch("/api/videos"),
		fetch("/api/folders"),
		fetch("/api/tags"),
		fetch("/api/collections"),
	]);
	ALL_VIDEOS = await vr.json();
	parseFolders(await fr.json());
	parseTags(await tr.json());
	parseCollections(await cr.json());
	computeDuplicateGroups();
	buildTagNav();
	buildFolderNav();
	buildCollectionNav();
	render();
	updateDuplicateBanner(true);
	populateUploadFolders();
	applyDefaultVideoFlags(modalVid);
}
async function refresh() {
	const [vr, fr, tr, cr] = await Promise.all([
		fetch("/api/videos"),
		fetch("/api/folders"),
		fetch("/api/tags"),
		fetch("/api/collections"),
	]);
	ALL_VIDEOS = await vr.json();
	parseFolders(await fr.json());
	parseTags(await tr.json());
	parseCollections(await cr.json());
	computeDuplicateGroups();
	buildTagNav();
	buildFolderNav();
	buildCollectionNav();
	render();
	updateDuplicateBanner(false);
	populateUploadFolders();
}

function computeDuplicateGroups() {
	const byHash = new Map();
	for (const v of ALL_VIDEOS) {
		if (!v.hash) continue;
		const members = byHash.get(v.hash) || [];
		members.push(v);
		byHash.set(v.hash, members);
	}

	DUPLICATE_GROUPS = [...byHash.entries()]
		.filter(([, members]) => members.length > 1)
		.map(([hash, members]) => {
			const sorted = [...members].sort(
				(a, b) =>
					new Date(b.modified || 0).getTime() -
						new Date(a.modified || 0).getTime() ||
					Number(b.size || 0) - Number(a.size || 0) ||
					a.path.localeCompare(b.path),
			);
			return {
				hash,
				members: sorted,
				keepPath: sorted[0]?.path || "",
			};
		}).sort((a, b) => b.members.length - a.members.length || a.hash.localeCompare(b.hash));
}

function updateDuplicateBanner(mayAutoOpen) {
	const groups = DUPLICATE_GROUPS.length;
	const files = DUPLICATE_GROUPS.reduce((acc, g) => acc + g.members.length, 0);
	if (!groups) {
		dupBanner.classList.remove("visible");
		duplicateBannerDismissed = false;
		autoOpenedDuplicates = false;
		return;
	}

	dupBannerText.textContent =
		groups +
		" duplicate group" +
		(groups === 1 ? "" : "s") +
		" found (" +
		files +
		" files).";
	const shouldShow = !duplicateBannerDismissed || dupModal.classList.contains("open");
	dupBanner.classList.toggle("visible", shouldShow);

	if (mayAutoOpen && !autoOpenedDuplicates && groups > 0) {
		openDuplicateModal();
		autoOpenedDuplicates = true;
	}
}

function openDuplicateModal() {
	renderDuplicateGroups();
	dupModal.classList.add("open");
	dupBanner.classList.add("visible");
}

function closeDuplicateModal() {
	dupModal.classList.remove("open");
	updateDuplicateBanner(false);
}

function renderDuplicateGroups() {
	const groups = DUPLICATE_GROUPS.length;
	if (!groups) {
		dupSubtitle.textContent = "No duplicates right now.";
		dupGroupsEl.innerHTML =
			'<div class="dup-empty">Everything is unique. Nice and tidy.</div>';
		dupResolveAllBtn.disabled = true;
		return;
	}

	dupSubtitle.textContent =
		groups + " group" + (groups === 1 ? "" : "s") + " need review";
	dupResolveAllBtn.disabled = false;
	dupGroupsEl.innerHTML = "";

	DUPLICATE_GROUPS.forEach((group, idx) => {
		const card = document.createElement("section");
		card.className = "dup-group";
		card.dataset.hash = group.hash;

		const header = document.createElement("div");
		header.className = "dup-group-header";
		header.innerHTML =
			'<div class="dup-group-title">Group ' +
			(idx + 1) +
			" (" +
			group.members.length +
			' copies)</div><button class="btn dup-resolve-btn">Delete non-kept</button>';
		header.querySelector(".dup-resolve-btn").addEventListener("click", async () => {
			await resolveDuplicateGroup(group.hash);
		});
		card.appendChild(header);

		const list = document.createElement("div");
		list.className = "dup-items";
		for (const v of group.members) {
			const isKeep = group.keepPath === v.path;
			const row = document.createElement("label");
			row.className = "dup-item" + (isKeep ? " keep" : "");
			row.innerHTML =
				'<input type="radio" name="dup-keep-' +
				escHtml(group.hash) +
				'" value="' +
				escHtml(v.path) +
				'" ' +
				(isKeep ? "checked" : "") +
				">" +
				'<div class="dup-item-meta"><span class="dup-item-name">' +
				escHtml(v.name) +
				"</span>" +
				'<span class="dup-item-path">' +
				escHtml(v.folder || "/") +
				" • " +
				escHtml(formatBytes(v.size || 0)) +
				" • " +
				escHtml(formatModified(v.modified)) +
				"</span></div>";
			const radio = row.querySelector("input");
			radio.addEventListener("change", () => {
				group.keepPath = v.path;
				list.querySelectorAll(".dup-item").forEach((el) => el.classList.remove("keep"));
				row.classList.add("keep");
			});
			list.appendChild(row);
		}
		card.appendChild(list);
		dupGroupsEl.appendChild(card);
	});
}

/**
 * 
 */
function buildTagNav() {
	tagList.innerHTML = "";
	for (const tag of ALL_TAGS) {
		const count = tagCountByID(tag.id);
		const btn = document.createElement("button");
		btn.className = "tag-btn" + (activeTagFilters.has(tag.id) ? " active" : "");
		btn.innerHTML =
			renderTagChip(tag) +
			'<span class="count">' +
			count +
			'</span><span class="tag-remove-btn" title="Delete tag">✕</span>';
		btn.addEventListener("click", () => {
			if (activeTagFilters.has(tag.id)) activeTagFilters.delete(tag.id);
			else activeTagFilters.add(tag.id);
			buildTagNav();
			render();
		});
		btn.querySelector(".tag-remove-btn").addEventListener("click", async (e) => {
			e.stopPropagation();
			if (!confirm('Delete tag "' + tag.name + '"?')) return;
			const ok = await deleteTag(tag.id);
			if (!ok) {
				toast("Tag delete failed", "error");
				return;
			}
			activeTagFilters.delete(tag.id);
			await refresh();
		});
		tagList.appendChild(btn);
	}
}

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

function countCollectionResolvedVideos(collection) {
	const hashes = new Set(collection.video_hashes || []);
	return ALL_VIDEOS.filter((v) => hashes.has(v.hash)).length;
}

function buildCollectionNav() {
	collectionList.innerHTML = "";
	for (const collection of ALL_COLLECTIONS) {
		const btn = document.createElement("button");
		btn.className =
			"collection-btn" + (collection.id === activeCollectionID ? " active" : "");
		btn.innerHTML =
			'<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/></svg>' +
			"<span>" +
			escHtml(collection.name) +
			"</span>" +
			'<span class="count">' +
			countCollectionResolvedVideos(collection) +
			'</span><button class="collection-rename-btn" title="Rename collection">✎</button><button class="collection-remove-btn" title="Delete collection">✕</button>';

		btn.addEventListener("click", () => openWatchCollection(collection.id));
		btn.querySelector(".collection-remove-btn").addEventListener("click", async (e) => {
			e.stopPropagation();
			if (!confirm('Delete collection "' + collection.name + '"?')) return;
			const ok = await deleteCollection(collection.id);
			if (!ok) {
				toast("Collection delete failed", "error");
				return;
			}
			if (activeCollectionID === collection.id) closeWatchCollection();
			await refresh();
			toast("Deleted " + collection.name, "success");
		});
		btn.querySelector(".collection-rename-btn").addEventListener("click", async (e) => {
			e.stopPropagation();
			const nextName = prompt("Rename collection", collection.name);
			if (!nextName || !nextName.trim()) return;
			const ok = await renameCollection(collection.id, nextName.trim());
			if (!ok) {
				toast("Collection rename failed", "error");
				return;
			}
			await refresh();
			if (activeCollectionID === collection.id) openWatchCollection(collection.id);
		});
		collectionList.appendChild(btn);
	}
}

function render() {
	const q = searchEl.value.toLowerCase(),
		sort = sortSel.value;
	filtered = ALL_VIDEOS.filter((v) => {
		const inF = activeFolder === "__all__" || (v.folder || "/") == activeFolder;
		const inFav = !showFavoritesOnly || Boolean(v.is_favorite);
		const inT =
			activeTagFilters.size === 0 ||
			[...activeTagFilters].every((tagID) => (v.tags || []).includes(tagID));
		const inS =
			!q ||
			v.name.toLowerCase().includes(q) ||
			(v.folder || "").toLowerCase().includes(q);
		return inF && inFav && inT && inS;
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
		if (sort === "modified")
			return (
				new Date(a.modified || 0).getTime() - new Date(b.modified || 0).getTime() ||
				a.name.localeCompare(b.name)
			);
		if (sort === "modified-desc")
			return (
				new Date(b.modified || 0).getTime() - new Date(a.modified || 0).getTime() ||
				a.name.localeCompare(b.name)
			);
		if (sort === "size")
			return (Number(a.size || 0) - Number(b.size || 0)) || a.name.localeCompare(b.name);
		if (sort === "size-desc")
			return (Number(b.size || 0) - Number(a.size || 0)) || a.name.localeCompare(b.name);
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
			'<button class="fav-toggle' +
			(v.is_favorite ? " active" : "") +
			'" title="' +
			(v.is_favorite ? "Remove favorite" : "Mark favorite") +
			'" aria-label="' +
			(v.is_favorite ? "Remove favorite" : "Mark favorite") +
			'">★</button>' +
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
			"</span>" +
			'<span class="card-path">' +
			escHtml(formatBytes(v.size || 0)) +
			" • " +
			escHtml(formatModified(v.modified)) +
			"</span>" +
			renderVideoTagChips(v.tags || []) +
			"</div>";
		const vid = card.querySelector("video");
		applyDefaultVideoFlags(vid);
		const favBtn = card.querySelector(".fav-toggle");
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
		favBtn.addEventListener("click", async (e) => {
			e.preventDefault();
			e.stopPropagation();
			const nextValue = !v.is_favorite;
			const ok = await setFavorite(v.hash, nextValue);
			if (!ok) {
				toast("Favorite update failed", "error");
				return;
			}
			v.is_favorite = nextValue;
			favBtn.classList.toggle("active", nextValue);
			favBtn.title = nextValue ? "Remove favorite" : "Mark favorite";
			favBtn.setAttribute(
				"aria-label",
				nextValue ? "Remove favorite" : "Mark favorite",
			);
		});
		card.addEventListener("click", (e) => {
			if (e.target.closest(".drag-handle")) return;
			if (e.target.closest(".fav-toggle")) return;
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
 * Formats bytes as a human-readable string.
 * @param {number} bytes
 * @returns {string}
 */
function formatBytes(bytes) {
	const n = Number(bytes) || 0;
	if (n < 1024) return n + " B";
	const units = ["KB", "MB", "GB", "TB"];
	let value = n;
	let idx = -1;
	while (value >= 1024 && idx < units.length - 1) {
		value /= 1024;
		idx++;
	}
	return value.toFixed(value >= 10 ? 0 : 1) + " " + units[idx];
}

/**
 * Formats an ISO date/time string for card display.
 * @param {string} iso
 * @returns {string}
 */
function formatModified(iso) {
	if (!iso) return "unknown";
	const d = new Date(iso);
	if (Number.isNaN(d.getTime())) return "unknown";
	return d.toLocaleString();
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
	ctxTagsEl = document.getElementById("ctx-tags"),
	ctxMoveEl = document.getElementById("ctx-move"),
	ctxDeleteEl = document.getElementById("ctx-delete");
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
ctxTagsEl.addEventListener("click", () => {
	if (ctxVideo) openTagModal(ctxVideo);
});
ctxMoveEl.addEventListener("click", () => {
	if (ctxVideo) openMoveModal(ctxVideo);
});
ctxDeleteEl.addEventListener("click", async () => {
	if (!ctxVideo) return;
	const ok = confirm(
		'Delete "' +
			ctxVideo.name +
			'"?\n\nThis removes the file from disk and cannot be undone.',
	);
	if (!ok) return;
	const res = await deleteVideoPath(ctxVideo.path);
	if (res.ok) {
		toast("Deleted " + ctxVideo.name, "success");
		await refresh();
	} else {
		toast("Delete failed: " + res.message, "error");
	}
});

function resolveTagTargets() {
	return tagTargetPaths
		.map((path) => ALL_VIDEOS.find((v) => v.path === path))
		.filter(Boolean);
}

function openTagModal(videoOrVideos) {
	const videos = Array.isArray(videoOrVideos) ? videoOrVideos : [videoOrVideos];
	tagTargetPaths = videos.map((v) => v.path).filter(Boolean);
	tagSubtitle.textContent =
		videos.length > 1 ? videos.length + " videos selected" : videos[0]?.name || "";
	tagNewInput.value = "";
	renderTagOptions();
	tagModal.classList.add("open");
}

function closeTagModal() {
	tagModal.classList.remove("open");
	tagTargetPaths = [];
}

function renderTagOptions() {
	tagOptions.innerHTML = "";
	const targets = resolveTagTargets();
	if (!targets.length) return;
	for (const tag of ALL_TAGS) {
		const selectedCount = targets.filter((v) => (v.tags || []).includes(tag.id)).length;
		const allSelected = selectedCount === targets.length;
		const partiallySelected = selectedCount > 0 && !allSelected;
		const row = document.createElement("label");
		row.className = "tag-option";
		row.innerHTML =
			'<input type="checkbox" ' +
			(allSelected ? "checked" : "") +
			'><span class="tag-option-name">' +
			renderTagChip(tag) +
			'</span><button class="tag-option-remove" type="button">delete</button>';
		const box = row.querySelector("input");
		box.indeterminate = partiallySelected;
		box.addEventListener("change", async () => {
			const liveTargets = resolveTagTargets();
			const hashes = [...new Set(liveTargets.map((v) => v.hash).filter(Boolean))];
			let ok = true;
			for (const hash of hashes) {
				const applied = await setTagAssignment(hash, tag.id, box.checked);
				if (!applied) {
					ok = false;
					break;
				}
			}
			if (!ok) {
				toast("Tag update failed", "error");
				await refresh();
				renderTagOptions();
				return;
			}
			await refresh();
			renderTagOptions();
		});
		row.querySelector(".tag-option-remove").addEventListener("click", async (e) => {
			e.preventDefault();
			e.stopPropagation();
			if (!confirm('Delete tag "' + tag.name + '"?')) return;
			const ok = await deleteTag(tag.id);
			if (!ok) {
				toast("Tag delete failed", "error");
				return;
			}
			activeTagFilters.delete(tag.id);
			await refresh();
			renderTagOptions();
		});
		tagOptions.appendChild(row);
	}
}

tagNewBtn.addEventListener("click", async () => {
	const name = tagNewInput.value.trim();
	if (!name) return;
	const created = await createTag(name);
	if (!created) {
		toast("Tag create failed", "error");
		return;
	}
	tagNewInput.value = "";
	await refresh();
	renderTagOptions();
});
tagNewInput.addEventListener("keydown", (e) => {
	if (e.key === "Enter") tagNewBtn.click();
});
tagClose.addEventListener("click", closeTagModal);
tagCancelBtn.addEventListener("click", closeTagModal);
tagModal.addEventListener("click", (e) => {
	if (e.target === tagModal) closeTagModal();
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

function resolveCollectionVideos(collection) {
	const byHash = new Map();
	for (const v of ALL_VIDEOS) {
		if (!v.hash || byHash.has(v.hash)) continue;
		byHash.set(v.hash, v);
	}
	const ordered = [];
	const missing = [];
	for (const hash of collection.video_hashes || []) {
		const v = byHash.get(hash);
		if (v) ordered.push(v);
		else missing.push(hash);
	}
	return { ordered, missing };
}

function openWatchCollection(collectionID) {
	const collection = COLLECTION_MAP[collectionID];
	if (!collection) return;
	activeCollectionID = collectionID;
	buildCollectionNav();
	const { ordered, missing } = resolveCollectionVideos(collection);
	watchCollectionTitle.textContent = collection.name;
	watchCollectionSubtitle.textContent =
		ordered.length +
		" video" +
		(ordered.length === 1 ? "" : "s") +
		(missing.length ? " • " + missing.length + " missing" : "");
	watchCollectionBody.innerHTML = "";
	if (!ordered.length) {
		watchCollectionBody.innerHTML =
			'<div class="watch-collection-empty">No playable videos in this collection yet.</div>';
	} else {
		const grid = document.createElement("div");
		grid.className = "watch-grid";
		for (const v of ordered) {
			const tile = document.createElement("div");
			tile.className = "watch-tile";
			tile.innerHTML =
				'<video controls loop preload="metadata" src="/video?path=' +
				encodeURIComponent(v.path) +
				'"></video><div class="watch-tile-meta"><span class="watch-tile-name" title="' +
				escHtml(v.name) +
				'">' +
				escHtml(v.name) +
				'</span><span class="watch-tile-path">' +
				escHtml(v.folder || "/") +
				'</span><button class="watch-tile-remove" title="Remove from collection">✕</button></div>';
			tile.querySelector(".watch-tile-remove").addEventListener("click", async () => {
				const ok = await setCollectionVideo(collection.id, v.hash, false);
				if (!ok) {
					toast("Remove failed", "error");
					return;
				}
				await refresh();
				openWatchCollection(collection.id);
			});
			grid.appendChild(tile);
		}
		watchCollectionBody.appendChild(grid);
	}
	watchCollectionModal.classList.add("open");
}

function closeWatchCollection() {
	watchCollectionModal.classList.remove("open");
	activeCollectionID = "";
	buildCollectionNav();
	watchCollectionBody
		.querySelectorAll("video")
		.forEach((v) => {
			v.pause();
			v.src = "";
		});
}

function openCollectionPicker(hashes) {
	collectionPickerTargetHashes = hashes.filter(Boolean);
	collectionPickerSelectedID = ALL_COLLECTIONS[0]?.id || "";
	collectionPickerSubtitle.textContent =
		collectionPickerTargetHashes.length +
		" video" +
		(collectionPickerTargetHashes.length === 1 ? "" : "s") +
		" selected";
	renderCollectionPickerOptions();
	collectionPickerModal.classList.add("open");
}

function closeCollectionPicker() {
	collectionPickerModal.classList.remove("open");
	collectionPickerTargetHashes = [];
	collectionPickerSelectedID = "";
}

let collectionPickerSelectedID = "";
function renderCollectionPickerOptions() {
	collectionPickerOptions.innerHTML = "";
	for (const collection of ALL_COLLECTIONS) {
		const option = document.createElement("div");
		option.className =
			"collection-picker-option" +
			(collection.id === collectionPickerSelectedID ? " selected" : "");
		option.innerHTML =
			'<span>' +
			escHtml(collection.name) +
			'</span><span class="count">' +
			countCollectionResolvedVideos(collection) +
			"</span>";
		option.addEventListener("click", () => {
			collectionPickerSelectedID = collection.id;
			renderCollectionPickerOptions();
		});
		collectionPickerOptions.appendChild(option);
	}
}

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
favoritesOnlyToggle.addEventListener("change", () => {
	showFavoritesOnly = favoritesOnlyToggle.checked;
	render();
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
selectTagBtn.addEventListener("click", () => {
	if (!selectedPaths.size) return;
	const targets = filtered.filter((v) => selectedPaths.has(v.path));
	if (!targets.length) return;
	openTagModal(targets);
});
selectCollectionBtn.addEventListener("click", () => {
	if (!selectedPaths.size) return;
	const targets = filtered.filter((v) => selectedPaths.has(v.path));
	const hashes = [...new Set(targets.map((v) => v.hash).filter(Boolean))];
	if (!hashes.length) {
		toast("No valid hashes found in selection", "error");
		return;
	}
	openCollectionPicker(hashes);
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

// tags (sidebar)
newTagBtn.addEventListener("click", () => {
	sidebarTagForm.classList.add("visible");
	sidebarTagInput.value = "";
	sidebarTagInput.focus();
});
function hideSidebarTagForm() {
	sidebarTagForm.classList.remove("visible");
	sidebarTagInput.value = "";
}
sidebarTagCancel.addEventListener("click", hideSidebarTagForm);
async function createSidebarTag() {
	const name = sidebarTagInput.value.trim();
	if (!name) return;
	const created = await createTag(name);
	if (!created) {
		toast("Failed: tag already exists or request failed", "error");
		return;
	}
	hideSidebarTagForm();
	toast("Created tag " + name, "success");
	await refresh();
}
sidebarTagConfirm.addEventListener("click", createSidebarTag);
sidebarTagInput.addEventListener("keydown", (e) => {
	if (e.key === "Enter") createSidebarTag();
	if (e.key === "Escape") hideSidebarTagForm();
});

// collections (sidebar)
newCollectionBtn.addEventListener("click", () => {
	sidebarCollectionForm.classList.add("visible");
	sidebarCollectionInput.value = "";
	sidebarCollectionInput.focus();
});
function hideSidebarCollectionForm() {
	sidebarCollectionForm.classList.remove("visible");
	sidebarCollectionInput.value = "";
}
sidebarCollectionCancel.addEventListener("click", hideSidebarCollectionForm);
async function createSidebarCollection() {
	const name = sidebarCollectionInput.value.trim();
	if (!name) return;
	const created = await createCollection(name);
	if (!created) {
		toast("Failed: collection already exists or request failed", "error");
		return;
	}
	hideSidebarCollectionForm();
	toast("Created collection " + name, "success");
	await refresh();
}
sidebarCollectionConfirm.addEventListener("click", createSidebarCollection);
sidebarCollectionInput.addEventListener("keydown", (e) => {
	if (e.key === "Enter") createSidebarCollection();
	if (e.key === "Escape") hideSidebarCollectionForm();
});

tagsCollapseBtn.addEventListener("click", () => {
	tagsSection.classList.toggle("collapsed");
});
collectionsCollapseBtn.addEventListener("click", () => {
	collectionsSection.classList.toggle("collapsed");
});
foldersCollapseBtn.addEventListener("click", () => {
	foldersSection.classList.toggle("collapsed");
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

async function deleteVideoPath(path) {
	try {
		const res = await fetch("/api/delete", {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ path }),
		});
		if (res.ok) return { ok: true, message: "" };
		return { ok: false, message: await res.text() };
	} catch (e) {
		return { ok: false, message: "request failed" };
	}
}

async function setFavorite(hash, favorite) {
	if (!hash) return false;
	try {
		const res = await fetch("/api/favorites/set", {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ hash, favorite }),
		});
		return res.ok;
	} catch (e) {
		return false;
	}
}

async function createTag(name) {
	try {
		const res = await fetch("/api/tags/create", {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ name }),
		});
		if (!res.ok) return null;
		return await res.json();
	} catch (e) {
		return null;
	}
}

async function deleteTag(tagID) {
	try {
		const res = await fetch("/api/tags/delete", {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ tag_id: tagID }),
		});
		return res.ok;
	} catch (e) {
		return false;
	}
}

async function setTagAssignment(hash, tagID, assigned) {
	try {
		const res = await fetch("/api/tags/assign", {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ hash, tag_id: tagID, assigned }),
		});
		return res.ok;
	} catch (e) {
		return false;
	}
}

async function createCollection(name) {
	try {
		const res = await fetch("/api/collections/create", {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ name }),
		});
		if (!res.ok) return null;
		return await res.json();
	} catch (e) {
		return null;
	}
}

async function renameCollection(id, name) {
	try {
		const res = await fetch("/api/collections/rename", {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ id, name }),
		});
		return res.ok;
	} catch (e) {
		return false;
	}
}

async function deleteCollection(id) {
	try {
		const res = await fetch("/api/collections/delete", {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ id }),
		});
		return res.ok;
	} catch (e) {
		return false;
	}
}

async function setCollectionVideo(id, hash, assigned) {
	try {
		const res = await fetch("/api/collections/videos/set", {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ id, hash, assigned }),
		});
		return res.ok;
	} catch (e) {
		return false;
	}
}

async function bulkAddCollectionVideos(id, hashes) {
	try {
		const res = await fetch("/api/collections/videos/bulk", {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ id, hashes }),
		});
		return res.ok;
	} catch (e) {
		return false;
	}
}

async function resolveDuplicateGroup(hash) {
	const group = DUPLICATE_GROUPS.find((g) => g.hash === hash);
	if (!group || group.members.length < 2) return;
	const toDelete = group.members.filter((m) => m.path !== group.keepPath);
	if (!toDelete.length) return;
	const ok = confirm(
		"Delete " +
			toDelete.length +
			' duplicate file(s) and keep "' +
			(group.members.find((m) => m.path === group.keepPath)?.name || "selected copy") +
			'"?',
	);
	if (!ok) return;

	let fail = 0;
	for (const item of toDelete) {
		const res = await deleteVideoPath(item.path);
		if (!res.ok) fail++;
	}
	await refresh();
	renderDuplicateGroups();
	if (fail) toast(fail + " duplicate delete(s) failed", "error");
	else toast("Duplicate group resolved", "success");
}

async function resolveAllDuplicateGroups() {
	const actions = DUPLICATE_GROUPS.map((g) => ({
		hash: g.hash,
		toDelete: g.members.filter((m) => m.path !== g.keepPath),
	}));
	const totalDeletes = actions.reduce((acc, g) => acc + g.toDelete.length, 0);
	if (!totalDeletes) return;

	const ok = confirm(
		"Resolve all groups by deleting " +
			totalDeletes +
			" file(s) and keeping your selected copy in each group?",
	);
	if (!ok) return;

	let fail = 0;
	for (const action of actions) {
		for (const item of action.toDelete) {
			const res = await deleteVideoPath(item.path);
			if (!res.ok) fail++;
		}
	}
	await refresh();
	renderDuplicateGroups();
	if (fail) toast(fail + " duplicate delete(s) failed", "error");
	else toast("All duplicate groups resolved", "success");
}

dupReviewBtn.addEventListener("click", openDuplicateModal);
dupDismissBtn.addEventListener("click", () => {
	duplicateBannerDismissed = true;
	dupBanner.classList.remove("visible");
});
dupClose.addEventListener("click", closeDuplicateModal);
dupCancelBtn.addEventListener("click", closeDuplicateModal);
dupModal.addEventListener("click", (e) => {
	if (e.target === dupModal) closeDuplicateModal();
});
dupResolveAllBtn.addEventListener("click", resolveAllDuplicateGroups);

collectionPickerClose.addEventListener("click", closeCollectionPicker);
collectionPickerCancelBtn.addEventListener("click", closeCollectionPicker);
collectionPickerModal.addEventListener("click", (e) => {
	if (e.target === collectionPickerModal) closeCollectionPicker();
});
collectionPickerNewBtn.addEventListener("click", async () => {
	const name = collectionPickerNewInput.value.trim();
	if (!name) return;
	const created = await createCollection(name);
	if (!created) {
		toast("Create collection failed", "error");
		return;
	}
	collectionPickerNewInput.value = "";
	await refresh();
	collectionPickerSelectedID = created.id;
	renderCollectionPickerOptions();
});
collectionPickerNewInput.addEventListener("keydown", (e) => {
	if (e.key === "Enter") collectionPickerNewBtn.click();
});
collectionPickerConfirmBtn.addEventListener("click", async () => {
	if (!collectionPickerSelectedID || !collectionPickerTargetHashes.length) return;
	const ok = await bulkAddCollectionVideos(
		collectionPickerSelectedID,
		collectionPickerTargetHashes,
	);
	if (!ok) {
		toast("Add to collection failed", "error");
		return;
	}
	await refresh();
	closeCollectionPicker();
	toast("Added to collection", "success");
});

watchCollectionClose.addEventListener("click", closeWatchCollection);
watchCollectionModal.addEventListener("click", (e) => {
	if (e.target === watchCollectionModal) closeWatchCollection();
});
watchCollectionDeleteBtn.addEventListener("click", async () => {
	if (!activeCollectionID) return;
	const c = COLLECTION_MAP[activeCollectionID];
	if (!c) return;
	if (!confirm('Delete collection "' + c.name + '"?')) return;
	const ok = await deleteCollection(activeCollectionID);
	if (!ok) {
		toast("Collection delete failed", "error");
		return;
	}
	closeWatchCollection();
	await refresh();
});
watchCollectionRenameBtn.addEventListener("click", async () => {
	if (!activeCollectionID) return;
	const c = COLLECTION_MAP[activeCollectionID];
	if (!c) return;
	const nextName = prompt("Rename collection", c.name);
	if (!nextName || !nextName.trim()) return;
	const ok = await renameCollection(activeCollectionID, nextName.trim());
	if (!ok) {
		toast("Collection rename failed", "error");
		return;
	}
	await refresh();
	openWatchCollection(activeCollectionID);
});

init();
