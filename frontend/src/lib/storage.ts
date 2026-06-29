export function loadNumericStorage(
	key: string,
	fallback: number,
	min: number,
	max: number,
): number {
	try {
		const raw = localStorage.getItem(key);
		if (raw === null) return fallback;
		const value = Number(raw);
		if (Number.isNaN(value)) return fallback;
		return Math.max(min, Math.min(max, value));
	} catch {
		return fallback;
	}
}

export function loadBooleanStorage(key: string, fallback: boolean): boolean {
	try {
		const raw = localStorage.getItem(key);
		if (raw === null) return fallback;
		return raw === 'true';
	} catch {
		return fallback;
	}
}

export function persistStorageValue(key: string, value: string | number | boolean): void {
	try {
		localStorage.setItem(key, String(value));
	} catch {
		/* ignore */
	}
}

export const WATCH_GRID_COLS_STORAGE_KEY = 'vidvault.watchGridCols';
export const WATCH_GRID_VOLUME_STORAGE_KEY = 'vidvault.watchCollectionVolume';
export const WATCH_GRID_MUTED_STORAGE_KEY = 'vidvault.watchCollectionMuted';
export const HOVER_PREVIEW_AUDIO_ENABLED_KEY = 'vidvault.hoverPreviewAudioEnabled';
export const HOVER_PREVIEW_VOLUME_KEY = 'vidvault.hoverPreviewVolume';
