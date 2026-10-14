<script lang="ts">
import { get } from 'svelte/store';
import Modal from './Modal.svelte';
import {
	allTags,
	allVideos,
	filters,
	getMoveFolderOptions,
	refresh,
	refreshBatches,
} from '../../lib/stores/library';
import {
	incomingModalBatchId,
	incomingModalOpen,
	incomingSelectedFolder,
	incomingSelectedTagIDs,
} from '../../lib/stores/ui';
import {
	clearNewStatus,
	moveVideoApi,
	setTagAssignmentsBulk,
} from '../../lib/api';
import { toast } from '../../lib/toast';

$: batchId = $incomingModalBatchId;
$: incoming = $allVideos.filter((v) => {
	if (!v.is_new) return false;
	if (batchId) return v.batch_id === batchId;
	return true;
});
$: folders = [...new Set([...getMoveFolderOptions(), '/'])].sort();

function close() {
	incomingModalOpen.set(false);
	incomingModalBatchId.set(null);
	incomingSelectedTagIDs.set(new Set());
	incomingSelectedFolder.set('/');
}

function toggleTag(tagID: string, checked: boolean) {
	incomingSelectedTagIDs.update((set) => {
		const next = new Set(set);
		if (checked) next.add(tagID);
		else next.delete(tagID);
		return next;
	});
}

async function markReviewed() {
	const paths = incoming.map((v) => v.path);
	if (!paths.length) {
		close();
		return;
	}
	const ok = await clearNewStatus(paths, true);
	if (!ok) {
		toast('Failed to mark incoming files as reviewed', 'error');
		return;
	}
	close();
	await refresh();
	await refreshBatches();
}

async function applyMassSort() {
	if (!incoming.length) {
		close();
		return;
	}
	const dest = get(incomingSelectedFolder);
	let errs = 0;
	const reviewedPaths: string[] = [];
	for (const video of incoming) {
		if ((video.folder || '/') === dest) {
			reviewedPaths.push(video.path);
			continue;
		}
		const res = await moveVideoApi(video.path, dest);
		if (!res.ok) errs++;
		else reviewedPaths.push(video.path);
	}
	const incomingHashes = [...new Set(incoming.map((v) => v.hash).filter(Boolean))];
	let tagErrs = 0;
	const tagIDs = get(incomingSelectedTagIDs);
	if (tagIDs.size && incomingHashes.length) {
		for (const tagID of tagIDs) {
			const ok = await setTagAssignmentsBulk(incomingHashes, tagID, true);
			if (!ok) tagErrs++;
		}
	}
	const ok = await clearNewStatus(reviewedPaths, false);
	if (!ok) toast('Moves completed but clear status failed', 'error');
	if (ok && !errs && !tagErrs && batchId) {
		filters.update((f) =>
			f.activeBatchId === batchId ? { ...f, activeBatchId: null } : f,
		);
	}
	close();
	await refresh();
	await refreshBatches();
	if (errs || tagErrs) {
		toast(
			(errs ? errs + ' move(s) failed. ' : '') +
				(tagErrs ? tagErrs + ' tag update(s) failed.' : ''),
			'error',
		);
	} else {
		toast('Incoming media mass sorted', 'success');
	}
}
</script>

<Modal id="incoming-modal" open={$incomingModalOpen} on:close={close}>
	<div class="move-inner">
		<div class="move-header">
			<div>
				<div class="move-title">Mass sort incoming media</div>
				<div class="move-subtitle" id="incoming-subtitle">
					{incoming.length} incoming file{incoming.length === 1 ? '' : 's'}
					{#if batchId}
						(in batch)
					{/if}
				</div>
			</div>
			<button class="close-btn" id="incoming-close" on:click={close}>✕</button>
		</div>
		<div class="move-body">
			<div>
				<div class="sidebar-label" style="padding-bottom:6px">Add tags to all incoming</div>
				<div id="incoming-tag-options">
					{#each $allTags as tag (tag.id)}
						{@const allSelected = incoming.length > 0 && incoming.every((v) => (v.tags || []).includes(tag.id))}
						<label class="tag-option">
							<input
								type="checkbox"
								checked={allSelected}
								on:change={(e) => toggleTag(tag.id, (e.target as HTMLInputElement).checked)}
							/>
							<span class="tag-option-name">
								<span class="tag-chip" style="--tag-color: {tag.color || '#888888'}">{tag.name}</span>
							</span>
						</label>
					{/each}
				</div>
			</div>
			<div>
				<div class="sidebar-label" style="padding-bottom:6px">Move all incoming to folder</div>
				<div id="incoming-folder-options">
					{#each folders as f (f)}
						<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
						<div
							class="folder-option"
							class:selected={$incomingSelectedFolder === f}
							on:click={() => incomingSelectedFolder.set(f)}
						>
							<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z" />
							</svg>
							{f === '/' ? 'Root (/)' : f}
						</div>
					{/each}
				</div>
			</div>
		</div>
		<div class="move-footer">
			<button class="btn" id="incoming-cancel-btn" on:click={close}>Close</button>
			<button class="btn active" id="incoming-confirm-btn" on:click={applyMassSort}>Apply mass sort</button>
			<button class="btn" id="incoming-mark-reviewed-btn" on:click={markReviewed}>Mark as resolved</button>
		</div>
	</div>
</Modal>
