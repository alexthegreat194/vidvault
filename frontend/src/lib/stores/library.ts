import { derived, get, writable } from 'svelte/store';
import * as api from '../api';
import {
	fetchAuthStatus,
	fetchCollections,
	fetchFolders,
	fetchTags,
	fetchTagAssignmentsForHashes,
	fetchVideosPage,
	streamVideos,
} from '../api';
import { toast } from '../toast';
import { folderFromRelPath, movedVideoRelPath } from '../video-helpers';
import type {
	Collection,
	DuplicateGroup,
	FolderInfo,
	ScanProgress,
	Tag,
	UploadBatch,
	Video,
} from '../types';
import {
	installFetch401Handler,
	setPinGateMode,
	setPinGateUnlocked,
} from '../auth';
import {
	autoOpenedDuplicates,
	dupModalOpen,
	duplicateBannerDismissed,
	incomingBannerDismissed,
	incomingModalBatchId,
	incomingModalOpen,
} from './ui';
import { closePlayer, playingPath } from './player';
import { exitSelectMode, selectedPaths } from './selection';

export interface FilterState {
	search: string;
	sort: string;
	activeFolder: string;
	activeTagFilters: Set<string>;
	showFavoritesOnly: boolean;
	activeBatchId: string | null;
}

export const discoveredVideos = writable<Video[]>([]);
export const downloadedVideos = writable<Video[]>([]);
export const allFolders = writable<string[]>([]);
export const folderMeta = writable<Record<string, FolderInfo>>({});
export const allTags = writable<Tag[]>([]);
export const tagAssignments = writable<Record<string, string[]>>({});
export const allCollections = writable<Collection[]>([]);
export const uploadBatches = writable<UploadBatch[]>([]);
export const duplicateGroups = writable<DuplicateGroup[]>([]);
export const isVideoDiscoveryLoading = writable(false);
export const scanProgress = writable<ScanProgress | null>(null);
// Mobile-only lazy loading: keep only the first few videos in memory until the user asks.
export const mobileLazyActive = writable(false);
export const mobileLazyHasMore = writable(false);
export const mobileLazyLoadingMore = writable(false);
export const pinGateVisible = writable(false);
export const pinGateError = writable('');
export const pinGateMode = writable(false);
export const pinGateUnlocked = writable(false);
export const showPinLockBtn = writable(false);

const savedSort = localStorage.getItem('vidvault-sort') || 'name';

export const filters = writable<FilterState>({
	search: '',
	sort: savedSort,
	activeFolder: '__all__',
	activeTagFilters: new Set(),
	showFavoritesOnly: false,
	activeBatchId: null,
});

filters.subscribe((f) => {
	localStorage.setItem('vidvault-sort', f.sort);
});

export const allVideos = derived(
	[discoveredVideos, downloadedVideos],
	([$discovered, $downloaded]) => {
		const mergedByPath = new Map<string, Video>();
		for (const video of $downloaded) {
			if (video?.path) mergedByPath.set(video.path, video);
		}
		for (const video of $discovered) {
			if (video?.path) mergedByPath.set(video.path, video);
		}
		return [...mergedByPath.values()];
	},
);

export const tagMap = derived(allTags, ($tags) =>
	Object.fromEntries($tags.map((t) => [t.id, t])),
);

export const collectionMap = derived(allCollections, ($cols) =>
	Object.fromEntries($cols.map((c) => [c.id, c])),
);

function sortVideos(videos: Video[], sort: string): Video[] {
	return [...videos].sort((a, b) => {
		if (sort === 'name') return a.name.localeCompare(b.name);
		if (sort === 'name-desc') return b.name.localeCompare(a.name);
		if (sort === 'folder')
			return (
				(a.folder || '').localeCompare(b.folder || '') ||
				a.name.localeCompare(b.name)
			);
		if (sort === 'ext')
			return a.ext.localeCompare(b.ext) || a.name.localeCompare(b.name);
		if (sort === 'modified')
			return (
				new Date(a.modified || 0).getTime() -
					new Date(b.modified || 0).getTime() ||
				a.name.localeCompare(b.name)
			);
		if (sort === 'modified-desc')
			return (
				new Date(b.modified || 0).getTime() -
					new Date(a.modified || 0).getTime() ||
				a.name.localeCompare(b.name)
			);
		if (sort === 'size')
			return (
				Number(a.size || 0) - Number(b.size || 0) ||
				a.name.localeCompare(b.name)
			);
		if (sort === 'size-desc')
			return (
				Number(b.size || 0) - Number(a.size || 0) ||
				a.name.localeCompare(b.name)
			);
		return 0;
	});
}

