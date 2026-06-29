<script lang="ts">
import { onMount } from 'svelte';
import './app.css';
import { escHtml, formatBytes, formatModified } from './lib/format';
import {
	loadBooleanStorage,
	loadNumericStorage,
	persistStorageValue,
	WATCH_GRID_COLS_STORAGE_KEY,
	WATCH_GRID_VOLUME_STORAGE_KEY,
	WATCH_GRID_MUTED_STORAGE_KEY,
	HOVER_PREVIEW_AUDIO_ENABLED_KEY,
	HOVER_PREVIEW_VOLUME_KEY,
} from './lib/storage';
import {
	installFetch401Handler,
	setOnPinLock,
	setPinGateMode,
	setPinGateUnlocked,
} from './lib/auth';

let gallery: HTMLElement | null = null;
let folderList: HTMLElement | null = null;
let incomingList: HTMLElement | null = null;
let tagList: HTMLElement | null = null;
let searchEl: HTMLInputElement | null = null;
let statsEl: HTMLElement | null = null;
let sortSel: HTMLSelectElement | null = null;
let gridBtn: HTMLElement | null = null;
let listBtn: HTMLElement | null = null;
let favoritesOnlyToggle: HTMLInputElement | null = null;
let modal: HTMLElement | null = null;
let modalVid: HTMLVideoElement | null = null;
let modalTitle: HTMLElement | null = null;
let modalPath: HTMLElement | null = null;
let prevBtn: HTMLElement | null = null;
let nextBtn: HTMLElement | null = null;
let closeBtn: HTMLElement | null = null;
let selectBtn: HTMLElement | null = null;
let selectBar: HTMLElement | null = null;
let selectCount: HTMLElement | null = null;
let selectTagBtn: HTMLElement | null = null;
let selectMoveBtn: HTMLElement | null = null;
let selectForgetBtn: HTMLElement | null = null;
let selectAllBtn: HTMLElement | null = null;
let selectClearBtn: HTMLElement | null = null;
let newFolderBtn: HTMLElement | null = null;
let sidebarFolderForm: HTMLElement | null = null;
let sidebarFolderInput: HTMLInputElement | null = null;
let sidebarFolderConfirm: HTMLElement | null = null;
let sidebarFolderCancel: HTMLElement | null = null;
let newTagBtn: HTMLElement | null = null;
let sidebarTagForm: HTMLElement | null = null;
let sidebarTagInput: HTMLInputElement | null = null;
let sidebarTagConfirm: HTMLElement | null = null;
let sidebarTagCancel: HTMLElement | null = null;
let newCollectionBtn: HTMLElement | null = null;
let sidebarCollectionForm: HTMLElement | null = null;
let sidebarCollectionInput: HTMLInputElement | null = null;
let sidebarCollectionConfirm: HTMLElement | null = null;
let sidebarCollectionCancel: HTMLElement | null = null;
let collectionList: HTMLElement | null = null;
let collectionsCollapseBtn: HTMLElement | null = null;
let tagsCollapseBtn: HTMLElement | null = null;
let foldersCollapseBtn: HTMLElement | null = null;
let dupBanner: HTMLElement | null = null;
let dupBannerText: HTMLElement | null = null;
let dupReviewBtn: HTMLElement | null = null;
let dupDismissBtn: HTMLElement | null = null;
let dupModal: HTMLElement | null = null;
let dupSubtitle: HTMLElement | null = null;
let dupGroupsEl: HTMLElement | null = null;
let dupClose: HTMLElement | null = null;
let dupCancelBtn: HTMLElement | null = null;
let dupResolveAllBtn: HTMLElement | null = null;
let incomingBanner: HTMLElement | null = null;
let incomingBannerText: HTMLElement | null = null;
let incomingReviewBtn: HTMLElement | null = null;
let incomingMassSortBtn: HTMLElement | null = null;
let incomingDismissBtn: HTMLElement | null = null;
let incomingMarkAllBtn: HTMLElement | null = null;
let incomingSidebarSection: HTMLElement | null = null;
let incomingModal: HTMLElement | null = null;
let incomingSubtitle: HTMLElement | null = null;
let incomingTagOptions: HTMLElement | null = null;
let incomingFolderOptions: HTMLElement | null = null;
let incomingClose: HTMLElement | null = null;
let incomingCancelBtn: HTMLElement | null = null;
let incomingConfirmBtn: HTMLElement | null = null;
let incomingMarkReviewedBtn: HTMLElement | null = null;
let tagModal: HTMLElement | null = null;
let tagSubtitle: HTMLElement | null = null;
let tagOptions: HTMLElement | null = null;
let tagPendingSummary: HTMLElement | null = null;
let tagClose: HTMLElement | null = null;
let tagApplyBtn: HTMLElement | null = null;
let tagCancelBtn: HTMLElement | null = null;
let tagNewInput: HTMLInputElement | null = null;
let tagNewBtn: HTMLElement | null = null;
let selectCollectionBtn: HTMLElement | null = null;
let collectionPickerModal: HTMLElement | null = null;
let collectionPickerSubtitle: HTMLElement | null = null;
let collectionPickerOptions: HTMLElement | null = null;
let collectionPickerClose: HTMLElement | null = null;
let collectionPickerCancelBtn: HTMLElement | null = null;
let collectionPickerConfirmBtn: HTMLElement | null = null;
let collectionPickerNewInput: HTMLInputElement | null = null;
let collectionPickerNewBtn: HTMLElement | null = null;
let watchCollectionModal: HTMLElement | null = null;
let watchCollectionTitle: HTMLElement | null = null;
let watchCollectionSubtitle: HTMLElement | null = null;
let watchCollectionBody: HTMLElement | null = null;
let watchCollectionClose: HTMLElement | null = null;
let watchCollectionDeleteBtn: HTMLElement | null = null;
let watchCollectionRenameBtn: HTMLElement | null = null;
let watchCollectionPlayAllBtn: HTMLElement | null = null;
let watchCollectionPauseAllBtn: HTMLElement | null = null;
let watchCollectionMuteAllBtn: HTMLElement | null = null;
let watchCollectionVolumeInput: HTMLInputElement | null = null;
let watchCollectionGridCols: HTMLSelectElement | null = null;
let previewSettingsModal: HTMLElement | null = null;
let previewSettingsClose: HTMLElement | null = null;
let previewSettingsLauncher: HTMLElement | null = null;
let previewSettingsAudioToggleBtn: HTMLElement | null = null;
let previewSettingsVolumeInput: HTMLInputElement | null = null;
let pinGateOverlay: HTMLElement | null = null;
let pinGateForm: HTMLElement | null = null;
let pinGateInput: HTMLInputElement | null = null;
let pinGateError: HTMLElement | null = null;
let ctxMenu: HTMLElement | null = null;
let ctxTagsEl: HTMLElement | null = null;
let ctxMoveEl: HTMLElement | null = null;
let ctxRenameEl: HTMLElement | null = null;
let ctxDeleteEl: HTMLElement | null = null;
let ctxForgetEl: HTMLElement | null = null;
let moveModal: HTMLElement | null = null;
let moveSubtitle: HTMLElement | null = null;
let folderOptions: HTMLElement | null = null;
let newFolderName: HTMLInputElement | null = null;
let addFolderBtn: HTMLElement | null = null;
let moveCancelBtn: HTMLElement | null = null;
let moveConfirmBtn: HTMLElement | null = null;
let moveClose: HTMLElement | null = null;
let uploadModal: HTMLElement | null = null;
let dropZone: HTMLElement | null = null;
let uploadInput: HTMLInputElement | null = null;
let uploadFolderSel: HTMLSelectElement | null = null;
let newFolderInput: HTMLInputElement | null = null;
let fileQueue: HTMLElement | null = null;
let uploadSummary: HTMLElement | null = null;
let uploadStartBtn: HTMLElement | null = null;
let uploadCancelBtn: HTMLElement | null = null;
let uploadClose: HTMLElement | null = null;
let openUploadBtn: HTMLElement | null = null;
let tagsSection: HTMLElement | null = null;
let foldersSection: HTMLElement | null = null;
let collectionsSection: HTMLElement | null = null;
let pinLockBtn: HTMLElement | null = null;

let ALL_VIDEOS = [],
	DISCOVERED_VIDEOS = [],
	DOWNLOADED_VIDEOS = [],
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
let incomingBannerDismissed = false;
let showFavoritesOnly = false;
let pinGateMode = false;
let pinGateUnlocked = false;
let isVideoDiscoveryLoading = false;
let cardDragGhostEl = null;
let tagTargetPaths = [];
let pendingTagActions = new Map();
let activeCollectionID = "";
const HOVER_PREVIEW_SEEK_SEC = 2;
const PREVIEW_AUDIO_SVG_OFF =
	'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5"/><line x1="23" y1="9" x2="17" y2="15"/><line x1="17" y1="9" x2="23" y2="15"/></svg>';
const PREVIEW_AUDIO_SVG_ON =
	'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5"/><path d="M15.54 8.46a5 5 0 0 1 0 7.07"/><path d="M19.07 4.93a10 10 0 0 1 0 14.14"/></svg>';
let hoverPreviewAudioEnabled = loadBooleanStorage(
	HOVER_PREVIEW_AUDIO_ENABLED_KEY,
	false,
);
let hoverPreviewVolume = loadNumericStorage(
	HOVER_PREVIEW_VOLUME_KEY,
	0.5,
	0,
	1,
);
let collectionPickerTargetHashes = [];
let activeIncomingFilter = false;
let incomingSelectedFolder = "/";
let incomingSelectedTagIDs = new Set();
const DEFAULT_LOOP_ALL_VIDEOS = true;

function applyDefaultVideoFlags(videoEl) {
	if (!videoEl) return;
	videoEl.loop = DEFAULT_LOOP_ALL_VIDEOS;
}

function applyHoverPreviewAudioToElement(videoEl) {
	if (!videoEl || !gallery.contains(videoEl)) return;
	videoEl.volume = hoverPreviewVolume;
	videoEl.muted = !hoverPreviewAudioEnabled;
}

function applyHoverPreviewAudioToAll() {
	gallery.querySelectorAll(".thumb video").forEach((v) => {
		v.volume = hoverPreviewVolume;
		v.muted = !hoverPreviewAudioEnabled;
	});
	syncPreviewSpeakerButtons();
	syncPreviewSettingsModalUI();
}

function syncPreviewSpeakerButtons() {
	const on = hoverPreviewAudioEnabled;
	const title = on
		? "Preview sound on for all cards (click to mute)"
		: "Preview sound off (click to enable for all cards)";
	const label = "Toggle hover preview sound for all cards";
	const svg = on ? PREVIEW_AUDIO_SVG_ON : PREVIEW_AUDIO_SVG_OFF;
	gallery.querySelectorAll(".preview-audio-btn").forEach((btn) => {
		btn.classList.toggle("active", on);
		btn.title = title;
		btn.setAttribute("aria-label", label);
		btn.setAttribute("aria-pressed", on ? "true" : "false");
		btn.innerHTML = svg;
	});
}

function syncPreviewSettingsModalUI() {
	if (previewSettingsVolumeInput) {
		previewSettingsVolumeInput.value = String(hoverPreviewVolume);
	}
	if (previewSettingsAudioToggleBtn) {
		previewSettingsAudioToggleBtn.classList.toggle(
			"active",
			hoverPreviewAudioEnabled,
		);
		previewSettingsAudioToggleBtn.textContent = hoverPreviewAudioEnabled
			? "On"
			: "Off";
	}
}

function openPreviewSettings() {
	if (!previewSettingsModal) return;
	syncPreviewSettingsModalUI();
	previewSettingsModal.classList.add("open");
}

function closePreviewSettings() {
	if (!previewSettingsModal) return;
	previewSettingsModal.classList.remove("open");
}

function setHoverPreviewAudioEnabled(next) {
	hoverPreviewAudioEnabled = !!next;
	persistStorageValue(
		HOVER_PREVIEW_AUDIO_ENABLED_KEY,
		hoverPreviewAudioEnabled,
	);
	applyHoverPreviewAudioToAll();
}

function toggleHoverPreviewAudio() {
	setHoverPreviewAudioEnabled(!hoverPreviewAudioEnabled);
}

function seekAndStartHoverPreview(card, vid) {
	if (!vid || !vid.src || card.dataset.hoverPreview !== "1") return;
	const doPlay = () => {
		if (card.dataset.hoverPreview !== "1") return;
		const dur = vid.duration;
		if (dur && !Number.isNaN(dur) && Number.isFinite(dur)) {
			const t = Math.min(
				HOVER_PREVIEW_SEEK_SEC,
				Math.max(0, dur - 0.05),
			);
			try {
				vid.currentTime = t;
			} catch (_) {}
		}
		applyHoverPreviewAudioToElement(vid);
		vid.play().catch(() => {});
	};
	if (vid.readyState >= HTMLMediaElement.HAVE_CURRENT_DATA) {
		doPlay();
	} else {
		vid.addEventListener("loadeddata", doPlay, { once: true });
	}
}

