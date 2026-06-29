import { derived, get, writable } from 'svelte/store';
import * as api from '../api';
import {
	fetchAuthStatus,
	fetchCollections,
	fetchFolders,
	fetchTags,
	fetchTagAssignmentsForHashes,
	streamVideos,
} from '../api';
import { toast } from '../toast';
import { folderFromRelPath, movedVideoRelPath } from '../video-helpers';
import type {
	Collection,
	DuplicateGroup,
	FolderInfo,
	Tag,
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
	activeIncomingFilter: boolean;
}

export const discoveredVideos = writable<Video[]>([]);
export const downloadedVideos = writable<Video[]>([]);
export const allFolders = writable<string[]>([]);
export const folderMeta = writable<Record<string, FolderInfo>>({});
export const allTags = writable<Tag[]>([]);
export const tagAssignments = writable<Record<string, string[]>>({});
export const allCollections = writable<Collection[]>([]);
export const duplicateGroups = writable<DuplicateGroup[]>([]);
export const isVideoDiscoveryLoading = writable(false);
export const pinGateVisible = writable(false);
export const pinGateError = writable('');
export const pinGateMode = writable(false);
export const pinGateUnlocked = writable(false);
export const showPinLockBtn = writable(false);

export const filters = writable<FilterState>({
	search: '',
	sort: 'name',
	activeFolder: '__all__',
	activeTagFilters: new Set(),
	showFavoritesOnly: false,
	activeIncomingFilter: false,
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
		const inIncoming = !f.activeIncomingFilter || Boolean(v.is_new);
		return inF && inFav && inT && inS && inIncoming;
	});
	return sortVideos(filtered, f.sort);
}

export const filteredVideos = derived(
	[allVideos, filters],
	([$videos, $filters]) => recomputeFiltered($videos, $filters),
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
	return get(allVideos).filter((v) => Boolean(v.is_new)).length;
}

export function tagCountByID(tagID: string): number {
	return get(allVideos).filter(
		(v) => Array.isArray(v.tags) && v.tags.includes(tagID),
	).length;
}

export function countCollectionResolvedVideos(collection: Collection): number {
	const hashes = new Set(collection.video_hashes || []);
	return get(allVideos).filter((v) => hashes.has(v.hash)).length;
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
	const incomingCount = countIncomingVideos();
	if (!incomingCount) {
		incomingBannerDismissed.set(false);
		if (get(incomingModalOpen)) incomingModalOpen.set(false);
	}
}

export function refreshDerivedState(mayAutoOpenDuplicates = false): void {
	computeDuplicateGroups();
	updateDuplicateBannerState(mayAutoOpenDuplicates);
	updateIncomingBannerState();
}

async function refreshTagAssignmentsForHashes(hashes: string[]): Promise<void> {
	const uniqueHashes = [...new Set((hashes || []).filter(Boolean))];
	if (!uniqueHashes.length) return;
	const payload = await fetchTagAssignmentsForHashes(uniqueHashes);
	const updates = Array.isArray(payload?.videos) ? payload.videos : [];
	const tagsByHash = Object.fromEntries(
		updates.map((u) => [u.hash, Array.isArray(u.tags) ? u.tags : []]),
	);
	allVideos; // trigger - we need to update videos in place
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
): Promise<void> {
	isVideoDiscoveryLoading.set(true);
	try {
		let lastRenderCount = 0;
		setDiscoveredVideos([]);
		refreshDerivedState(mayAutoOpenDuplicates);
		const streamed = await streamVideos((video) => {
			// progressive - handled after stream completes in batches
		});
		// Re-stream with batching like original
		const res = await fetch('/api/videos/stream');
		if (!res.ok || !res.body) {
			await refreshVideosFallback(mayAutoOpenDuplicates);
			return;
		}
		const reader = res.body.getReader();
		const decoder = new TextDecoder();
		let buffer = '';
		const batch: Video[] = [];
		setDiscoveredVideos([]);
		refreshDerivedState(false);
		while (true) {
			const { value, done } = await reader.read();
			if (done) break;
			buffer += decoder.decode(value, { stream: true });
			const lines = buffer.split('\n');
			buffer = lines.pop() || '';
			for (const line of lines) {
				const trimmed = line.trim();
				if (!trimmed) continue;
				try {
					const event = JSON.parse(trimmed);
					if (event.type === 'video' && event.payload) {
						batch.push(event.payload);
						if (batch.length - lastRenderCount >= 20) {
							setDiscoveredVideos([...batch]);
							refreshDerivedState(false);
							lastRenderCount = batch.length;
						}
					}
				} catch {
					/* skip */
				}
			}
		}
		setDiscoveredVideos(batch.length ? batch : streamed);
		isVideoDiscoveryLoading.set(false);
		refreshDerivedState(mayAutoOpenDuplicates);
	} catch {
		await refreshVideosFallback(mayAutoOpenDuplicates);
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
	setDiscoveredVideos([]);
	refreshDerivedState(true);
	await refreshVideosProgressive(true);
}

export async function refresh(): Promise<void> {
	const [fr, tr, cr] = await Promise.all([
		fetchFolders(),
		fetchTags(),
		fetchCollections(),
	]);
	parseFolders(fr);
	parseTags(tr);
	parseCollections(cr);
	await refreshVideosProgressive(false);
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
		const localOk = moveVideoLocally(srcPath, destFolder);
		if (!localOk) await refresh();
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
		toast('Moved to ' + destFolder, 'success');
	} else {
		toast('Move failed: ' + (await res.text()), 'error');
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
		return;
	}
	if (!st.pinRequired) {
		setPinGateMode(false);
		pinGateMode.set(false);
		setPinGateUnlocked(false);
		pinGateUnlocked.set(false);
		showPinLockBtn.set(false);
		await init();
		return;
	}
	setPinGateMode(true);
	pinGateMode.set(true);
	installFetch401Handler();
	if (st.authenticated) {
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