export function recomputeFiltered(
	videos: Video[],
	f: FilterState,
): Video[] {
	const q = f.search.toLowerCase();
	const filtered = videos.filter((v) => {
		const inF = f.activeFolder === '__all__' || (v.folder || '/') == f.activeFolder;
		const inFav = !f.showFavoritesOnly || Boolean(v.is_favorite);
		const inT =
			f.activeTagFilters.size === 0 ||
			[...f.activeTagFilters].every((tagID) => (v.tags || []).includes(tagID));
		const inS =
			!q ||
			v.name.toLowerCase().includes(q) ||
			(v.folder || '').toLowerCase().includes(q);
		const inBatch = !f.activeBatchId || v.batch_id === f.activeBatchId;
		return inF && inFav && inT && inS && inBatch;
	});
	return sortVideos(filtered, f.sort);
}

export const filteredVideos = derived(
	[allVideos, filters],
	([$videos, $filters]) => recomputeFiltered($videos, $filters),
);

export const incomingVideoCount = derived(allVideos, ($v) =>
	$v.filter((video) => Boolean(video.is_new)).length,
);

export const batchesNeedingSort = derived(uploadBatches, ($batches) =>
	$batches.filter((b) => b.needs_sorting && b.new_count > 0),
);

export const activeBatch = derived(
	[filters, uploadBatches],
	([$filters, $batches]) =>
		$filters.activeBatchId
			? $batches.find((b) => b.id === $filters.activeBatchId) ?? null
			: null,
);

export const tagVideoCounts = derived(allVideos, ($v) => {
	const counts: Record<string, number> = {};
	for (const video of $v) {
		for (const tagID of video.tags ?? []) {
			counts[tagID] = (counts[tagID] ?? 0) + 1;
		}
	}
	return counts;
});

export const folderVideoCounts = derived(allVideos, ($v) => {
	const counts: Record<string, number> = { __all__: $v.length };
	for (const video of $v) {
		const folder = video.folder || '/';
		counts[folder] = (counts[folder] ?? 0) + 1;
	}
	return counts;
});

export const collectionVideoCounts = derived(
	[allVideos, allCollections],
	([$v, $cols]) => {
		const byHash = new Set($v.map((video) => video.hash).filter(Boolean));
		return Object.fromEntries(
			$cols.map((c) => [
				c.id,
				(c.video_hashes ?? []).filter((h) => byHash.has(h)).length,
			]),
		);
	},
);

export function videoByPath(path: string): Video | undefined {
	return get(allVideos).find((v) => v.path === path);
}

export function adjacentInFiltered(path: string, dir: -1 | 1): Video | undefined {
	const list = get(filteredVideos);
	const idx = list.findIndex((v) => v.path === path);
	if (idx < 0) return undefined;
	const next = idx + dir;
	if (next < 0 || next >= list.length) return undefined;
	return list[next];
}

export function countIncomingVideos(): number {
	return get(incomingVideoCount);
}

export function tagCountByID(tagID: string): number {
	return get(tagVideoCounts)[tagID] ?? 0;
}

export function countCollectionResolvedVideos(collection: Collection): number {
	return get(collectionVideoCounts)[collection.id] ?? 0;
}

export function resolveCollectionVideos(collection: Collection): {
	ordered: Video[];
	missing: string[];
} {
	const byHash = new Map<string, Video>();
	for (const v of get(allVideos)) {
		if (!v.hash || byHash.has(v.hash)) continue;
		byHash.set(v.hash, v);
	}
	const ordered: Video[] = [];
	const missing: string[] = [];
	for (const hash of collection.video_hashes || []) {
		const v = byHash.get(hash);
		if (v) ordered.push(v);
		else missing.push(hash);
	}
	return { ordered, missing };
}

function parseFolders(data: FolderInfo[]): void {
	allFolders.set(data.map((f) => f.name));
	folderMeta.set(Object.fromEntries(data.map((f) => [f.name, f])));
}

