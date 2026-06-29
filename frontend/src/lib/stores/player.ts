import { writable } from 'svelte/store';

export const playingPath = writable<string | null>(null);

export function openPlayer(path: string): void {
	playingPath.set(path);
}

export function closePlayer(): void {
	playingPath.set(null);
}
