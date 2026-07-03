<script lang="ts">
import { onMount } from 'svelte';
import {
	ctxMenuOpen,
	ctxMenuPos,
	ctxVideo,
	hideCtxMenu,
	moveModalOpen,
	moveTarget,
	tagModalOpen,
	tagTargetPaths,
} from '../lib/stores/ui';
import { deleteVideoPath, forgetVideos, renameVideoPath } from '../lib/api';
import {
	patchVideoPath,
	removeDiscoveredVideo,
	refreshDerivedState,
	upsertDiscoveredVideo,
} from '../lib/stores/library';
import { toast } from '../lib/toast';

function handleDocumentClick() {
	hideCtxMenu();
}

onMount(() => {
	document.addEventListener('click', handleDocumentClick);
	return () => document.removeEventListener('click', handleDocumentClick);
});

function openTagModal() {
	const v = $ctxVideo;
	if (!v) return;
	tagTargetPaths.set([v.path]);
	tagModalOpen.set(true);
	hideCtxMenu();
}

function openMoveModal() {
	const v = $ctxVideo;
	if (!v) return;
	moveTarget.set(v);
	moveModalOpen.set(true);
	hideCtxMenu();
}

async function handleRename() {
	const v = $ctxVideo;
	if (!v) return;
	const next = prompt('Rename file (new base name)', v.name);
	if (next == null) return;
	const trimmed = next.trim();
	if (!trimmed) return;
	const res = await renameVideoPath(v.path, trimmed);
	if (res.ok) {
		if (res.video) {
			patchVideoPath(res.old_path || v.path, res.video);
			refreshDerivedState(false);
		}
		toast('Renamed file', 'success');
	} else {
		toast('Rename failed: ' + res.message, 'error');
	}
	hideCtxMenu();
}

async function handleForget() {
	const v = $ctxVideo;
	if (!v) return;
	const ok = await forgetVideos([v.path], [], false);
	if (!ok) {
		toast('Failed to forget file', 'error');
		return;
	}
	toast('File forgotten and re-queued', 'success');
	upsertDiscoveredVideo({ ...v, is_new: true });
	refreshDerivedState(false);
	hideCtxMenu();
}

async function handleDelete() {
	const v = $ctxVideo;
	if (!v) return;
	const ok = confirm(
		`Delete "${v.name}"?\n\nThis removes the file from disk and cannot be undone.`,
	);
	if (!ok) return;
	const res = await deleteVideoPath(v.path);
	if (res.ok) {
		removeDiscoveredVideo(v.path);
		refreshDerivedState(false);
		toast('Deleted ' + v.name, 'success');
	} else {
		toast('Delete failed: ' + res.message, 'error');
	}
	hideCtxMenu();
}
</script>

<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
<div
	id="ctx-menu"
	class:open={$ctxMenuOpen}
	style:left="{$ctxMenuPos.x}px"
	style:top="{$ctxMenuPos.y}px"
>
	<div class="ctx-item" id="ctx-tags" on:click={openTagModal}>
		<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
			<path d="M20.59 13.41 11 3H4v7l9.59 9.59a2 2 0 0 0 2.82 0l4.18-4.18a2 2 0 0 0 0-2.82z" /><path d="M7 7h.01" />
		</svg>
		Edit tags…
	</div>
	<div class="ctx-item" id="ctx-move" on:click={openMoveModal}>
		<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
			<path d="M5 9l4-4 4 4" /><path d="M9 5v14" /><path d="M19 15l-4 4-4-4" /><path d="M15 19V5" />
		</svg>
		Move to folder…
	</div>
	<div class="ctx-item" id="ctx-rename" on:click={handleRename}>
		<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
			<path d="M12 20h9" /><path d="M16.5 3.5a2.12 2.12 0 0 1 3 3L7 19l-4 1 1-4Z" />
		</svg>
		Rename file…
	</div>
	<div class="ctx-item danger" id="ctx-delete" on:click={handleDelete}>
		<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
			<path d="M3 6h18" /><path d="M8 6V4h8v2" /><path d="M19 6l-1 14H6L5 6" /><path d="M10 11v6" /><path d="M14 11v6" />
		</svg>
		Delete file…
	</div>
	<div class="ctx-item" id="ctx-forget" on:click={handleForget}>
		<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
			<path d="M3 12h18" /><path d="M12 3v18" />
		</svg>
		Forget (mark as new)…
	</div>
</div>