function parseTags(data: { tags?: Tag[]; assignments?: Record<string, string[]> }): void {
	const tags = Array.isArray(data.tags) ? data.tags : [];
	allTags.set(tags);
	tagAssignments.set(
		data && typeof data.assignments === 'object' && data.assignments
			? data.assignments
			: {},
	);
}

function parseCollections(data: { collections?: Collection[] }): void {
	const raw = data && Array.isArray(data.collections) ? data.collections : [];
	const cols = raw.map((c) => ({
		id: c.id,
		name: c.name,
		video_hashes: Array.isArray(c.video_hashes) ? c.video_hashes : [],
		created_at: c.created_at || '',
		updated_at: c.updated_at || '',
	}));
	allCollections.set(cols);
}

function setDiscoveredVideos(videos: Video[]): void {
	const list = Array.isArray(videos) ? videos : [];
	const discoveredPaths = new Set(list.map((v) => v.path).filter(Boolean));
	discoveredVideos.set(list);
	downloadedVideos.update((dl) => dl.filter((v) => !discoveredPaths.has(v.path)));
}

export function upsertDiscoveredVideo(video: Video): void {
	if (!video?.path) return;
	downloadedVideos.update((dl) => dl.filter((v) => v.path !== video.path));
	discoveredVideos.update((list) => {
		const i = list.findIndex((v) => v.path === video.path);
		if (i >= 0) {
			const next = [...list];
			next[i] = video;
			return next;
		}
		return [...list, video];
	});
}

export function removeDiscoveredVideo(path: string): void {
	if (!path) return;
	discoveredVideos.update((list) => list.filter((v) => v.path !== path));
	downloadedVideos.update((list) => list.filter((v) => v.path !== path));
}

export function patchVideoPath(oldPath: string, video: Video): void {
	if (!oldPath || !video?.path) return;
	removeDiscoveredVideo(oldPath);
	upsertDiscoveredVideo(video);
}

export function upsertDownloadedVideo(video: Video): void {
	if (!video?.path) return;
	downloadedVideos.update((dl) => {
		const idx = dl.findIndex((v) => v.path === video.path);
		if (idx >= 0) {
			const next = [...dl];
			next[idx] = video;
			return next;
		}
		return [...dl, video];
	});
}

export function computeDuplicateGroups(): void {
	const byHash = new Map<string, Video[]>();
	for (const v of get(allVideos)) {
		if (!v.hash) continue;
		const members = byHash.get(v.hash) || [];
		members.push(v);
		byHash.set(v.hash, members);
	}
	const groups: DuplicateGroup[] = [...byHash.entries()]
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
				keepPath: sorted[0]?.path || '',
			};
		})
		.sort(
			(a, b) =>
				b.members.length - a.members.length || a.hash.localeCompare(b.hash),
		);
	duplicateGroups.set(groups);
}

export function updateDuplicateBannerState(mayAutoOpen: boolean): void {
	const groups = get(duplicateGroups);
	if (!groups.length) {
		duplicateBannerDismissed.set(false);
		autoOpenedDuplicates.set(false);
		return;
	}
	if (
		mayAutoOpen &&
		!get(autoOpenedDuplicates) &&
		groups.length > 0
	) {
		dupModalOpen.set(true);
		autoOpenedDuplicates.set(true);
	}
}

export function updateIncomingBannerState(): void {
	const needingSort = get(batchesNeedingSort);
	if (!needingSort.length) {
		incomingBannerDismissed.set(false);
		if (get(incomingModalOpen)) {
			incomingModalOpen.set(false);
			incomingModalBatchId.set(null);
		}
	}
}

export function refreshDerivedState(mayAutoOpenDuplicates = false): void {
	computeDuplicateGroups();
	updateDuplicateBannerState(mayAutoOpenDuplicates);
	updateIncomingBannerState();
}

export async function refreshTagAssignmentsForHashes(hashes: string[]): Promise<void> {
	const uniqueHashes = [...new Set((hashes || []).filter(Boolean))];
	if (!uniqueHashes.length) return;
	const payload = await fetchTagAssignmentsForHashes(uniqueHashes);
	const updates = Array.isArray(payload?.videos) ? payload.videos : [];
	const tagsByHash = Object.fromEntries(
		updates.map((u) => [u.hash, Array.isArray(u.tags) ? u.tags : []]),
	);
	const updateList = (videos: Video[]) => {
		for (const video of videos) {
			if (!video?.hash) continue;
			if (Object.prototype.hasOwnProperty.call(tagsByHash, video.hash)) {
				video.tags = tagsByHash[video.hash];
			}
		}
	};
	discoveredVideos.update((v) => {
		updateList(v);
		return [...v];
	});
	downloadedVideos.update((v) => {
		updateList(v);
		return [...v];
	});
}