function stopHoverPreview(card, vid) {
	card.dataset.hoverPreview = "0";
	if (!vid || !vid.src) return;
	vid.pause();
	const dur = vid.duration;
	if (dur && !Number.isNaN(dur) && Number.isFinite(dur)) {
		try {
			vid.currentTime = Math.min(
				HOVER_PREVIEW_SEEK_SEC,
				Math.max(0, dur - 0.05),
			);
		} catch (_) {}
	}
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
	const raw = data && Array.isArray(data.collections) ? data.collections : [];
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

function rebuildAllVideos() {
	const mergedByPath = new Map();
	for (const video of DOWNLOADED_VIDEOS) {
		if (video && video.path) mergedByPath.set(video.path, video);
	}
	for (const video of DISCOVERED_VIDEOS) {
		if (video && video.path) mergedByPath.set(video.path, video);
	}
	ALL_VIDEOS = [...mergedByPath.values()];
}

function setDiscoveredVideos(videos) {
	DISCOVERED_VIDEOS = Array.isArray(videos) ? videos : [];
	const discoveredPaths = new Set(
		DISCOVERED_VIDEOS.map((v) => v.path).filter(Boolean),
	);
	DOWNLOADED_VIDEOS = DOWNLOADED_VIDEOS.filter(
		(v) => !discoveredPaths.has(v.path),
	);
	rebuildAllVideos();
}

function upsertDownloadedVideo(video) {
	if (!video || !video.path) return;
	const idx = DOWNLOADED_VIDEOS.findIndex((v) => v.path === video.path);
	if (idx >= 0) DOWNLOADED_VIDEOS[idx] = video;
	else DOWNLOADED_VIDEOS.push(video);
	rebuildAllVideos();
}

function refreshSidebarsOnly() {
	computeDuplicateGroups();
	buildIncomingNav();
	buildTagNav();
	buildFolderNav();
	buildCollectionNav();
}

function refreshDerivedUI(mayAutoOpenDuplicates) {
	refreshSidebarsOnly();
	render();
	updateDuplicateBanner(Boolean(mayAutoOpenDuplicates));
	updateIncomingBanner();
	populateUploadFolders();
}

function recomputeFiltered() {
	const q = searchEl.value.toLowerCase(),
		sort = sortSel.value;
	filtered = ALL_VIDEOS.filter((v) => {
		const inF =
			activeFolder === "__all__" || (v.folder || "/") == activeFolder;
		const inFav = !showFavoritesOnly || Boolean(v.is_favorite);
		const inT =
			activeTagFilters.size === 0 ||
			[...activeTagFilters].every((tagID) =>
				(v.tags || []).includes(tagID),
			);
		const inS =
			!q ||
			v.name.toLowerCase().includes(q) ||
			(v.folder || "").toLowerCase().includes(q);
		const inIncoming = !activeIncomingFilter || Boolean(v.is_new);
		return inF && inFav && inT && inS && inIncoming;
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
				new Date(a.modified || 0).getTime() -
					new Date(b.modified || 0).getTime() ||
				a.name.localeCompare(b.name)
			);
		if (sort === "modified-desc")
			return (
				new Date(b.modified || 0).getTime() -
					new Date(a.modified || 0).getTime() ||
				a.name.localeCompare(b.name)
			);
		if (sort === "size")
			return (
				Number(a.size || 0) - Number(b.size || 0) ||
				a.name.localeCompare(b.name)
			);
		if (sort === "size-desc")
			return (
				Number(b.size || 0) - Number(a.size || 0) ||
				a.name.localeCompare(b.name)
			);
		return 0;
	});
}

function updateGalleryStatsMarkup() {
	const filteredCountMarkup =
		"<b>" + filtered.length + "</b> / " + ALL_VIDEOS.length + " videos";
	if (isVideoDiscoveryLoading) {
		statsEl.innerHTML =
			'<div class="stats-loading">' +
			'<span class="stats-loading-label">' +
			filteredCountMarkup +
			" · scanning library (" +
			ALL_VIDEOS.length +
			' found)</span><span class="stats-loading-bar" aria-hidden="true"><span class="stats-loading-fill"></span></span></div>';
	} else {
		statsEl.innerHTML = filteredCountMarkup;
	}
}

function patchGalleryTagBadgesForHashes(hashes) {
	const hashSet = new Set((hashes || []).filter(Boolean));
	if (!hashSet.size) return;
	const escapePath =
		typeof CSS !== "undefined" && typeof CSS.escape === "function"
			? (p) => CSS.escape(p)
			: (p) =>
					String(p).replace(/\\/g, "\\\\").replace(/"/g, '\\"');
	for (const video of ALL_VIDEOS) {
		if (!video || !video.path || !video.hash || !hashSet.has(video.hash))
			continue;
		const card = gallery.querySelector(
			'.card[data-path="' + escapePath(video.path) + '"]',
		);
		if (!card) continue;
		const badges = card.querySelector(".card-badges");
		if (!badges) continue;
		const inner = renderVideoTagChipsInner(video.tags || []);
		let tagSpan = badges.querySelector(".card-tags.card-tags--badges");
		if (inner) {
			if (!tagSpan) {
				tagSpan = document.createElement("span");
				tagSpan.className = "card-tags card-tags--badges";
				badges.appendChild(tagSpan);
			}
			tagSpan.innerHTML = inner;
		} else if (tagSpan) {
			tagSpan.remove();
		}
	}
}

function refreshAfterTagDataChange(prevFilteredPaths, affectedHashes) {
	refreshSidebarsOnly();
	recomputeFiltered();
	const nextPaths = filtered.map((v) => v.path);
	const unchanged =
		prevFilteredPaths.length === nextPaths.length &&
		prevFilteredPaths.every((p, i) => p === nextPaths[i]);
	if (unchanged) {
		updateGalleryStatsMarkup();
		patchGalleryTagBadgesForHashes(affectedHashes);
	} else {
		render();
	}
	updateDuplicateBanner(false);
	updateIncomingBanner();
	populateUploadFolders();
}

function countIncomingVideos() {
	return ALL_VIDEOS.filter((v) => Boolean(v.is_new)).length;
}

function buildIncomingNav() {
	incomingList.innerHTML = "";
	const incomingCount = countIncomingVideos();
	if (incomingCount === 0) {
		activeIncomingFilter = false;
	}
	incomingMarkAllBtn.style.display = incomingCount > 0 ? "" : "none";
	if (incomingSidebarSection) {
		incomingSidebarSection.classList.toggle(
			"incoming-sidebar--disabled",
			incomingCount === 0,
		);
	}
	const btn = document.createElement("button");
	btn.type = "button";
	btn.className =
		"tag-btn need-sorting-nav" +
		(activeIncomingFilter ? " active" : "");
	btn.disabled = incomingCount === 0;
	btn.setAttribute("aria-disabled", incomingCount === 0 ? "true" : "false");
	btn.title =
		incomingCount === 0
			? "No incoming files need sorting"
			: activeIncomingFilter
				? "Show all videos"
				: "Show only incoming (need sorting)";
	btn.innerHTML =
		'<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2v20"/><path d="M2 12h20"/></svg>' +
		"<span>Need sorting</span>" +
		'<span class="count">' +
		incomingCount +
		"</span>";
	if (incomingCount > 0) {
		btn.addEventListener("click", () => {
			activeIncomingFilter = !activeIncomingFilter;
			if (activeIncomingFilter) activeFolder = "__all__";
			buildIncomingNav();
			buildFolderNav();
			render();
		});
	}
	incomingList.appendChild(btn);
}

function tagCountByID(tagID) {
	return ALL_VIDEOS.filter(
		(v) => Array.isArray(v.tags) && v.tags.includes(tagID),
	).length;
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

function renderVideoTagChipsInner(tagIDs) {
	if (!Array.isArray(tagIDs) || !tagIDs.length) return "";
	const chips = tagIDs
		.map((id) => TAG_MAP[id])
		.filter(Boolean)
		.map((tag) => renderTagChip(tag))
		.join("");
	return chips || "";
}
function renderVideoTagChips(tagIDs) {
	const chips = renderVideoTagChipsInner(tagIDs);
	if (!chips) return "";
	return '<div class="card-tags">' + chips + "</div>";
}

function setPinOverlayVisible(show, message) {
	if (!pinGateOverlay) return;
	pinGateOverlay.hidden = !show;
	pinGateOverlay.setAttribute("aria-hidden", show ? "false" : "true");
	document.body.classList.toggle("pin-gate-open", Boolean(show));
	if (pinGateInput) {
		pinGateInput.disabled = !show;
		if (!show) pinGateInput.readOnly = false;
	}
	if (pinGateError) {
		if (message) {
			pinGateError.hidden = false;
			pinGateError.textContent = message;
		} else {
			pinGateError.hidden = true;
			pinGateError.textContent = "";
		}
	}
	if (show && pinGateInput) {
		pinGateInput.readOnly = true;
		pinGateInput.addEventListener(
			"focus",
			() => {
				pinGateInput.readOnly = false;
			},
			{ once: true },
		);
		queueMicrotask(() => pinGateInput.focus());
	}
}

function updatePinLockButton() {
	if (!pinLockBtn) return;
	pinLockBtn.hidden = !(pinGateMode && pinGateUnlocked);
}

async function bootstrap() {
	let st;
	try {
		const res = await fetch("/api/auth/status");
		st = await res.json();
	} catch (e) {
		console.error(e);
		return;
	}
	if (!st.pinRequired) {
		setPinGateMode(false);
		pinGateMode = false;
		setPinGateUnlocked(false);
		pinGateUnlocked = false;
		if (pinLockBtn) {
			pinLockBtn.remove();
			pinLockBtn = null;
		}
		await init();
		return;
	}
	setPinGateMode(true);
	pinGateMode = true;
	installFetch401Handler();
	if (st.authenticated) {
		setPinGateUnlocked(true);
		pinGateUnlocked = true;
		setPinOverlayVisible(false);
		updatePinLockButton();
		await init();
		return;
	}
	setPinGateUnlocked(false);
	pinGateUnlocked = false;
	updatePinLockButton();
	setPinOverlayVisible(true);
}

async function init() {
	const [fr, tr, cr] = await Promise.all([
		fetch("/api/folders"),
		fetch("/api/tags"),
		fetch("/api/collections"),
	]);
	parseFolders(await fr.json());
	parseTags(await tr.json());
	parseCollections(await cr.json());
	setDiscoveredVideos([]);
	refreshDerivedUI(true);
	applyDefaultVideoFlags(modalVid);
	await refreshVideosProgressive(true);
	syncPreviewSettingsModalUI();
}
async function refresh() {
	const [fr, tr, cr] = await Promise.all([
		fetch("/api/folders"),
		fetch("/api/tags"),
		fetch("/api/collections"),
	]);
	parseFolders(await fr.json());
	parseTags(await tr.json());
	parseCollections(await cr.json());
	await refreshVideosProgressive(false);
}

async function refreshTagsOnly() {
	const tr = await fetch("/api/tags");
	parseTags(await tr.json());
}

async function refreshCollectionsOnly() {
	const cr = await fetch("/api/collections");
	parseCollections(await cr.json());
}

async function refreshFoldersOnly() {
	const fr = await fetch("/api/folders");
	parseFolders(await fr.json());
}

async function refreshTagAssignmentsForHashes(hashes) {
	const uniqueHashes = [...new Set((hashes || []).filter(Boolean))];
	if (!uniqueHashes.length) return;
	const res = await fetch("/api/tags/videos", {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify({ hashes: uniqueHashes }),
	});
	if (!res.ok) throw new Error("failed to refresh tag assignments");
	const payload = await res.json();
	const updates = Array.isArray(payload?.videos) ? payload.videos : [];
	const tagsByHash = Object.fromEntries(
		updates.map((u) => [u.hash, Array.isArray(u.tags) ? u.tags : []]),
	);
	for (const video of ALL_VIDEOS) {
		if (!video || !video.hash) continue;
		if (Object.prototype.hasOwnProperty.call(tagsByHash, video.hash)) {
			video.tags = tagsByHash[video.hash];
		}
	}
}

async function refreshVideosFallback(mayAutoOpenDuplicates) {
	const vr = await fetch("/api/videos");
	setDiscoveredVideos(await vr.json());
	isVideoDiscoveryLoading = false;
	refreshDerivedUI(mayAutoOpenDuplicates);
}

async function refreshVideosProgressive(mayAutoOpenDuplicates) {
	isVideoDiscoveryLoading = true;
	try {
		const res = await fetch("/api/videos/stream");
		if (!res.ok || !res.body) {
			await refreshVideosFallback(mayAutoOpenDuplicates);
			return;
		}
		const reader = res.body.getReader();
		const decoder = new TextDecoder();
		let buffer = "";
		let streamed = [];
		let lastRenderCount = 0;
		setDiscoveredVideos([]);
		refreshDerivedUI(mayAutoOpenDuplicates);
		while (true) {
			const { value, done } = await reader.read();
			if (done) break;
			buffer += decoder.decode(value, { stream: true });
			const lines = buffer.split("\n");
			buffer = lines.pop() || "";
			for (const line of lines) {
				const trimmed = line.trim();
				if (!trimmed) continue;
				let event;
				try {
					event = JSON.parse(trimmed);
				} catch (e) {
					continue;
				}
				if (event.type === "video" && event.payload) {
					streamed.push(event.payload);
					if (streamed.length - lastRenderCount >= 20) {
						setDiscoveredVideos(streamed);
						refreshDerivedUI(false);
						lastRenderCount = streamed.length;
					}
				}
			}
		}
		setDiscoveredVideos(streamed);
		isVideoDiscoveryLoading = false;
		refreshDerivedUI(mayAutoOpenDuplicates);
	} catch (e) {
		await refreshVideosFallback(mayAutoOpenDuplicates);
	}
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
		})
		.sort(
			(a, b) =>
				b.members.length - a.members.length ||
				a.hash.localeCompare(b.hash),
		);
}

