<script lang="ts">
import { get } from 'svelte/store';
import {
	allFolders,
	allVideos,
	filters,
	folderMeta,
	folderVideoCounts,
	moveVideo,
	refresh,
	refreshFoldersOnly,
} from '../../lib/stores/library';
import { activeCollectionID, watchCollectionOpen } from '../../lib/stores/ui';
import { createCollectionFromFolder, mkdirApi, rmdirApi } from '../../lib/api';
import { toast } from '../../lib/toast';

export let collapsed = false;

let showForm = false;
let folderInput = '';
let dragOverFolder: string | null = null;

$: folders = (() => {
	const merged = new Set([
		...$allVideos.map((v) => v.folder || '/'),
		...$allFolders,
	]);
	return ['__all__', ...[...merged].sort()];
})();

function selectFolder(f: string) {
	filters.update((prev) => ({
		...prev,
		activeFolder: f,
		activeBatchId: f !== '__all__' && prev.activeBatchId ? null : prev.activeBatchId,
	}));
}

function showCreateForm() {
	collapsed = false;
	showForm = true;
	folderInput = '';
}

function hideForm() {
	showForm = false;
	folderInput = '';
}

async function createFolder() {
	const name = folderInput.trim();
	if (!name) return;
	const res = await mkdirApi(name);
	if (res.ok) {
		hideForm();
		toast('Created ' + name, 'success');
		try {
			await refreshFoldersOnly();
		} catch {
			await refresh();
		}
	} else {
		toast('Failed: ' + (await res.text()), 'error');
	}
}

async function deleteFolder(f: string, e: MouseEvent) {
	e.stopPropagation();
	const meta = get(folderMeta)[f];
	const warnMsg = meta?.has_other_files
		? '\n\n⚠ This folder contains non-video files. They will also be moved to root.'
		: '';
	if (!confirm(`Delete folder "${f}"? All files inside will be moved to root.${warnMsg}`)) return;
	const res = await rmdirApi(f);
	if (res.ok) {
		if (get(filters).activeFolder === f) {
			filters.update((prev) => ({ ...prev, activeFolder: '__all__' }));
		}
		toast('Deleted ' + f, 'success');
		await refresh();
	} else {
		toast('Delete failed: ' + (await res.text()), 'error');
	}
}

async function makeCollectionFromFolder(f: string, e: MouseEvent) {
	e.stopPropagation();
	const suggestedName = f;
	const inputName = prompt('Create collection from folder', suggestedName);
	if (inputName === null) return;
	const name = inputName.trim() || suggestedName;
	const result = await createCollectionFromFolder(f, name);
	if (!result.ok) {
		toast('Create collection failed: ' + (result.error || 'unknown error'), 'error');
		return;
	}
	await refresh();
	if (result.collection) {
		activeCollectionID.set(result.collection.id);
		watchCollectionOpen.set(true);
	}
	toast('Collection created from folder', 'success');
}

function handleDragOver(e: DragEvent, f: string) {
	e.preventDefault();
	dragOverFolder = f;
}

function handleDragLeave(f: string) {
	if (dragOverFolder === f) dragOverFolder = null;
}

async function handleDrop(e: DragEvent, f: string) {
	e.preventDefault();
	dragOverFolder = null;
	const src = e.dataTransfer?.getData('text/plain');
	if (src) await moveVideo(src, f);
}
</script>

<div class="sidebar-section" class:collapsed>
	<div class="sidebar-header">
		<span class="sidebar-label">Folders</span>
		<div class="sidebar-actions">
			<button class="new-folder-btn" id="new-folder-btn" title="New folder" on:click={showCreateForm}>+</button>
			<button class="collapse-btn" id="folders-collapse-btn" title="Collapse folders" on:click={() => (collapsed = !collapsed)}>
				▾
			</button>
		</div>
	</div>
	<div class="sidebar-folder-form" class:visible={showForm} id="sidebar-folder-form">
		<input
			id="sidebar-folder-input"
			placeholder="folder name…"
			autocomplete="off"
			bind:value={folderInput}
			on:keydown={(e) => {
				if (e.key === 'Enter') createFolder();
				if (e.key === 'Escape') hideForm();
			}}
		/>
		<button id="sidebar-folder-confirm" title="Create" on:click={createFolder}>✓</button>
		<button id="sidebar-folder-cancel" title="Cancel" on:click={hideForm}>✕</button>
	</div>
	<div id="folder-list">
		{#each folders as f (f)}
			<div
				class="folder-btn"
				class:active={$filters.activeFolder === f}
				class:drag-over={dragOverFolder === f}
				data-folder={f}
				role="button"
				tabindex="0"
				title={f === '__all__' ? 'All files' : f}
				on:click={() => selectFolder(f)}
				on:keydown={(e) => e.key === 'Enter' && selectFolder(f)}
				on:dragover={(e) => f !== '__all__' && handleDragOver(e, f)}
				on:dragleave={() => f !== '__all__' && handleDragLeave(f)}
				on:drop={(e) => f !== '__all__' && handleDrop(e, f)}
			>
				<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z" />
				</svg>
				<span class="folder-name">{f === '__all__' ? 'All files' : f}</span>
				{#if f !== '__all__' && $folderMeta[f]?.has_other_files}
					<span class="folder-warn" title="Contains non-video files">⚠</span>
				{/if}
				<span class="count">{$folderVideoCounts[f] ?? 0}</span>
				{#if f !== '__all__' && f !== '/'}
					<button class="watch-folder-collection-btn" title="Make collection from this folder" on:click={(e) => makeCollectionFromFolder(f, e)}>◫</button>
					<button class="del-folder-btn" title="Delete folder (moves files to root)" on:click={(e) => deleteFolder(f, e)}>✕</button>
				{/if}
			</div>
		{/each}
	</div>
</div>
