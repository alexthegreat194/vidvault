import { writable } from 'svelte/store';
import {
	loadBooleanStorage,
	loadNumericStorage,
	persistStorageValue,
	WATCH_GRID_COLS_STORAGE_KEY,
	WATCH_GRID_VOLUME_STORAGE_KEY,
	WATCH_GRID_MUTED_STORAGE_KEY,
	HOVER_PREVIEW_AUDIO_ENABLED_KEY,
	HOVER_PREVIEW_VOLUME_KEY,
} from '../storage';
import type { Video } from '../types';

export const duplicateBannerDismissed = writable(false);
export const autoOpenedDuplicates = writable(false);
export const incomingBannerDismissed = writable(false);

export const dupModalOpen = writable(false);
export const incomingModalOpen = writable(false);
export const tagModalOpen = writable(false);
export const moveModalOpen = writable(false);
export const uploadModalOpen = writable(false);
export const collectionPickerOpen = writable(false);
export const watchCollectionOpen = writable(false);
export const previewSettingsOpen = writable(false);

export const activeCollectionID = writable('');
export const galleryListView = writable(false);

export const hoverPreviewAudioEnabled = writable(
	loadBooleanStorage(HOVER_PREVIEW_AUDIO_ENABLED_KEY, false),
);
export const hoverPreviewVolume = writable(
	loadNumericStorage(HOVER_PREVIEW_VOLUME_KEY, 0.5, 0, 1),
);

export const watchGridCols = writable(
	loadNumericStorage(WATCH_GRID_COLS_STORAGE_KEY, 4, 1, 4),
);
export const watchCollectionMasterVolume = writable(
	loadNumericStorage(WATCH_GRID_VOLUME_STORAGE_KEY, 1, 0, 1),
);
export const watchCollectionMasterMuted = writable(
	loadBooleanStorage(WATCH_GRID_MUTED_STORAGE_KEY, false),
);

export const ctxMenuOpen = writable(false);
export const ctxMenuPos = writable({ x: 0, y: 0 });
export const ctxVideo = writable<Video | null>(null);

export const moveTarget = writable<Video | null>(null);
export const moveSelectedFolder = writable<string | null>(null);

export const tagTargetPaths = writable<string[]>([]);
export const pendingTagActions = writable<Map<string, 'add' | 'remove'>>(new Map());

export const collectionPickerTargetHashes = writable<string[]>([]);
export const collectionPickerSelectedID = writable('');

export const incomingSelectedFolder = writable('/');
export const incomingSelectedTagIDs = writable<Set<string>>(new Set());

hoverPreviewAudioEnabled.subscribe((v) => {
	persistStorageValue(HOVER_PREVIEW_AUDIO_ENABLED_KEY, v);
});
hoverPreviewVolume.subscribe((v) => {
	persistStorageValue(HOVER_PREVIEW_VOLUME_KEY, v);
});
watchGridCols.subscribe((v) => {
	persistStorageValue(WATCH_GRID_COLS_STORAGE_KEY, v);
});
watchCollectionMasterVolume.subscribe((v) => {
	persistStorageValue(WATCH_GRID_VOLUME_STORAGE_KEY, v);
});
watchCollectionMasterMuted.subscribe((v) => {
	persistStorageValue(WATCH_GRID_MUTED_STORAGE_KEY, v);
});

export function showCtxMenu(e: MouseEvent, video: Video): void {
	ctxVideo.set(video);
	ctxMenuPos.set({
		x: Math.min(e.clientX, window.innerWidth - 180),
		y: Math.min(e.clientY, window.innerHeight - 60),
	});
	ctxMenuOpen.set(true);
}

export function hideCtxMenu(): void {
	ctxMenuOpen.set(false);
}

export function setHoverPreviewAudioEnabled(next: boolean): void {
	hoverPreviewAudioEnabled.set(!!next);
}

export function toggleHoverPreviewAudio(): void {
	hoverPreviewAudioEnabled.update((v) => !v);
}