function updateDuplicateBanner(mayAutoOpen) {
	const groups = DUPLICATE_GROUPS.length;
	const files = DUPLICATE_GROUPS.reduce(
		(acc, g) => acc + g.members.length,
		0,
	);
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
	const shouldShow =
		!duplicateBannerDismissed || dupModal.classList.contains("open");
	dupBanner.classList.toggle("visible", shouldShow);

	if (mayAutoOpen && !autoOpenedDuplicates && groups > 0) {
		openDuplicateModal();
		autoOpenedDuplicates = true;
	}
}

function updateIncomingBanner() {
	const incomingCount = countIncomingVideos();
	if (!incomingCount) {
		incomingBanner.classList.remove("visible");
		incomingBannerDismissed = false;
		if (incomingModal.classList.contains("open")) closeIncomingModal();
		return;
	}
	incomingBannerText.textContent =
		incomingCount +
		" new file" +
		(incomingCount === 1 ? "" : "s") +
		" need sorting.";
	const shouldShow =
		!incomingBannerDismissed || incomingModal.classList.contains("open");
	incomingBanner.classList.toggle("visible", shouldShow);
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
		header
			.querySelector(".dup-resolve-btn")
			.addEventListener("click", async () => {
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
				list.querySelectorAll(".dup-item").forEach((el) =>
					el.classList.remove("keep"),
				);
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
		btn.className =
			"tag-btn" + (activeTagFilters.has(tag.id) ? " active" : "");
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
		btn.querySelector(".tag-remove-btn").addEventListener(
			"click",
			async (e) => {
				e.stopPropagation();
				if (!confirm('Delete tag "' + tag.name + '"?')) return;
				const ok = await deleteTag(tag.id);
				if (!ok) {
					toast("Tag delete failed", "error");
					return;
				}
				activeTagFilters.delete(tag.id);
				try {
					recomputeFiltered();
					const prevFilteredPaths = filtered.map((v) => v.path);
					const affectedHashes = [
						...new Set(
							ALL_VIDEOS.filter((v) =>
								Array.isArray(v.tags) && v.tags.includes(tag.id),
							)
								.map((v) => v.hash)
								.filter(Boolean),
						),
					];
					await refreshTagsOnly();
					await refreshTagAssignmentsForHashes(affectedHashes);
					refreshAfterTagDataChange(prevFilteredPaths, affectedHashes);
				} catch (e) {
					await refresh();
				}
			},
		);
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
		btn.title = f === "__all__" ? "All files" : f;
		const hasOther = f !== "__all__" && FOLDER_META[f]?.has_other_files;
		btn.innerHTML =
			'<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>' +
			'<span class="folder-name">' +
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
			if (activeIncomingFilter) {
				activeIncomingFilter = false;
				buildIncomingNav();
			}
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
			if (f !== "/") {
				const makeCollectionBtn = document.createElement("button");
				makeCollectionBtn.className = "watch-folder-collection-btn";
				makeCollectionBtn.title = "Make collection from this folder";
				makeCollectionBtn.textContent = "◫";
				makeCollectionBtn.addEventListener("click", async (e) => {
					e.stopPropagation();
					const suggestedName = f;
					const inputName = prompt(
						"Create collection from folder",
						suggestedName,
					);
					if (inputName === null) return;
					const name = inputName.trim() || suggestedName;
					const result = await createCollectionFromFolder(f, name);
					if (!result.ok) {
						toast(
							"Create collection failed: " + (result.error || "unknown error"),
							"error",
						);
						return;
					}
					await refresh();
					openWatchCollection(result.collection.id);
					toast("Collection created from folder", "success");
				});
				btn.appendChild(makeCollectionBtn);

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
			"collection-btn" +
			(collection.id === activeCollectionID ? " active" : "");
		btn.innerHTML =
			'<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/></svg>' +
			"<span>" +
			escHtml(collection.name) +
			"</span>" +
			'<span class="count">' +
			countCollectionResolvedVideos(collection) +
			'</span><button class="collection-rename-btn" title="Rename collection">✎</button><button class="collection-remove-btn" title="Delete collection">✕</button>';

		btn.addEventListener("click", () => openWatchCollection(collection.id));
		btn.querySelector(".collection-remove-btn").addEventListener(
			"click",
			async (e) => {
				e.stopPropagation();
				if (!confirm('Delete collection "' + collection.name + '"?'))
					return;
				const ok = await deleteCollection(collection.id);
				if (!ok) {
					toast("Collection delete failed", "error");
					return;
				}
				if (activeCollectionID === collection.id)
					closeWatchCollection();
				await refresh();
				toast("Deleted " + collection.name, "success");
			},
		);
		btn.querySelector(".collection-rename-btn").addEventListener(
			"click",
			async (e) => {
				e.stopPropagation();
				const nextName = prompt("Rename collection", collection.name);
				if (!nextName || !nextName.trim()) return;
				const ok = await renameCollection(
					collection.id,
					nextName.trim(),
				);
				if (!ok) {
					toast("Collection rename failed", "error");
					return;
				}
				await refresh();
				if (activeCollectionID === collection.id)
					openWatchCollection(collection.id);
			},
		);
		collectionList.appendChild(btn);
	}
}

const CARD_DRAG_GHOST_SIZE = 32;
const CARD_DRAG_GHOST_OFFSET = CARD_DRAG_GHOST_SIZE / 2;

/**
 * Small transparent video icon used as the HTML5 drag image so the whole card
 * (including video) is not shown while dragging.
 * @returns {HTMLDivElement}
 */
function cardDragGhost() {
	if (cardDragGhostEl) return cardDragGhostEl;
	cardDragGhostEl = document.createElement("div");
	cardDragGhostEl.className = "card-drag-ghost";
	cardDragGhostEl.setAttribute("aria-hidden", "true");
	cardDragGhostEl.innerHTML =
		'<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" aria-hidden="true">' +
		'<rect x="2.5" y="5" width="19" height="14" rx="2" stroke="rgba(255,255,255,.65)" stroke-width="1.5" fill="rgba(255,255,255,.08)"/>' +
		'<polygon points="10,9.5 10,14.5 15,12" fill="rgba(255,255,255,.58)"/>' +
		"</svg>";
	document.body.appendChild(cardDragGhostEl);
	return cardDragGhostEl;
}

function render() {
	recomputeFiltered();
	updateGalleryStatsMarkup();
	gallery.innerHTML = "";
	if (!filtered.length) {
		if (isVideoDiscoveryLoading && ALL_VIDEOS.length === 0) {
			gallery.innerHTML =
				'<div class="empty"><strong>Loading videos...</strong>Please wait while files are discovered.</div>';
			return;
		}
		const otherWarn =
			activeFolder !== "__all__" &&
			FOLDER_META[activeFolder]?.has_other_files
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
			'<div class="check-overlay"><svg viewBox="0 0 10 8" fill="none" width="10" height="8"><polyline points="1,4 3.5,7 9,1" stroke="#000" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/></svg></div>' +
			'<button type="button" class="preview-audio-btn' +
			(hoverPreviewAudioEnabled ? " active" : "") +
			'" title="' +
			(hoverPreviewAudioEnabled
				? "Preview sound on for all cards (click to mute)"
				: "Preview sound off (click to enable for all cards)") +
			'" aria-label="Toggle hover preview sound for all cards" aria-pressed="' +
			(hoverPreviewAudioEnabled ? "true" : "false") +
			'">' +
			(hoverPreviewAudioEnabled
				? PREVIEW_AUDIO_SVG_ON
				: PREVIEW_AUDIO_SVG_OFF) +
			"</button></div>" +
			'<div class="card-meta"><span class="card-name" title="' +
			escHtml(v.name) +
			'">' +
			escHtml(v.name) +
			"</span>" +
			'<span class="card-path">' +
			escHtml(v.folder || "/") +
			"</span>" +
			'<div class="card-badges"><span class="card-ext">' +
			escHtml(v.ext.slice(1)) +
			"</span>" +
			(v.is_new ? '<span class="card-ext">new</span>' : "") +
			(() => {
				const inner = renderVideoTagChipsInner(v.tags || []);
				return inner
					? '<span class="card-tags card-tags--badges">' + inner + "</span>"
					: "";
			})() +
			"</div>" +
			'<span class="card-path">' +
			escHtml(formatBytes(v.size || 0)) +
			" • " +
			escHtml(formatModified(v.modified)) +
			"</span>" +
			"</div>";
		const vid = card.querySelector("video");
		applyDefaultVideoFlags(vid);
		card.dataset.hoverPreview = "0";
		const favBtn = card.querySelector(".fav-toggle");
		const previewAudioBtn = card.querySelector(".preview-audio-btn");
		const obs = new IntersectionObserver(
			(entries) => {
				if (entries[0].isIntersecting) {
					vid.src =
						"/video?path=" + encodeURIComponent(v.path) + "#t=2";
					vid.addEventListener(
						"loadeddata",
						() => {
							vid.classList.add("loaded");
							applyHoverPreviewAudioToElement(vid);
							if (card.dataset.hoverPreview === "1") {
								seekAndStartHoverPreview(card, vid);
							}
						},
						{ once: true },
					);
					obs.disconnect();
				}
			},
			{ threshold: 0.1 },
		);
		obs.observe(card);
		card.addEventListener("mouseenter", () => {
			card.dataset.hoverPreview = "1";
			seekAndStartHoverPreview(card, vid);
		});
		card.addEventListener("mouseleave", () => {
			stopHoverPreview(card, vid);
		});
		previewAudioBtn.addEventListener("click", (e) => {
			e.preventDefault();
			e.stopPropagation();
			toggleHoverPreviewAudio();
		});
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
			if (e.target.closest(".preview-audio-btn")) return;
			if (e.target.closest(".check-overlay")) {
				if (!selectMode) {
					selectMode = true;
					document.body.classList.add("select-mode");
					selectBtn.classList.add("active");
				}
				toggleSelect(v.path, card);
				return;
			}
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
			try {
				e.dataTransfer.setDragImage(
					cardDragGhost(),
					CARD_DRAG_GHOST_OFFSET,
					CARD_DRAG_GHOST_OFFSET,
				);
			} catch (_) {
				/* setDragImage unsupported */
			}
		});
		card.addEventListener("dragend", () =>
			card.classList.remove("dragging"),
		);
		if (selectMode && selectedPaths.has(v.path)) {
			card.classList.add("selected");
		}
		gallery.appendChild(card);
	});
	if (activeIncomingFilter && filtered.length) {
		const actions = document.createElement("div");
		actions.style.gridColumn = "1 / -1";
		actions.style.display = "flex";
		actions.style.justifyContent = "center";
		actions.style.paddingTop = "8px";
		const btn = document.createElement("button");
		btn.className = "btn";
		btn.textContent = "Mark as resolved";
		btn.addEventListener("click", async () => {
			const paths = filtered.map((v) => v.path).filter(Boolean);
			if (!paths.length) return;
			const ok = await clearNewStatus(paths, false);
			if (!ok) {
				toast("Failed to mark filtered videos as reviewed", "error");
				return;
			}
			await refresh();
			toast("Marked filtered videos as reviewed", "success");
		});
		actions.appendChild(btn);
		gallery.appendChild(actions);
	}
}

