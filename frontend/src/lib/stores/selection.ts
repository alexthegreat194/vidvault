import { writable } from 'svelte/store';

export const selectMode = writable(false);
export const selectedPaths = writable<Set<string>>(new Set());

export function toggleSelectPath(path: string): void {
	selectedPaths.update((set) => {
		const next = new Set(set);
		if (next.has(path)) next.delete(path);
		else next.add(path);
		return next;
	});
}

export function clearSelection(): void {
	selectedPaths.set(new Set());
}

export function exitSelectMode(): void {
	selectMode.set(false);
	selectedPaths.set(new Set());
	if (typeof document !== 'undefined') {
		document.body.classList.remove('select-mode');
	}
}

export function enterSelectMode(): void {
	selectMode.set(true);
	if (typeof document !== 'undefined') {
		document.body.classList.add('select-mode');
	}
}

export function selectAllPaths(paths: string[]): void {
	selectedPaths.set(new Set(paths));
}