export async function refreshTagsOnly(): Promise<void> {
	const data = await fetchTags();
	parseTags(data);
}

export async function refreshBatches(): Promise<void> {
	try {
		const batches = await api.fetchBatches();
		uploadBatches.set(Array.isArray(batches) ? batches : []);
	} catch {
		uploadBatches.set([]);
	}
}

export async function refreshCollectionsOnly(): Promise<void> {
	const data = await fetchCollections();
	parseCollections(data);
}

export async function refreshFoldersOnly(): Promise<void> {
	const data = await fetchFolders();
	parseFolders(data);
}

async function refreshVideosFallback(mayAutoOpenDuplicates: boolean): Promise<void> {
	const videos = await api.fetchVideos();
	setDiscoveredVideos(videos);
	isVideoDiscoveryLoading.set(false);
	refreshDerivedState(mayAutoOpenDuplicates);
}

export async function refreshVideosProgressive(
	mayAutoOpenDuplicates: boolean,
	opts: { rescan?: boolean; clearFirst?: boolean } = {},
): Promise<void> {
	isVideoDiscoveryLoading.set(true);
	scanProgress.set(null);
	if (opts.clearFirst) {
		setDiscoveredVideos([]);
	}
	try {
		const pendingUpserts: Video[] = [];
		let flushScheduled = false;
		const FLUSH_INTERVAL = 200;

		function flushPending() {
			flushScheduled = false;
			if (!pendingUpserts.length) return;
			const batch = pendingUpserts.splice(0);
			discoveredVideos.update((list) => {
				const byPath = new Map(list.map((v) => [v.path, v]));
				for (const v of batch) {
					if (v?.path) byPath.set(v.path, v);
				}
				return [...byPath.values()];
			});
			downloadedVideos.update((dl) => {
				const batchPaths = new Set(batch.map((v) => v.path).filter(Boolean));
				return dl.filter((v) => !batchPaths.has(v.path));
			});
			updateIncomingBannerState();
		}

		function scheduleFlush() {
			if (!flushScheduled) {
				flushScheduled = true;
				setTimeout(flushPending, FLUSH_INTERVAL);
			}
		}

		await streamVideos(
			(video) => {
				pendingUpserts.push(video);
				scheduleFlush();
			},
			{
				rescan: opts.rescan,
				onProgress: (p) => scanProgress.set(p),
				onRemoved: (path) => removeDiscoveredVideo(path),
			},
		);
		flushPending();
		// Note: the `streamVideos()` callback already queued everything for batched flushing.
		isVideoDiscoveryLoading.set(false);
		scanProgress.set(null);
		refreshDerivedState(mayAutoOpenDuplicates);
	} catch {
		await refreshVideosFallback(mayAutoOpenDuplicates);
		scanProgress.set(null);
	}
}

function isMobileClient(): boolean {
	if (typeof navigator === 'undefined') return false;
	const w = window as any;
	const mm = typeof w?.matchMedia === 'function' ? w.matchMedia('(max-width: 600px)') : null;
	if (mm && typeof mm.matches === 'boolean') return mm.matches;
	return /Android|iPhone|iPad|iPod/i.test(navigator.userAgent || '');
}

const MOBILE_TOP_N = 5;
const MOBILE_PAGE_SIZE = 25;
let mobileLazyOffset = 0;

async function initMobileLazy(rescan: boolean): Promise<void> {
	mobileLazyActive.set(true);
	mobileLazyLoadingMore.set(false);
	mobileLazyOffset = 0;

	// Keep memory low: only load a small subset of videos.
	discoveredVideos.set([]);
	downloadedVideos.set([]);
	isVideoDiscoveryLoading.set(true);
	scanProgress.set(null);

	const firstRaw = await fetchVideosPage(0, MOBILE_TOP_N + 1, { rescan });
	const first = firstRaw.slice(0, MOBILE_TOP_N);
	setDiscoveredVideos(first);
	mobileLazyOffset = first.length;
	mobileLazyHasMore.set(firstRaw.length > MOBILE_TOP_N);

	isVideoDiscoveryLoading.set(false);
	scanProgress.set(null);
	// Avoid auto-opening modals with partial data.
	refreshDerivedState(false);
}