/**
 * Opens the video lightbox for the item at index idx within the current
 * filtered list.
 * @param {number} idx
 */
function openModal(idx) {
	gallery.querySelectorAll(".thumb video").forEach((el) => {
		el.pause();
	});
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

/** Rebuilds the destination folder <select> in the upload modal from the
 *  current ALL_FOLDERS / ALL_VIDEOS state, preserving the previous selection. */
function populateUploadFolders() {
	if (!uploadFolderSel) return;
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

// --- hoisted from wireAppEvents (module scope) ---
let ctxVideo = null;
let moveTarget = null;
let moveSelectedFolder = null;
let uploadFiles = [];
let collectionPickerSelectedID = "";
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

function showCtxMenu(e, v) {
	ctxVideo = v;
	ctxMenu.style.left = Math.min(e.clientX, window.innerWidth - 180) + "px";
	ctxMenu.style.top = Math.min(e.clientY, window.innerHeight - 60) + "px";
	ctxMenu.classList.add("open");
}
function resolveTagTargets() {
	return tagTargetPaths
		.map((path) => ALL_VIDEOS.find((v) => v.path === path))
		.filter(Boolean);
}
function openTagModal(videoOrVideos) {
	const videos = Array.isArray(videoOrVideos)
		? videoOrVideos
		: [videoOrVideos];
	tagTargetPaths = videos.map((v) => v.path).filter(Boolean);
	tagSubtitle.textContent =
		videos.length > 1
			? videos.length + " videos selected"
			: videos[0]?.name || "";
	pendingTagActions = new Map();
	tagNewInput.value = "";
	renderTagOptions();
	tagModal.classList.add("open");
}
function closeTagModal() {
	tagModal.classList.remove("open");
	tagTargetPaths = [];
	pendingTagActions = new Map();
}
function getPendingTagActionEntries() {
	const validTagIDs = new Set(ALL_TAGS.map((tag) => tag.id));
	const entries = [];
	for (const [tagID, action] of pendingTagActions.entries()) {
		if (!validTagIDs.has(tagID)) continue;
		if (action !== "add" && action !== "remove") continue;
		entries.push([tagID, action]);
	}
	return entries;
}
function renderTagPendingSummary() {
	const entries = getPendingTagActionEntries();
	let addCount = 0;
	let removeCount = 0;
	for (const [, action] of entries) {
		if (action === "add") addCount++;
		if (action === "remove") removeCount++;
	}
	const parts = [];
	if (addCount) parts.push(addCount + " add");
	if (removeCount) parts.push(removeCount + " remove");
	tagPendingSummary.textContent = parts.length
		? "Pending: " + parts.join(", ")
		: "No pending tag changes";
	tagApplyBtn.disabled = entries.length === 0;
}
function renderTagOptions() {
	tagOptions.innerHTML = "";
	const targets = resolveTagTargets();
	if (!targets.length) {
		renderTagPendingSummary();
		return;
	}
	for (const tag of ALL_TAGS) {
		const selectedCount = targets.filter((v) =>
			(v.tags || []).includes(tag.id),
		).length;
		const pendingAction = pendingTagActions.get(tag.id);
		const row = document.createElement("div");
		row.className = "tag-option";
		const stateText =
			pendingAction === "add"
				? "pending add"
				: pendingAction === "remove"
					? "pending remove"
					: selectedCount === 0
						? "none selected"
						: selectedCount === targets.length
							? "all selected"
							: selectedCount + "/" + targets.length + " selected";
		row.innerHTML =
			'<span class="tag-option-name">' +
			renderTagChip(tag) +
			'<span class="tag-option-state">' +
			escHtml(stateText) +
			'</span></span><button class="btn tag-option-action' +
			(pendingAction === "add" ? " active" : "") +
			'" type="button" data-action="add">Add</button><button class="btn tag-option-action' +
			(pendingAction === "remove" ? " active" : "") +
			'" type="button" data-action="remove">Remove</button><button class="tag-option-remove" type="button">delete</button>';
		row.querySelector('[data-action="add"]').addEventListener("click", () => {
			const prev = pendingTagActions.get(tag.id);
			if (prev === "add") pendingTagActions.delete(tag.id);
			else pendingTagActions.set(tag.id, "add");
			renderTagOptions();
		});
		row
			.querySelector('[data-action="remove"]')
			.addEventListener("click", () => {
				const prev = pendingTagActions.get(tag.id);
				if (prev === "remove") pendingTagActions.delete(tag.id);
				else pendingTagActions.set(tag.id, "remove");
				renderTagOptions();
			});
		row.querySelector(".tag-option-remove").addEventListener(
			"click",
			async (e) => {
				e.preventDefault();
				e.stopPropagation();
				if (!confirm('Delete tag "' + tag.name + '"?')) return;
				const ok = await deleteTag(tag.id);
				if (!ok) {
					toast("Tag delete failed", "error");
					return;
				}
				activeTagFilters.delete(tag.id);
				try {
					recomputeFiltered();
					const prevFilteredPaths = filtered.map((v) => v.path);
					const affectedHashes = [
						...new Set(
							ALL_VIDEOS.filter((v) =>
								Array.isArray(v.tags) && v.tags.includes(tag.id),
							)
								.map((v) => v.hash)
								.filter(Boolean),
						),
					];
					await refreshTagsOnly();
					await refreshTagAssignmentsForHashes(affectedHashes);
					refreshAfterTagDataChange(prevFilteredPaths, affectedHashes);
				} catch (e) {
					await refresh();
				}
				pendingTagActions.delete(tag.id);
				renderTagOptions();
			},
		);
		tagOptions.appendChild(row);
	}
	renderTagPendingSummary();
}
function openMoveModal(v) {
	moveTarget = v;
	moveSubtitle.textContent =
		selectedPaths.size > 1
			? selectedPaths.size + " items selected"
			: v.name;
	moveSelectedFolder = v.folder || "/";
	newFolderName.value = "";
	renderFolderOptions();
	moveModal.classList.add("open");
}
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
function closeMoveModal() {
	moveModal.classList.remove("open");
}
function movedVideoRelPath(srcPath, destFolder) {
	const slash = srcPath.lastIndexOf("/");
	const base = slash >= 0 ? srcPath.slice(slash + 1) : srcPath;
	let dest = destFolder == null ? "" : String(destFolder).trim();
	if (dest === "." || dest === "/") dest = "";
	const rel = dest
		? dest.replace(/^[/\\]+/, "").replace(/[/\\]+$/, "") + "/" + base
		: base;
	return rel.replace(/\\/g, "/");
}
function folderFromRelPath(relPath) {
	const i = relPath.lastIndexOf("/");
	if (i < 0) return "/";
	const dir = relPath.slice(0, i);
	return dir === "" ? "/" : dir.replace(/\\/g, "/");
}
function galleryCardSelectorForPath(p) {
	const esc =
		typeof CSS !== "undefined" && typeof CSS.escape === "function"
			? CSS.escape(p)
			: String(p).replace(/\\/g, "\\\\").replace(/"/g, '\\"');
	return '.card[data-path="' + esc + '"]';
}
function removeGalleryCardByPath(path) {
	const card = gallery.querySelector(galleryCardSelectorForPath(path));
	if (card) card.remove();
}
function patchVideoCardDomAfterPathChange(card, v) {
	card.dataset.path = v.path;
	const folderSpans = card.querySelectorAll(".card-path");
	if (folderSpans[0]) folderSpans[0].textContent = v.folder || "/";
	const vid = card.querySelector("video");
	if (vid) {
		vid.classList.remove("loaded");
		vid.src = "/video?path=" + encodeURIComponent(v.path) + "#t=2";
		vid.addEventListener(
			"loadeddata",
			() => {
				vid.classList.add("loaded");
				applyHoverPreviewAudioToElement(vid);
				if (card.dataset.hoverPreview === "1") {
					seekAndStartHoverPreview(card, vid);
				}
			},
			{ once: true },
		);
	}
	if (
		modal.classList.contains("open") &&
		currentIdx >= 0 &&
		filtered[currentIdx] === v
	) {
		modalPath.textContent = v.folder || "/";
		modalVid.src = "/video?path=" + encodeURIComponent(v.path);
	}
}
function applyVideoMoveLocally(srcPath, destFolder) {
	const vi = ALL_VIDEOS.findIndex((v) => v && v.path === srcPath);
	if (vi < 0) return false;
	const v = ALL_VIDEOS[vi];
	const newPath = movedVideoRelPath(srcPath, destFolder);
	const newFolder = folderFromRelPath(newPath);
	const viewingAll = activeFolder === "__all__";
	const nk = newFolder || "/";
	const leavesCurrentFolder = !viewingAll && nk != activeFolder;

	if (
		modal.classList.contains("open") &&
		currentIdx >= 0 &&
		filtered[currentIdx] &&
		filtered[currentIdx].path === srcPath &&
		leavesCurrentFolder
	) {
		closeModal();
	}

	if (selectedPaths.has(srcPath)) {
		selectedPaths.delete(srcPath);
		if (!leavesCurrentFolder) selectedPaths.add(newPath);
	}

	if (leavesCurrentFolder) {
		ALL_VIDEOS.splice(vi, 1);
		removeGalleryCardByPath(srcPath);
		recomputeFiltered();
		updateGalleryStatsMarkup();
		if (selectMode) updateSelectBar();
		return true;
	}

	const oldPath = v.path;
	v.path = newPath;
	v.folder = newFolder;
	recomputeFiltered();

	if (!filtered.includes(v)) {
		removeGalleryCardByPath(oldPath);
		updateGalleryStatsMarkup();
		if (selectMode) updateSelectBar();
		return true;
	}

	if (sortSel.value === "folder") {
		render();
		if (selectMode) updateSelectBar();
		return true;
	}

	const card = gallery.querySelector(galleryCardSelectorForPath(oldPath));
	if (card) patchVideoCardDomAfterPathChange(card, v);
	updateGalleryStatsMarkup();
	if (selectMode) updateSelectBar();
	return true;
}
async function moveVideo(srcPath, destFolder) {
	const res = await fetch("/api/move", {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify({ path: srcPath, dest_folder: destFolder }),
	});
	if (res.ok) {
		const localOk = applyVideoMoveLocally(srcPath, destFolder);
		if (!localOk) await refresh();
		else {
			try {
				await refreshFoldersOnly();
			} catch (_) {
				/* folder list best-effort */
			}
			buildFolderNav();
			populateUploadFolders();
			computeDuplicateGroups();
			updateDuplicateBanner(false);
			updateIncomingBanner();
		}
		toast("Moved to " + destFolder, "success");
	} else {
		toast("Move failed: " + (await res.text()), "error");
	}
}
function openIncomingModal() {
	const incoming = ALL_VIDEOS.filter((v) => Boolean(v.is_new));
	incomingSubtitle.textContent =
		incoming.length +
		" incoming file" +
		(incoming.length === 1 ? "" : "s");
	incomingSelectedFolder = "/";
	incomingSelectedTagIDs = new Set();
	renderIncomingTagOptions(incoming);
	renderIncomingFolderOptions();
	incomingModal.classList.add("open");
}
function closeIncomingModal() {
	incomingModal.classList.remove("open");
}
function renderIncomingFolderOptions() {
	const folders = [
		...new Set([...ALL_FOLDERS, ...ALL_VIDEOS.map((v) => v.folder || "/"), "/"]),
	].sort();
	incomingFolderOptions.innerHTML = "";
	for (const f of folders) {
		const opt = document.createElement("div");
		opt.className =
			"folder-option" + (f === incomingSelectedFolder ? " selected" : "");
		opt.innerHTML =
			'<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>' +
			(f === "/" ? "Root (/)" : escHtml(f));
		opt.addEventListener("click", () => {
			incomingSelectedFolder = f;
			renderIncomingFolderOptions();
		});
		incomingFolderOptions.appendChild(opt);
	}
}
function renderIncomingTagOptions(incoming) {
	incomingTagOptions.innerHTML = "";
	const incomingHashes = new Set(incoming.map((v) => v.hash).filter(Boolean));
	for (const tag of ALL_TAGS) {
		const selectedCount = incoming.filter(
			(v) => incomingHashes.has(v.hash) && (v.tags || []).includes(tag.id),
		).length;
		const allSelected = incoming.length > 0 && selectedCount === incoming.length;
		const row = document.createElement("label");
		row.className = "tag-option";
		row.innerHTML =
			'<input type="checkbox" ' +
			(allSelected ? "checked" : "") +
			'><span class="tag-option-name">' +
			renderTagChip(tag) +
			"</span>";
		const box = row.querySelector("input");
		box.addEventListener("change", () => {
			if (box.checked) incomingSelectedTagIDs.add(tag.id);
			else incomingSelectedTagIDs.delete(tag.id);
		});
		incomingTagOptions.appendChild(row);
	}
}
async function clearNewStatus(paths, all = false) {
	const res = await fetch("/api/videos/new/clear", {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify({ all, paths }),
	});
	return res.ok;
}
async function forgetVideos(paths, hashes = [], all = false) {
	const res = await fetch("/api/videos/new/forget", {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify({ all, paths, hashes }),
	});
	return res.ok;
}
function closeUploadModal() {
	uploadModal.classList.remove("open");
}
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
function setItemStatus(id, cls, text) {
	const el = fileQueue.querySelector(
		'[data-id="' + id + '"] .file-item-status',
	);
	if (el) {
		el.className = "file-item-status " + cls;
		el.textContent = text;
	}
}
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
		applyWatchGridClass(grid);
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
			tile.querySelector(".watch-tile-remove").addEventListener(
				"click",
				async () => {
					const ok = await setCollectionVideo(
						collection.id,
						v.hash,
						false,
					);
					if (!ok) {
						toast("Remove failed", "error");
						return;
					}
					await refresh();
					openWatchCollection(collection.id);
				},
			);
			const tileVideo = tile.querySelector("video");
			applyDefaultVideoFlags(tileVideo);
			tileVideo.volume = watchCollectionMasterVolume;
			tileVideo.muted = watchCollectionMasterMuted;
			grid.appendChild(tile);
		}
		watchCollectionBody.appendChild(grid);
	}
	syncWatchCollectionControls();
	watchCollectionModal.classList.add("open");
}
function closeWatchCollection() {
	watchCollectionModal.classList.remove("open");
	activeCollectionID = "";
	buildCollectionNav();
	watchCollectionBody.querySelectorAll("video").forEach((v) => {
		v.pause();
		v.src = "";
	});
}
function getWatchCollectionVideos() {
	return [...watchCollectionBody.querySelectorAll(".watch-grid video")];
}
function applyWatchGridClass(gridEl) {
	if (!gridEl) return;
	gridEl.classList.remove("cols-1", "cols-2", "cols-3", "cols-4");
	gridEl.classList.add("cols-" + watchGridCols);
}
function syncWatchCollectionControls() {
	if (watchCollectionVolumeInput) {
		watchCollectionVolumeInput.value = String(watchCollectionMasterVolume);
	}
	if (watchCollectionGridCols) {
		watchCollectionGridCols.value = String(watchGridCols);
	}
	if (watchCollectionMuteAllBtn) {
		watchCollectionMuteAllBtn.textContent = watchCollectionMasterMuted
			? "Unmute all"
			: "Mute all";
	}
}
function applyWatchMasterAudioToAll() {
	const videos = getWatchCollectionVideos();
	for (const video of videos) {
		video.volume = watchCollectionMasterVolume;
		video.muted = watchCollectionMasterMuted;
	}
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
function renderCollectionPickerOptions() {
	collectionPickerOptions.innerHTML = "";
	for (const collection of ALL_COLLECTIONS) {
		const option = document.createElement("div");
		option.className =
			"collection-picker-option" +
			(collection.id === collectionPickerSelectedID ? " selected" : "");
		option.innerHTML =
			"<span>" +
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
function updateSelectBar() {
	const n = selectedPaths.size;
	selectCount.textContent = n + " selected";
	selectBar.classList.toggle("visible", n > 0);
}
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
function unloadLibraryAfterPinLock() {
	closeModal();
	currentIdx = -1;
	DOWNLOADED_VIDEOS = [];
	DISCOVERED_VIDEOS = [];
	rebuildAllVideos();
	isVideoDiscoveryLoading = false;
	exitSelectMode();
	refreshDerivedUI(false);
}
function hideSidebarTagForm() {
	sidebarTagForm.classList.remove("visible");
	sidebarTagInput.value = "";
}
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
	try {
		await refreshTagsOnly();
		buildTagNav();
	} catch (e) {
		await refresh();
	}
}
function hideSidebarCollectionForm() {
	sidebarCollectionForm.classList.remove("visible");
	sidebarCollectionInput.value = "";
}
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
	try {
		await refreshCollectionsOnly();
		buildCollectionNav();
	} catch (e) {
		await refresh();
	}
}
function hideSidebarFolderForm() {
	sidebarFolderForm.classList.remove("visible");
	sidebarFolderInput.value = "";
}
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
		try {
			await refreshFoldersOnly();
			buildFolderNav();
			populateUploadFolders();
			if (moveModal.classList.contains("open")) renderFolderOptions();
		} catch (e) {
			await refresh();
		}
	} else {
		toast("Failed: " + (await res.text()), "error");
	}
}
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
async function renameVideoPath(path, name) {
	try {
		const res = await fetch("/api/rename", {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ path, name }),
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
async function setTagAssignmentsBulk(hashes, tagID, assigned) {
	try {
		const uniqueHashes = [...new Set((hashes || []).filter(Boolean))];
		if (!uniqueHashes.length) return true;
		const res = await fetch("/api/tags/assign/bulk", {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({
				hashes: uniqueHashes,
				tag_id: tagID,
				assigned,
			}),
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
async function createCollectionFromFolder(folder, name) {
	try {
		const res = await fetch("/api/collections/from-folder", {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ folder, name }),
		});
		if (!res.ok) {
			return { ok: false, error: await res.text() };
		}
		return { ok: true, collection: await res.json() };
	} catch (e) {
		return { ok: false, error: "request failed" };
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
			(group.members.find((m) => m.path === group.keepPath)?.name ||
				"selected copy") +
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
function initDomRefs() {
	gallery = document.getElementById("gallery") as typeof gallery;
	pinLockBtn = document.getElementById("pin-lock-btn") as typeof pinLockBtn;
	folderList = document.getElementById("folder-list") as typeof folderList;
	incomingList = document.getElementById("incoming-list") as typeof incomingList;
	tagList = document.getElementById("tag-list") as typeof tagList;
	searchEl = document.getElementById("search") as typeof searchEl;
	statsEl = document.getElementById("stats") as typeof statsEl;
	sortSel = document.getElementById("sort-select") as typeof sortSel;
	gridBtn = document.getElementById("grid-btn") as typeof gridBtn;
	listBtn = document.getElementById("list-btn") as typeof listBtn;
	favoritesOnlyToggle = document.getElementById("favorites-only-toggle") as typeof favoritesOnlyToggle;
	modal = document.getElementById("modal") as typeof modal;
	modalVid = document.getElementById("modal-video") as typeof modalVid;
	modalTitle = document.getElementById("modal-title") as typeof modalTitle;
	modalPath = document.getElementById("modal-path") as typeof modalPath;
	prevBtn = document.getElementById("prev-btn") as typeof prevBtn;
	nextBtn = document.getElementById("next-btn") as typeof nextBtn;
	closeBtn = document.getElementById("close-btn") as typeof closeBtn;
	selectBtn = document.getElementById("select-btn") as typeof selectBtn;
	selectBar = document.getElementById("select-bar") as typeof selectBar;
	selectCount = document.getElementById("select-count") as typeof selectCount;
	selectTagBtn = document.getElementById("select-tag-btn") as typeof selectTagBtn;
	selectMoveBtn = document.getElementById("select-move-btn") as typeof selectMoveBtn;
	selectForgetBtn = document.getElementById("select-forget-btn") as typeof selectForgetBtn;
	selectAllBtn = document.getElementById("select-all-btn") as typeof selectAllBtn;
	selectClearBtn = document.getElementById("select-clear-btn") as typeof selectClearBtn;
	newFolderBtn = document.getElementById("new-folder-btn") as typeof newFolderBtn;
	sidebarFolderForm = document.getElementById("sidebar-folder-form") as typeof sidebarFolderForm;
	sidebarFolderInput = document.getElementById("sidebar-folder-input") as typeof sidebarFolderInput;
	sidebarFolderConfirm = document.getElementById("sidebar-folder-confirm") as typeof sidebarFolderConfirm;
	sidebarFolderCancel = document.getElementById("sidebar-folder-cancel") as typeof sidebarFolderCancel;
	newTagBtn = document.getElementById("new-tag-btn") as typeof newTagBtn;
	sidebarTagForm = document.getElementById("sidebar-tag-form") as typeof sidebarTagForm;
	sidebarTagInput = document.getElementById("sidebar-tag-input") as typeof sidebarTagInput;
	sidebarTagConfirm = document.getElementById("sidebar-tag-confirm") as typeof sidebarTagConfirm;
	sidebarTagCancel = document.getElementById("sidebar-tag-cancel") as typeof sidebarTagCancel;
	newCollectionBtn = document.getElementById("new-collection-btn") as typeof newCollectionBtn;
	sidebarCollectionForm = document.getElementById("sidebar-collection-form") as typeof sidebarCollectionForm;
	sidebarCollectionInput = document.getElementById("sidebar-collection-input") as typeof sidebarCollectionInput;
	sidebarCollectionConfirm = document.getElementById("sidebar-collection-confirm") as typeof sidebarCollectionConfirm;
	sidebarCollectionCancel = document.getElementById("sidebar-collection-cancel") as typeof sidebarCollectionCancel;
	collectionList = document.getElementById("collection-list") as typeof collectionList;
	collectionsCollapseBtn = document.getElementById("collections-collapse-btn") as typeof collectionsCollapseBtn;
	tagsCollapseBtn = document.getElementById("tags-collapse-btn") as typeof tagsCollapseBtn;
	foldersCollapseBtn = document.getElementById("folders-collapse-btn") as typeof foldersCollapseBtn;
	dupBanner = document.getElementById("dup-banner") as typeof dupBanner;
	dupBannerText = document.getElementById("dup-banner-text") as typeof dupBannerText;
	dupReviewBtn = document.getElementById("dup-review-btn") as typeof dupReviewBtn;
	dupDismissBtn = document.getElementById("dup-dismiss-btn") as typeof dupDismissBtn;
	dupModal = document.getElementById("dup-modal") as typeof dupModal;
	dupSubtitle = document.getElementById("dup-subtitle") as typeof dupSubtitle;
	dupGroupsEl = document.getElementById("dup-groups") as typeof dupGroupsEl;
	dupClose = document.getElementById("dup-close") as typeof dupClose;
	dupCancelBtn = document.getElementById("dup-cancel-btn") as typeof dupCancelBtn;
	dupResolveAllBtn = document.getElementById("dup-resolve-all-btn") as typeof dupResolveAllBtn;
	incomingBanner = document.getElementById("incoming-banner") as typeof incomingBanner;
	incomingBannerText = document.getElementById("incoming-banner-text") as typeof incomingBannerText;
	incomingReviewBtn = document.getElementById("incoming-review-btn") as typeof incomingReviewBtn;
	incomingMassSortBtn = document.getElementById("incoming-mass-sort-btn") as typeof incomingMassSortBtn;
	incomingDismissBtn = document.getElementById("incoming-dismiss-btn") as typeof incomingDismissBtn;
	incomingMarkAllBtn = document.getElementById("incoming-mark-all-btn") as typeof incomingMarkAllBtn;
	incomingSidebarSection = document.getElementById("incoming-sidebar-section") as typeof incomingSidebarSection;
	incomingModal = document.getElementById("incoming-modal") as typeof incomingModal;
	incomingSubtitle = document.getElementById("incoming-subtitle") as typeof incomingSubtitle;
	incomingTagOptions = document.getElementById("incoming-tag-options") as typeof incomingTagOptions;
	incomingFolderOptions = document.getElementById("incoming-folder-options") as typeof incomingFolderOptions;
	incomingClose = document.getElementById("incoming-close") as typeof incomingClose;
	incomingCancelBtn = document.getElementById("incoming-cancel-btn") as typeof incomingCancelBtn;
	incomingConfirmBtn = document.getElementById("incoming-confirm-btn") as typeof incomingConfirmBtn;
	incomingMarkReviewedBtn = document.getElementById("incoming-mark-reviewed-btn") as typeof incomingMarkReviewedBtn;
	tagModal = document.getElementById("tag-modal") as typeof tagModal;
	tagSubtitle = document.getElementById("tag-subtitle") as typeof tagSubtitle;
	tagOptions = document.getElementById("tag-options") as typeof tagOptions;
	tagPendingSummary = document.getElementById("tag-pending-summary") as typeof tagPendingSummary;
	tagClose = document.getElementById("tag-close") as typeof tagClose;
	tagApplyBtn = document.getElementById("tag-apply-btn") as typeof tagApplyBtn;
	tagCancelBtn = document.getElementById("tag-cancel-btn") as typeof tagCancelBtn;
	tagNewInput = document.getElementById("tag-new-input") as typeof tagNewInput;
	tagNewBtn = document.getElementById("tag-new-btn") as typeof tagNewBtn;
	selectCollectionBtn = document.getElementById("select-collection-btn") as typeof selectCollectionBtn;
	collectionPickerModal = document.getElementById("collection-picker-modal") as typeof collectionPickerModal;
	collectionPickerSubtitle = document.getElementById("collection-picker-subtitle") as typeof collectionPickerSubtitle;
	collectionPickerOptions = document.getElementById("collection-picker-options") as typeof collectionPickerOptions;
	collectionPickerClose = document.getElementById("collection-picker-close") as typeof collectionPickerClose;
	collectionPickerCancelBtn = document.getElementById("collection-picker-cancel-btn") as typeof collectionPickerCancelBtn;
	collectionPickerConfirmBtn = document.getElementById("collection-picker-confirm-btn") as typeof collectionPickerConfirmBtn;
	collectionPickerNewInput = document.getElementById("collection-picker-new-input") as typeof collectionPickerNewInput;
	collectionPickerNewBtn = document.getElementById("collection-picker-new-btn") as typeof collectionPickerNewBtn;
	watchCollectionModal = document.getElementById("watch-collection-modal") as typeof watchCollectionModal;
	watchCollectionTitle = document.getElementById("watch-collection-title") as typeof watchCollectionTitle;
	watchCollectionSubtitle = document.getElementById("watch-collection-subtitle") as typeof watchCollectionSubtitle;
	watchCollectionBody = document.getElementById("watch-collection-body") as typeof watchCollectionBody;
	watchCollectionClose = document.getElementById("watch-collection-close") as typeof watchCollectionClose;
	watchCollectionDeleteBtn = document.getElementById("watch-collection-delete-btn") as typeof watchCollectionDeleteBtn;
	watchCollectionRenameBtn = document.getElementById("watch-collection-rename-btn") as typeof watchCollectionRenameBtn;
	watchCollectionPlayAllBtn = document.getElementById("watch-collection-play-all-btn") as typeof watchCollectionPlayAllBtn;
	watchCollectionPauseAllBtn = document.getElementById("watch-collection-pause-all-btn") as typeof watchCollectionPauseAllBtn;
	watchCollectionMuteAllBtn = document.getElementById("watch-collection-mute-all-btn") as typeof watchCollectionMuteAllBtn;
	watchCollectionVolumeInput = document.getElementById("watch-collection-volume-input") as typeof watchCollectionVolumeInput;
	watchCollectionGridCols = document.getElementById("watch-collection-grid-cols") as typeof watchCollectionGridCols;
	previewSettingsModal = document.getElementById("preview-settings-modal") as typeof previewSettingsModal;
	previewSettingsClose = document.getElementById("preview-settings-close") as typeof previewSettingsClose;
	previewSettingsLauncher = document.getElementById("preview-settings-launcher") as typeof previewSettingsLauncher;
	previewSettingsAudioToggleBtn = document.getElementById("preview-settings-audio-toggle") as typeof previewSettingsAudioToggleBtn;
	previewSettingsVolumeInput = document.getElementById("preview-settings-volume") as typeof previewSettingsVolumeInput;
	pinGateOverlay = document.getElementById("pin-gate-overlay") as typeof pinGateOverlay;
	pinGateForm = document.getElementById("pin-gate-form") as typeof pinGateForm;
	pinGateInput = document.getElementById("pin-gate-input") as typeof pinGateInput;
	pinGateError = document.getElementById("pin-gate-error") as typeof pinGateError;
	ctxMenu = document.getElementById("ctx-menu") as typeof ctxMenu;
	ctxTagsEl = document.getElementById("ctx-tags") as typeof ctxTagsEl;
	ctxMoveEl = document.getElementById("ctx-move") as typeof ctxMoveEl;
	ctxRenameEl = document.getElementById("ctx-rename") as typeof ctxRenameEl;
	ctxDeleteEl = document.getElementById("ctx-delete") as typeof ctxDeleteEl;
	ctxForgetEl = document.getElementById("ctx-forget") as typeof ctxForgetEl;
	moveModal = document.getElementById("move-modal") as typeof moveModal;
	moveSubtitle = document.getElementById("move-subtitle") as typeof moveSubtitle;
	folderOptions = document.getElementById("folder-options") as typeof folderOptions;
	newFolderName = document.getElementById("new-folder-name") as typeof newFolderName;
	addFolderBtn = document.getElementById("add-folder-btn") as typeof addFolderBtn;
	moveCancelBtn = document.getElementById("move-cancel-btn") as typeof moveCancelBtn;
	moveConfirmBtn = document.getElementById("move-confirm-btn") as typeof moveConfirmBtn;
	moveClose = document.getElementById("move-close") as typeof moveClose;
	uploadModal = document.getElementById("upload-modal") as typeof uploadModal;
	dropZone = document.getElementById("drop-zone") as typeof dropZone;
	uploadInput = document.getElementById("upload-input") as typeof uploadInput;
	uploadFolderSel = document.getElementById("upload-folder-select") as typeof uploadFolderSel;
	newFolderInput = document.getElementById("new-folder-input") as typeof newFolderInput;
	fileQueue = document.getElementById("file-queue") as typeof fileQueue;
	uploadSummary = document.getElementById("upload-summary") as typeof uploadSummary;
	uploadStartBtn = document.getElementById("upload-start-btn") as typeof uploadStartBtn;
	uploadCancelBtn = document.getElementById("upload-cancel-btn") as typeof uploadCancelBtn;
	uploadClose = document.getElementById("upload-close") as typeof uploadClose;
	openUploadBtn = document.getElementById("open-upload-btn") as typeof openUploadBtn;
	pinLockBtn = document.getElementById("pin-lock-btn");
	if (tagsCollapseBtn) tagsSection = tagsCollapseBtn.closest(".sidebar-section");
	if (foldersCollapseBtn) foldersSection = foldersCollapseBtn.closest(".sidebar-section");
	if (collectionsCollapseBtn) collectionsSection = collectionsCollapseBtn.closest(".sidebar-section");
}

function wireAppEvents() {

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
		if (
			e.key === "Escape" &&
			previewSettingsModal &&
			previewSettingsModal.classList.contains("open")
		) {
			closePreviewSettings();
			return;
		}
		if (!modal.classList.contains("open")) return;
		if (e.key === "Escape") closeModal();
		if (e.key === "ArrowLeft") prevBtn.click();
		if (e.key === "ArrowRight") nextBtn.click();
	});
	
	// context menu
	
	document.addEventListener("click", () => ctxMenu.classList.remove("open"));
	ctxTagsEl.addEventListener("click", () => {
		if (ctxVideo) openTagModal(ctxVideo);
	});
	ctxMoveEl.addEventListener("click", () => {
		if (ctxVideo) openMoveModal(ctxVideo);
	});
	ctxRenameEl.addEventListener("click", async () => {
		if (!ctxVideo) return;
		const next = prompt("Rename file (new base name)", ctxVideo.name);
		if (next == null) return;
		const trimmed = next.trim();
		if (!trimmed) return;
		const res = await renameVideoPath(ctxVideo.path, trimmed);
		if (res.ok) {
			toast("Renamed file", "success");
			await refresh();
		} else {
			toast("Rename failed: " + res.message, "error");
		}
	});
	ctxForgetEl.addEventListener("click", async () => {
		if (!ctxVideo) return;
		const ok = await forgetVideos([ctxVideo.path], [], false);
		if (!ok) {
			toast("Failed to forget file", "error");
			return;
		}
		toast("File forgotten and re-queued", "success");
		await refresh();
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

	tagNewBtn.addEventListener("click", async () => {
		const name = tagNewInput.value.trim();
		if (!name) return;
		const created = await createTag(name);
		if (!created) {
			toast("Tag create failed", "error");
			return;
		}
		tagNewInput.value = "";
		try {
			await refreshTagsOnly();
			buildTagNav();
		} catch (e) {
			await refresh();
		}
		renderTagOptions();
	});
	tagNewInput.addEventListener("keydown", (e) => {
		if (e.key === "Enter") tagNewBtn.click();
	});
	tagClose.addEventListener("click", closeTagModal);
	tagApplyBtn.addEventListener("click", async () => {
		const entries = getPendingTagActionEntries();
		if (!entries.length) return;
		const liveTargets = resolveTagTargets();
		const hashes = [...new Set(liveTargets.map((v) => v.hash).filter(Boolean))];
		if (!hashes.length) {
			toast("No taggable videos selected", "error");
			return;
		}
	
		const addingNames = entries
			.filter(([, action]) => action === "add")
			.map(([tagID]) => TAG_MAP[tagID]?.name || tagID);
		const removingNames = entries
			.filter(([, action]) => action === "remove")
			.map(([tagID]) => TAG_MAP[tagID]?.name || tagID);
		const confirmLines = [
			"Apply tag changes to " + hashes.length + " videos?",
		];
		if (addingNames.length) confirmLines.push("Add: " + addingNames.join(", "));
		if (removingNames.length)
			confirmLines.push("Remove: " + removingNames.join(", "));
		if (!confirm(confirmLines.join("\n"))) return;
	
		recomputeFiltered();
		const prevFilteredPaths = filtered.map((v) => v.path);
	
		let failures = 0;
		for (const [tagID, action] of entries) {
			const ok = await setTagAssignmentsBulk(hashes, tagID, action === "add");
			if (!ok) failures++;
		}
		if (failures) {
			toast(failures + " tag update(s) failed", "error");
			await refresh();
			renderTagOptions();
			return;
		}
	
		try {
			await refreshTagsOnly();
			await refreshTagAssignmentsForHashes(hashes);
			refreshAfterTagDataChange(prevFilteredPaths, hashes);
		} catch (e) {
			await refresh();
		}
		pendingTagActions = new Map();
		renderTagOptions();
		toast("Tag changes applied", "success");
		if (selectMode) {
			closeTagModal();
			exitSelectMode();
		}
	});
	tagCancelBtn.addEventListener("click", closeTagModal);
	tagModal.addEventListener("click", (e) => {
		if (e.target === tagModal) closeTagModal();
	});
	
	// move modal

	/** Rebuilds the folder option list inside the move modal, reflecting the
	 *  current ALL_FOLDERS / ALL_VIDEOS state and highlighting moveSelectedFolder. */
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

	// upload modal

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
	
	/** Updates the upload summary line and enables/disables the Upload button
	 *  based on the number of valid (non-error) files in uploadFiles. */
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
				const res = await fetch("/api/upload", {
					method: "POST",
					body: fd,
				});
				const results = await res.json();
				const r = results[0];
				if (r && r.error) {
					setItemStatus(item.id, "error", r.error);
				} else {
					setItemStatus(item.id, "done", "✓ done");
					item.status = "done";
					const guessedFolder =
						dest && dest !== "/" && dest !== "__new__" ? dest : "/";
					const extRaw = item.file.name.includes(".")
						? "." + item.file.name.split(".").pop().toLowerCase()
						: "";
					upsertDownloadedVideo({
						name: item.file.name,
						path:
							guessedFolder === "/"
								? item.file.name
								: guessedFolder + "/" + item.file.name,
						folder: guessedFolder,
						ext: extRaw,
						size: item.file.size,
						modified: new Date().toISOString(),
						hash: "",
						is_favorite: false,
						tags: [],
					});
					refreshDerivedUI(false);
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
	favoritesOnlyToggle.addEventListener("change", () => {
		showFavoritesOnly = favoritesOnlyToggle.checked;
		render();
	});
	
	// select mode
	/** Syncs the floating select bar's count label and visibility with selectedPaths. */

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
		const first =
			filtered.find((v) => selectedPaths.has(v.path)) || filtered[0];
		openMoveModal(first);
	});
	selectForgetBtn.addEventListener("click", async () => {
		if (!selectedPaths.size) return;
		const paths = [...selectedPaths];
		const ok = await forgetVideos(paths, [], false);
		if (!ok) {
			toast("Forget operation failed", "error");
			return;
		}
		await refresh();
		exitSelectMode();
		toast("Forgot " + paths.length + " file(s)", "success");
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
		tagsSection.classList.remove("collapsed");
		sidebarTagForm.classList.add("visible");
		sidebarTagInput.value = "";
		sidebarTagInput.focus();
	});
	sidebarTagCancel.addEventListener("click", hideSidebarTagForm);
	sidebarTagConfirm.addEventListener("click", createSidebarTag);
	sidebarTagInput.addEventListener("keydown", (e) => {
		if (e.key === "Enter") createSidebarTag();
		if (e.key === "Escape") hideSidebarTagForm();
	});
	
	// collections (sidebar)
	newCollectionBtn.addEventListener("click", () => {
		collectionsSection.classList.remove("collapsed");
		sidebarCollectionForm.classList.add("visible");
		sidebarCollectionInput.value = "";
		sidebarCollectionInput.focus();
	});
	sidebarCollectionCancel.addEventListener("click", hideSidebarCollectionForm);
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
		foldersSection.classList.remove("collapsed");
		sidebarFolderForm.classList.add("visible");
		sidebarFolderInput.value = "";
		sidebarFolderInput.focus();
	});
	sidebarFolderCancel.addEventListener("click", hideSidebarFolderForm);
	sidebarFolderConfirm.addEventListener("click", createFolder);
	sidebarFolderInput.addEventListener("keydown", (e) => {
		if (e.key === "Enter") createFolder();
		if (e.key === "Escape") hideSidebarFolderForm();
	});

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
	
	incomingReviewBtn.addEventListener("click", () => {
		activeIncomingFilter = true;
		activeFolder = "__all__";
		buildIncomingNav();
		buildFolderNav();
		render();
	});
	incomingMassSortBtn.addEventListener("click", openIncomingModal);
	incomingDismissBtn.addEventListener("click", () => {
		incomingBannerDismissed = true;
		incomingBanner.classList.remove("visible");
	});
	incomingClose.addEventListener("click", closeIncomingModal);
	incomingCancelBtn.addEventListener("click", closeIncomingModal);
	incomingModal.addEventListener("click", (e) => {
		if (e.target === incomingModal) closeIncomingModal();
	});
	incomingMarkAllBtn.addEventListener("click", async () => {
		const incomingPaths = ALL_VIDEOS.filter((v) => Boolean(v.is_new)).map((v) => v.path);
		if (!incomingPaths.length) return;
		const ok = await clearNewStatus(incomingPaths, false);
		if (!ok) {
			toast("Failed to mark incoming files as reviewed", "error");
			return;
		}
		await refresh();
		toast("Marked all incoming as reviewed", "success");
	});
	incomingMarkReviewedBtn.addEventListener("click", async () => {
		const incoming = ALL_VIDEOS.filter((v) => Boolean(v.is_new)).map((v) => v.path);
		if (!incoming.length) {
			closeIncomingModal();
			return;
		}
		const ok = await clearNewStatus(incoming, true);
		if (!ok) {
			toast("Failed to mark incoming files as reviewed", "error");
			return;
		}
		closeIncomingModal();
		await refresh();
	});
	incomingConfirmBtn.addEventListener("click", async () => {
		const incoming = ALL_VIDEOS.filter((v) => Boolean(v.is_new));
		if (!incoming.length) {
			closeIncomingModal();
			return;
		}
		let errs = 0;
		const reviewedPaths = [];
		for (const video of incoming) {
			if ((video.folder || "/") === incomingSelectedFolder) {
				reviewedPaths.push(video.path);
				continue;
			}
			const res = await fetch("/api/move", {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({
					path: video.path,
					dest_folder: incomingSelectedFolder,
				}),
			});
			if (!res.ok) errs++;
			else reviewedPaths.push(video.path);
		}
		const incomingHashes = [...new Set(incoming.map((v) => v.hash).filter(Boolean))];
		let tagErrs = 0;
		if (incomingSelectedTagIDs.size && incomingHashes.length) {
			for (const tagID of incomingSelectedTagIDs) {
				const ok = await setTagAssignmentsBulk(incomingHashes, tagID, true);
				if (!ok) tagErrs++;
			}
		}
		const ok = await clearNewStatus(reviewedPaths, false);
		if (!ok) {
			toast("Moves completed but clear status failed", "error");
		}
		if (ok && !errs && !tagErrs) {
			activeIncomingFilter = false;
		}
		closeIncomingModal();
		await refresh();
		if (errs || tagErrs) {
			toast(
				(errs ? errs + " move(s) failed. " : "") +
					(tagErrs ? tagErrs + " tag update(s) failed." : ""),
				"error",
			);
		} else {
			toast("Incoming media mass sorted", "success");
		}
	});
	
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
		try {
			await refreshCollectionsOnly();
			buildCollectionNav();
		} catch (e) {
			await refresh();
		}
		collectionPickerSelectedID = created.id;
		renderCollectionPickerOptions();
	});
	collectionPickerNewInput.addEventListener("keydown", (e) => {
		if (e.key === "Enter") collectionPickerNewBtn.click();
	});
	collectionPickerConfirmBtn.addEventListener("click", async () => {
		if (!collectionPickerSelectedID || !collectionPickerTargetHashes.length)
			return;
		const collectionId = collectionPickerSelectedID;
		const ok = await bulkAddCollectionVideos(
			collectionId,
			collectionPickerTargetHashes,
		);
		if (!ok) {
			toast("Add to collection failed", "error");
			return;
		}
		try {
			await refreshCollectionsOnly();
			buildCollectionNav();
			if (activeCollectionID === collectionId) openWatchCollection(collectionId);
		} catch (e) {
			await refresh();
		}
		closeCollectionPicker();
		if (selectMode) exitSelectMode();
		toast("Added to collection", "success");
	});
	
	watchCollectionClose.addEventListener("click", closeWatchCollection);
	watchCollectionPlayAllBtn.addEventListener("click", () => {
		const videos = getWatchCollectionVideos();
		for (const video of videos) {
			video.play().catch(() => {});
		}
	});
	watchCollectionPauseAllBtn.addEventListener("click", () => {
		const videos = getWatchCollectionVideos();
		for (const video of videos) {
			video.pause();
		}
	});
	watchCollectionMuteAllBtn.addEventListener("click", () => {
		watchCollectionMasterMuted = !watchCollectionMasterMuted;
		persistStorageValue(WATCH_GRID_MUTED_STORAGE_KEY, watchCollectionMasterMuted);
		applyWatchMasterAudioToAll();
		syncWatchCollectionControls();
	});
	watchCollectionVolumeInput.addEventListener("input", () => {
		watchCollectionMasterVolume = Math.max(
			0,
			Math.min(1, Number(watchCollectionVolumeInput.value) || 0),
		);
		persistStorageValue(WATCH_GRID_VOLUME_STORAGE_KEY, watchCollectionMasterVolume);
		applyWatchMasterAudioToAll();
	});
	watchCollectionGridCols.addEventListener("change", () => {
		watchGridCols = Math.max(
			1,
			Math.min(4, Number(watchCollectionGridCols.value) || 4),
		);
		persistStorageValue(WATCH_GRID_COLS_STORAGE_KEY, watchGridCols);
		const grid = watchCollectionBody.querySelector(".watch-grid");
		applyWatchGridClass(grid);
		syncWatchCollectionControls();
	});
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
	
	if (previewSettingsLauncher) {
		previewSettingsLauncher.addEventListener("click", () => {
			openPreviewSettings();
		});
	}
	if (previewSettingsClose) {
		previewSettingsClose.addEventListener("click", () => {
			closePreviewSettings();
		});
	}
	if (previewSettingsModal) {
		previewSettingsModal.addEventListener("click", (e) => {
			if (e.target === previewSettingsModal) closePreviewSettings();
		});
	}
	if (previewSettingsAudioToggleBtn) {
		previewSettingsAudioToggleBtn.addEventListener("click", () => {
			toggleHoverPreviewAudio();
		});
	}
	if (previewSettingsVolumeInput) {
		previewSettingsVolumeInput.addEventListener("input", () => {
			hoverPreviewVolume = Math.max(
				0,
				Math.min(1, Number(previewSettingsVolumeInput.value) || 0),
			);
			persistStorageValue(HOVER_PREVIEW_VOLUME_KEY, hoverPreviewVolume);
			applyHoverPreviewAudioToAll();
		});
	}
	
	if (pinGateForm) {
		pinGateForm.addEventListener("submit", async (e) => {
			e.preventDefault();
			const pin = (pinGateInput && pinGateInput.value) || "";
			if (pinGateError) {
				pinGateError.hidden = true;
				pinGateError.textContent = "";
			}
			try {
				const res = await fetch("/api/auth/pin", {
					method: "POST",
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify({ pin }),
				});
				let j = {};
				try {
					j = await res.json();
				} catch {
					/* ignore */
				}
				if (!res.ok || !j.ok) {
					setPinOverlayVisible(true, "Incorrect PIN.");
					if (pinGateInput) pinGateInput.select();
					return;
				}
				pinGateUnlocked = true;
				setPinOverlayVisible(false);
				updatePinLockButton();
				await init();
			} catch (err) {
				console.error(err);
				setPinOverlayVisible(true, "Could not reach server.");
			}
		});
	}
	if (pinLockBtn) {
		pinLockBtn.addEventListener("click", async () => {
			try {
				const res = await fetch("/api/auth/lock", { method: "POST" });
				if (!res.ok) return;
				unloadLibraryAfterPinLock();
				pinGateUnlocked = false;
				updatePinLockButton();
				if (pinGateInput) pinGateInput.value = "";
				setPinOverlayVisible(true);
			} catch (e) {
				console.error(e);
			}
		});
	}
	
}

	onMount(() => {
	initDomRefs();
	setOnPinLock(() => {
		setPinGateUnlocked(false);
		pinGateUnlocked = false;
		updatePinLockButton();
		unloadLibraryAfterPinLock();
		setPinOverlayVisible(true, 'Session locked or PIN required.');
	});
	wireAppEvents();
	bootstrap();
});
</script>

<div id="pin-gate-overlay" class="pin-gate-overlay" aria-hidden="true" hidden>
  <div class="pin-gate-dialog" role="dialog" aria-labelledby="pin-gate-title">
    <div class="logo pin-gate-logo">VID<span>VAULT</span></div>
    <p id="pin-gate-title" class="pin-gate-title">Enter PIN to continue</p>
    <form id="pin-gate-form" class="pin-gate-form" autocomplete="off" data-lpignore="true" data-form-type="other">
      <input type="password" id="pin-gate-input" class="pin-gate-input" name="vv_pin" autocomplete="off" autocorrect="off" autocapitalize="off" spellcheck="false" inputmode="numeric" aria-label="PIN" data-lpignore="true" data-1p-ignore data-bwignore data-form-type="other">
      <p id="pin-gate-error" class="pin-gate-error" hidden></p>
      <button type="submit" class="btn pin-gate-submit">Unlock</button>
    </form>
  </div>
</div>

<header>
  <div class="logo">VID<span>VAULT</span></div>
  <div class="stats" id="stats"></div>
  <div class="controls">
    <label class="favorites-only-toggle" title="Show favorites only">
      <span class="favorites-switch">
        <input type="checkbox" id="favorites-only-toggle" class="favorites-switch-input">
        <span class="favorites-switch-ui" aria-hidden="true">
          <span class="favorites-switch-track"></span>
          <span class="favorites-switch-thumb">
            <svg class="favorites-switch-star" xmlns="http://www.w3.org/2000/svg" width="11" height="11" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M12 2l2.4 7.4h7.6l-6 4.6 2.3 7-6.3-4.6-6.3 4.6 2.3-7-6-4.6h7.6z"/></svg>
          </span>
        </span>
      </span>
      <span class="favorites-only-label">Favorites only</span>
    </label>
    <div class="search-wrap">
      <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/></svg>
      <input id="search" type="search" placeholder="filter files…" autocomplete="off">
    </div>
    <select class="btn" id="sort-select">
      <option value="name">name ↑</option>
      <option value="name-desc">name ↓</option>
      <option value="folder">folder</option>
      <option value="ext">type</option>
      <option value="modified-desc">modified (newest)</option>
      <option value="modified">modified (oldest)</option>
      <option value="size-desc">size (largest)</option>
      <option value="size">size (smallest)</option>
    </select>
    <div class="btn-group">
      <button class="btn active" id="grid-btn" title="Grid view">⊞</button>
      <button class="btn"        id="list-btn" title="List view">☰</button>
    </div>
    <button class="btn" id="select-btn" title="Select multiple">Select</button>
    <button type="button" class="btn pin-lock-btn" id="pin-lock-btn" title="Lock session" hidden>
      <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
      Lock
    </button>
    <button class="btn upload-btn" id="open-upload-btn">
      <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
      Upload
    </button>
  </div>
</header>

<div class="dup-banner" id="dup-banner">
  <span class="dup-banner-text" id="dup-banner-text"></span>
  <div class="dup-banner-actions">
    <button class="btn active" id="dup-review-btn">Review duplicates</button>
    <button class="btn" id="dup-dismiss-btn">Dismiss</button>
  </div>
</div>

<div class="incoming-banner" id="incoming-banner">
  <span class="incoming-banner-text" id="incoming-banner-text"></span>
  <div class="incoming-banner-actions">
    <button class="btn" id="incoming-review-btn">View</button>
    <button class="btn active" id="incoming-mass-sort-btn">Mass sort</button>
    <button class="btn" id="incoming-dismiss-btn">Dismiss</button>
  </div>
</div>

<div class="shell">
  <aside>
    <div class="sidebar-section" id="incoming-sidebar-section">
      <div class="sidebar-header">
        <span class="sidebar-label">Incoming</span>
      </div>
      <div id="incoming-list"></div>
      <div class="sidebar-inline-actions">
        <button class="btn" id="incoming-mark-all-btn">Mark as resolved</button>
      </div>
    </div>

    <div class="sidebar-section">
      <div class="sidebar-header" id="tags-header">
        <span class="sidebar-label">Tags</span>
        <div class="sidebar-actions">
          <button class="new-folder-btn" id="new-tag-btn" title="New tag">+</button>
          <button class="collapse-btn" id="tags-collapse-btn" title="Collapse tags">▾</button>
        </div>
      </div>
      <div class="sidebar-tag-form" id="sidebar-tag-form">
        <input id="sidebar-tag-input" placeholder="tag name…" autocomplete="off">
        <button id="sidebar-tag-confirm" title="Create">✓</button>
        <button id="sidebar-tag-cancel" title="Cancel">✕</button>
      </div>
      <div id="tag-list"></div>
    </div>

    <div class="sidebar-section">
      <div class="sidebar-header" id="collections-header">
        <span class="sidebar-label">Collections</span>
        <div class="sidebar-actions">
          <button class="new-folder-btn" id="new-collection-btn" title="New collection">+</button>
          <button class="collapse-btn" id="collections-collapse-btn" title="Collapse collections">▾</button>
        </div>
      </div>
      <div class="sidebar-collection-form" id="sidebar-collection-form">
        <input id="sidebar-collection-input" placeholder="collection name…" autocomplete="off">
        <button id="sidebar-collection-confirm" title="Create">✓</button>
        <button id="sidebar-collection-cancel" title="Cancel">✕</button>
      </div>
      <div id="collection-list"></div>
    </div>

    <div class="sidebar-section">
      <div class="sidebar-header">
        <span class="sidebar-label">Folders</span>
        <div class="sidebar-actions">
          <button class="new-folder-btn" id="new-folder-btn" title="New folder">+</button>
          <button class="collapse-btn" id="folders-collapse-btn" title="Collapse folders">▾</button>
        </div>
      </div>
      <div class="sidebar-folder-form" id="sidebar-folder-form">
        <input id="sidebar-folder-input" placeholder="folder name…" autocomplete="off">
        <button id="sidebar-folder-confirm" title="Create">✓</button>
        <button id="sidebar-folder-cancel" title="Cancel">✕</button>
      </div>
      <div id="folder-list"></div>
    </div>

    <button type="button" class="btn preview-settings-launcher" id="preview-settings-launcher" title="Preview settings" aria-label="Preview settings">
      <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
      <span class="preview-settings-launcher-label">settings</span>
    </button>
  </aside>
  <main><div id="gallery"></div></main>
</div>

<div id="ctx-menu">
  <div class="ctx-item" id="ctx-tags">
    <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20.59 13.41 11 3H4v7l9.59 9.59a2 2 0 0 0 2.82 0l4.18-4.18a2 2 0 0 0 0-2.82z"/><path d="M7 7h.01"/></svg>
    Edit tags…
  </div>
  <div class="ctx-item" id="ctx-move">
    <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 9l4-4 4 4"/><path d="M9 5v14"/><path d="M19 15l-4 4-4-4"/><path d="M15 19V5"/></svg>
    Move to folder…
  </div>
  <div class="ctx-item" id="ctx-rename">
    <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 20h9"/><path d="M16.5 3.5a2.12 2.12 0 0 1 3 3L7 19l-4 1 1-4Z"/></svg>
    Rename file…
  </div>
  <div class="ctx-item danger" id="ctx-delete">
    <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 6h18"/><path d="M8 6V4h8v2"/><path d="M19 6l-1 14H6L5 6"/><path d="M10 11v6"/><path d="M14 11v6"/></svg>
    Delete file…
  </div>
  <div class="ctx-item" id="ctx-forget">
    <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 12h18"/><path d="M12 3v18"/></svg>
    Forget (mark as new)…
  </div>
</div>

<div id="move-modal">
  <div class="move-inner">
    <div class="move-header">
      <div><div class="move-title">Move to folder</div><div class="move-subtitle" id="move-subtitle"></div></div>
      <button class="close-btn" id="move-close">✕</button>
    </div>
    <div class="move-body">
      <div class="move-new-folder">
        <input id="new-folder-name" placeholder="+ new folder name…" autocomplete="off">
        <button class="btn" id="add-folder-btn">Create</button>
      </div>
      <div id="folder-options"></div>
    </div>
    <div class="move-footer">
      <button class="btn" id="move-cancel-btn">Cancel</button>
      <button class="btn active" id="move-confirm-btn">Move</button>
    </div>
  </div>
</div>

<div id="upload-modal">
  <div class="upload-inner">
    <div class="upload-header">
      <div class="upload-title">Upload videos</div>
      <button class="close-btn" id="upload-close">✕</button>
    </div>
    <div class="upload-body">
      <div class="drop-zone" id="drop-zone">
        <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
        <strong>Drop videos here</strong>
        <span>or click to browse — webm, mp4, mkv, mov, avi…</span>
        <input type="file" id="upload-input" multiple accept=".webm,.mp4,.mpv,.mkv,.mov,.avi,.m4v,.ogv">
      </div>
      <div class="upload-folder-row">
        <label>Destination</label>
        <select id="upload-folder-select"></select>
        <input id="new-folder-input" placeholder="new/folder/name">
      </div>
      <div class="file-queue" id="file-queue"></div>
    </div>
    <div class="upload-footer">
      <span class="upload-summary" id="upload-summary"></span>
      <div style="display:flex;gap:8px">
        <button class="btn" id="upload-cancel-btn">Cancel</button>
        <button class="btn active" id="upload-start-btn" disabled>Upload</button>
      </div>
    </div>
  </div>
</div>

<div id="tag-modal">
  <div class="tag-inner">
    <div class="tag-header">
      <div>
        <div class="tag-title">Edit tags</div>
        <div class="tag-subtitle" id="tag-subtitle"></div>
      </div>
      <button class="close-btn" id="tag-close">✕</button>
    </div>
    <div class="tag-body">
      <div class="tag-new-row">
        <input id="tag-new-input" placeholder="+ new tag name…" autocomplete="off">
        <button class="btn" id="tag-new-btn">Create</button>
      </div>
      <div id="tag-options"></div>
    </div>
    <div class="tag-footer">
      <span class="tag-pending-summary" id="tag-pending-summary"></span>
      <button class="btn active" id="tag-apply-btn" disabled>Apply changes</button>
      <button class="btn" id="tag-cancel-btn">Close</button>
    </div>
  </div>
</div>

<div id="dup-modal">
  <div class="dup-inner">
    <div class="dup-header">
      <div>
        <div class="dup-title">Duplicate manager</div>
        <div class="dup-subtitle" id="dup-subtitle"></div>
      </div>
      <button class="close-btn" id="dup-close">✕</button>
    </div>
    <div class="dup-body" id="dup-groups"></div>
    <div class="dup-footer">
      <button class="btn" id="dup-cancel-btn">Close</button>
      <button class="btn active" id="dup-resolve-all-btn">Resolve all groups</button>
    </div>
  </div>
</div>

<div id="incoming-modal">
  <div class="move-inner">
    <div class="move-header">
      <div>
        <div class="move-title">Mass sort incoming media</div>
        <div class="move-subtitle" id="incoming-subtitle"></div>
      </div>
      <button class="close-btn" id="incoming-close">✕</button>
    </div>
    <div class="move-body">
      <div>
        <div class="sidebar-label" style="padding-bottom:6px">Add tags to all incoming</div>
        <div id="incoming-tag-options"></div>
      </div>
      <div>
        <div class="sidebar-label" style="padding-bottom:6px">Move all incoming to folder</div>
        <div id="incoming-folder-options"></div>
      </div>
    </div>
    <div class="move-footer">
      <button class="btn" id="incoming-cancel-btn">Close</button>
      <button class="btn active" id="incoming-confirm-btn">Apply mass sort</button>
      <button class="btn" id="incoming-mark-reviewed-btn">Mark as resolved</button>
    </div>
  </div>
</div>

<div id="modal">
  <div class="modal-inner">
    <div class="modal-header">
      <span class="modal-title" id="modal-title"></span>
      <span class="modal-path"  id="modal-path"></span>
      <div class="modal-nav">
        <button class="nav-btn" id="prev-btn">‹</button>
        <button class="nav-btn" id="next-btn">›</button>
      </div>
      <button class="close-btn" id="close-btn">✕</button>
    </div>
    <video id="modal-video" controls autoplay loop></video>
  </div>
</div>

<div class="select-bar" id="select-bar">
  <span class="select-count" id="select-count">0 selected</span>
  <button class="btn" id="select-tag-btn">Tags…</button>
  <button class="btn" id="select-collection-btn">Add to collection…</button>
  <button class="btn active" id="select-move-btn">Move…</button>
  <button class="btn" id="select-forget-btn">Forget…</button>
  <button class="btn" id="select-all-btn">All</button>
  <button class="btn" id="select-clear-btn">Clear</button>
</div>

<div id="collection-picker-modal">
  <div class="collection-picker-inner">
    <div class="collection-picker-header">
      <div>
        <div class="collection-picker-title">Add to collection</div>
        <div class="collection-picker-subtitle" id="collection-picker-subtitle"></div>
      </div>
      <button class="close-btn" id="collection-picker-close">✕</button>
    </div>
    <div class="collection-picker-body">
      <div class="collection-new-row">
        <input id="collection-picker-new-input" placeholder="+ new collection name…" autocomplete="off">
        <button class="btn" id="collection-picker-new-btn">Create</button>
      </div>
      <div id="collection-picker-options"></div>
    </div>
    <div class="collection-picker-footer">
      <button class="btn" id="collection-picker-cancel-btn">Cancel</button>
      <button class="btn active" id="collection-picker-confirm-btn">Add selected</button>
    </div>
  </div>
</div>

<div id="watch-collection-modal">
  <div class="watch-collection-inner">
    <div class="watch-collection-header">
      <div>
        <div class="watch-collection-title" id="watch-collection-title"></div>
        <div class="watch-collection-subtitle" id="watch-collection-subtitle"></div>
      </div>
      <div class="watch-collection-actions">
        <button class="btn" id="watch-collection-play-all-btn">Play all</button>
        <button class="btn" id="watch-collection-pause-all-btn">Pause all</button>
        <button class="btn" id="watch-collection-mute-all-btn">Mute all</button>
        <label class="watch-volume-control" for="watch-collection-volume-input">
          Vol
          <input id="watch-collection-volume-input" type="range" min="0" max="1" step="0.05" value="1">
        </label>
        <label class="watch-grid-control" for="watch-collection-grid-cols">
          Grid
          <select class="btn" id="watch-collection-grid-cols">
            <option value="4">4</option>
            <option value="3">3</option>
            <option value="2">2</option>
            <option value="1">1</option>
          </select>
        </label>
        <button class="btn" id="watch-collection-rename-btn">Rename</button>
        <button class="btn" id="watch-collection-delete-btn">Delete</button>
        <button class="close-btn" id="watch-collection-close">✕</button>
      </div>
    </div>
    <div class="watch-collection-body" id="watch-collection-body"></div>
  </div>
</div>

<div id="preview-settings-modal">
  <div class="preview-settings-inner">
    <div class="preview-settings-header">
      <div class="preview-settings-title">Preview settings</div>
      <button type="button" class="close-btn" id="preview-settings-close" aria-label="Close">✕</button>
    </div>
    <div class="preview-settings-body">
      <div class="preview-settings-row">
        <span class="preview-settings-label">Hover preview audio</span>
        <button type="button" class="btn" id="preview-settings-audio-toggle">Off</button>
      </div>
      <label class="preview-settings-row preview-settings-volume-row" for="preview-settings-volume">
        <span class="preview-settings-label">Preview volume</span>
        <input id="preview-settings-volume" type="range" min="0" max="1" step="0.05" value="0.5">
      </label>
    </div>
  </div>
</div>