export async function loadMoreMobileVideos(): Promise<void> {
	if (!get(mobileLazyActive)) return;
	if (get(mobileLazyLoadingMore)) return;
	if (!get(mobileLazyHasMore)) return;

	mobileLazyLoadingMore.set(true);
	try {
		const nextRaw = await fetchVideosPage(
			mobileLazyOffset,
			MOBILE_PAGE_SIZE + 1,
			{ rescan: false },
		);
		const next = nextRaw.slice(0, MOBILE_PAGE_SIZE);
		const paths = new Set(next.map((v) => v?.path).filter(Boolean));

		discoveredVideos.update((list) => {
			const byPath = new Map<string, Video>();
			for (const v of list) if (v?.path) byPath.set(v.path, v);
			for (const v of next) if (v?.path) byPath.set(v.path, v);
			return [...byPath.values()];
		});

		downloadedVideos.update((dl) => {
			if (!paths.size) return dl;
			return dl.filter((v) => v?.path && !paths.has(v.path));
		});

		mobileLazyOffset += next.length;
		mobileLazyHasMore.set(nextRaw.length > MOBILE_PAGE_SIZE && next.length > 0);
		refreshDerivedState(false);
	} catch {
		// On failure, stop further attempts to avoid infinite retries.
		mobileLazyHasMore.set(false);
	} finally {
		mobileLazyLoadingMore.set(false);
	}
}

export async function init(): Promise<void> {
	const [fr, tr, cr] = await Promise.all([
		fetchFolders(),
		fetchTags(),
		fetchCollections(),
	]);
	parseFolders(fr);
	parseTags(tr);
	parseCollections(cr);
	if (isMobileClient()) {
		await Promise.all([refreshBatches(), initMobileLazy(false)]);
		return;
	}
	await Promise.all([refreshBatches(), refreshVideosProgressive(true, { clearFirst: false })]);
}

export async function refresh(opts: { rescan?: boolean } = {}): Promise<void> {
	if (isMobileClient()) {
		const [fr, tr, cr] = await Promise.all([
			fetchFolders(),
			fetchTags(),
			fetchCollections(),
		]);
		parseFolders(fr);
		parseTags(tr);
		parseCollections(cr);
		await Promise.all([refreshBatches(), initMobileLazy(opts.rescan ?? true)]);
		return;
	}
	const [fr, tr, cr] = await Promise.all([
		fetchFolders(),
		fetchTags(),
		fetchCollections(),
	]);
	parseFolders(fr);
	parseTags(tr);
	parseCollections(cr);
	await Promise.all([
		refreshBatches(),
		refreshVideosProgressive(false, { rescan: opts.rescan ?? true, clearFirst: true }),
	]);
}

export function moveVideoLocally(srcPath: string, destFolder: string): boolean {
	const videos = get(allVideos);
	const vi = videos.findIndex((v) => v?.path === srcPath);
	if (vi < 0) return false;
	const v = videos[vi];
	const newPath = movedVideoRelPath(srcPath, destFolder);
	const newFolder = folderFromRelPath(newPath);
	const f = get(filters);
	const viewingAll = f.activeFolder === '__all__';
	const nk = newFolder || '/';
	const leavesCurrentFolder = !viewingAll && nk != f.activeFolder;

	const playing = get(playingPath);
	if (playing === srcPath && leavesCurrentFolder) {
		closePlayer();
	}

	const sel = get(selectedPaths);
	if (sel.has(srcPath)) {
		selectedPaths.update((s) => {
			const next = new Set(s);
			next.delete(srcPath);
			if (!leavesCurrentFolder) next.add(newPath);
			return next;
		});
	}

	const removeFromStores = (path: string) => {
		discoveredVideos.update((list) => list.filter((x) => x.path !== path));
		downloadedVideos.update((list) => list.filter((x) => x.path !== path));
	};

	if (leavesCurrentFolder) {
		removeFromStores(srcPath);
		refreshDerivedState(false);
		return true;
	}

	const oldPath = v.path;
	v.path = newPath;
	v.folder = newFolder;

	// Force store update
	discoveredVideos.update((list) => [...list]);
	downloadedVideos.update((list) => [...list]);

	const filtered = recomputeFiltered(get(allVideos), get(filters));
	if (!filtered.some((x) => x.path === newPath)) {
		removeFromStores(oldPath);
		refreshDerivedState(false);
		return true;
	}

	refreshDerivedState(false);
	return true;
}

export async function moveVideo(srcPath: string, destFolder: string): Promise<void> {
	const res = await api.moveVideoApi(srcPath, destFolder);
	if (res.ok) {
		if (res.video) {
			patchVideoPath(res.old_path || srcPath, res.video);
			try {
				await refreshFoldersOnly();
			} catch {
				/* best-effort */
			}
			computeDuplicateGroups();
			updateDuplicateBannerState(false);
			updateIncomingBannerState();
		} else {
			const localOk = moveVideoLocally(srcPath, destFolder);
			if (!localOk) await refresh({ rescan: true });
			else {
				try {
					await refreshFoldersOnly();
				} catch {
					/* best-effort */
				}
				computeDuplicateGroups();
				updateDuplicateBannerState(false);
				updateIncomingBannerState();
			}
		}
		toast('Moved to ' + destFolder, 'success');
	} else {
		toast('Move failed: ' + res.message, 'error');
	}
}

export async function refreshAfterTagDataChange(
	prevFilteredPaths: string[],
	affectedHashes: string[],
): Promise<void> {
	refreshDerivedState(false);
	const nextPaths = get(filteredVideos).map((v) => v.path);
	const unchanged =
		prevFilteredPaths.length === nextPaths.length &&
		prevFilteredPaths.every((p, i) => p === nextPaths[i]);
	if (!unchanged) {
		// reactive {#each} handles re-render
	}
	updateDuplicateBannerState(false);
	updateIncomingBannerState();
	void affectedHashes; // store updates handle tag badges
}

export function unloadLibraryAfterPinLock(): void {
	closePlayer();
	discoveredVideos.set([]);
	downloadedVideos.set([]);
	isVideoDiscoveryLoading.set(false);
	mobileLazyActive.set(false);
	mobileLazyHasMore.set(false);
	mobileLazyLoadingMore.set(false);
	scanProgress.set(null);
	exitSelectMode();
	refreshDerivedState(false);
}

export function setPinOverlayVisible(show: boolean, message = ''): void {
	pinGateVisible.set(show);
	pinGateError.set(message);
	if (typeof document !== 'undefined') {
		document.body.classList.toggle('pin-gate-open', show);
	}
}

export async function bootstrap(): Promise<void> {
	let st;
	try {
		st = await fetchAuthStatus();
	} catch (e) {
		console.error(e);
		setPinGateMode(true);
		pinGateMode.set(true);
		setPinGateUnlocked(false);
		pinGateUnlocked.set(false);
		showPinLockBtn.set(false);
		setPinOverlayVisible(true, 'Could not reach server.');
		return;
	}
	if (!st.pinRequired) {
		setPinGateMode(false);
		pinGateMode.set(false);
		setPinGateUnlocked(false);
		pinGateUnlocked.set(false);
		showPinLockBtn.set(false);
		setPinOverlayVisible(false);
		await init();
		return;
	}
	installFetch401Handler();
	if (st.authenticated) {
		setPinGateMode(true);
		pinGateMode.set(true);
		setPinGateUnlocked(true);
		pinGateUnlocked.set(true);
		setPinOverlayVisible(false);
		showPinLockBtn.set(true);
		await init();
		return;
	}
	setPinGateUnlocked(false);
	pinGateUnlocked.set(false);
	showPinLockBtn.set(false);
	setPinOverlayVisible(true);
	setPinGateMode(true);
	pinGateMode.set(true);
}

export function getUploadFolderOptions(): string[] {
	return [
		...new Set([
			'/',
			...get(allFolders),
			...get(allVideos).map((v) => v.folder || '/'),
		]),
	].sort();
}

export function getMoveFolderOptions(): string[] {
	return [
		...new Set([
			...get(allFolders),
			...get(allVideos).map((v) => v.folder || '/'),
		]),
	].sort();
}

export function getSidebarFolders(): string[] {
	const merged = new Set([
		...get(allVideos).map((v) => v.folder || '/'),
		...get(allFolders),
	]);
	return ['__all__', ...[...merged].sort()];
}
