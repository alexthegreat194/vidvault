<script lang="ts">
import { get } from 'svelte/store';
import Modal from './Modal.svelte';
import {
	allTags,
	allVideos,
	filteredVideos,
	refresh,
	refreshAfterTagDataChange,
	refreshTagsOnly,
	refreshTagAssignmentsForHashes,
	tagMap,
} from '../../lib/stores/library';
import {
	pendingTagActions,
	tagModalOpen,
	tagTargetPaths,
} from '../../lib/stores/ui';
import { createTag, deleteTag, setTagAssignmentsBulk } from '../../lib/api';
import { exitSelectMode, selectMode } from '../../lib/stores/selection';
import { toast } from '../../lib/toast';

let tagNewInput = '';

$: targets = $tagTargetPaths
	.map((path) => get(allVideos).find((v) => v.path === path))
	.filter(Boolean);

$: subtitle =
	targets.length > 1
		? targets.length + ' videos selected'
		: targets[0]?.name || '';

function close() {
	tagModalOpen.set(false);
	tagTargetPaths.set([]);
	pendingTagActions.set(new Map());
	tagNewInput = '';
}

function getPendingEntries(): [string, 'add' | 'remove'][] {
	const valid = new Set(get(allTags).map((t) => t.id));
	const entries: [string, 'add' | 'remove'][] = [];
	for (const [tagID, action] of get(pendingTagActions).entries()) {
		if (!valid.has(tagID)) continue;
		if (action === 'add' || action === 'remove') entries.push([tagID, action]);
	}
	return entries;
}

$: entries = getPendingEntries();
$: addCount = entries.filter(([, a]) => a === 'add').length;
$: removeCount = entries.filter(([, a]) => a === 'remove').length;
$: pendingSummary = (() => {
	const parts: string[] = [];
	if (addCount) parts.push(addCount + ' add');
	if (removeCount) parts.push(removeCount + ' remove');
	return parts.length ? 'Pending: ' + parts.join(', ') : 'No pending tag changes';
})();

function togglePending(tagID: string, action: 'add' | 'remove') {
	pendingTagActions.update((map) => {
		const next = new Map(map);
		if (next.get(tagID) === action) next.delete(tagID);
		else next.set(tagID, action);
		return next;
	});
}

async function createTagLocal() {
	const name = tagNewInput.trim();
	if (!name) return;
	const created = await createTag(name);
	if (!created) {
		toast('Tag create failed', 'error');
		return;
	}
	tagNewInput = '';
	try {
		await refreshTagsOnly();
	} catch {
		await refresh();
	}
}

async function deleteTagLocal(tagID: string, name: string) {
	if (!confirm(`Delete tag "${name}"?`)) return;
	const ok = await deleteTag(tagID);
	if (!ok) {
		toast('Tag delete failed', 'error');
		return;
	}
	// handled like original
	try {
		await refreshTagsOnly();
		await refresh();
	} catch {
		await refresh();
	}
	pendingTagActions.update((m) => {
		const next = new Map(m);
		next.delete(tagID);
		return next;
	});
}

async function applyChanges() {
	const ents = getPendingEntries();
	if (!ents.length) return;
	const hashes = [...new Set(targets.map((v) => v.hash).filter(Boolean))];
	if (!hashes.length) {
		toast('No taggable videos selected', 'error');
		return;
	}
	const addingNames = ents
		.filter(([, a]) => a === 'add')
		.map(([id]) => get(tagMap)[id]?.name || id);
	const removingNames = ents
		.filter(([, a]) => a === 'remove')
		.map(([id]) => get(tagMap)[id]?.name || id);
	const lines = ['Apply tag changes to ' + hashes.length + ' videos?'];
	if (addingNames.length) lines.push('Add: ' + addingNames.join(', '));
	if (removingNames.length) lines.push('Remove: ' + removingNames.join(', '));
	if (!confirm(lines.join('\n'))) return;

	const prevFilteredPaths = get(filteredVideos).map((v) => v.path);
	let failures = 0;
	for (const [tagID, action] of ents) {
		const ok = await setTagAssignmentsBulk(hashes, tagID, action === 'add');
		if (!ok) failures++;
	}
	if (failures) {
		toast(failures + ' tag update(s) failed', 'error');
		await refresh();
		return;
	}
	try {
		await refreshTagsOnly();
		await refreshTagAssignmentsForHashes(hashes);
		await refreshAfterTagDataChange(prevFilteredPaths, hashes);
	} catch {
		await refresh();
	}
	pendingTagActions.set(new Map());
	toast('Tag changes applied', 'success');
	if (get(selectMode)) {
		close();
		exitSelectMode();
	}
}
</script>

<Modal id="tag-modal" open={$tagModalOpen} on:close={close}>
	<div class="tag-inner">
		<div class="tag-header">
			<div>
				<div class="tag-title">Edit tags</div>
				<div class="tag-subtitle" id="tag-subtitle">{subtitle}</div>
			</div>
			<button class="close-btn" id="tag-close" on:click={close}>✕</button>
		</div>
		<div class="tag-body">
			<div class="tag-new-row">
				<input id="tag-new-input" placeholder="+ new tag name…" autocomplete="off" bind:value={tagNewInput} on:keydown={(e) => e.key === 'Enter' && createTagLocal()} />
				<button class="btn" id="tag-new-btn" on:click={createTagLocal}>Create</button>
			</div>
			<div id="tag-options">
				{#each $allTags as tag (tag.id)}
					{@const selectedCount = targets.filter((v) => (v.tags || []).includes(tag.id)).length}
					{@const pending = $pendingTagActions.get(tag.id)}
					{@const stateText =
						pending === 'add'
							? 'pending add'
							: pending === 'remove'
								? 'pending remove'
								: selectedCount === 0
									? 'none selected'
									: selectedCount === targets.length
										? 'all selected'
										: selectedCount + '/' + targets.length + ' selected'}
					<div class="tag-option">
						<span class="tag-option-name">
							<span class="tag-chip" style="--tag-color: {tag.color || '#888888'}">{tag.name}</span>
							<span class="tag-option-state">{stateText}</span>
						</span>
						<button class="btn tag-option-action" class:active={pending === 'add'} type="button" on:click={() => togglePending(tag.id, 'add')}>Add</button>
						<button class="btn tag-option-action" class:active={pending === 'remove'} type="button" on:click={() => togglePending(tag.id, 'remove')}>Remove</button>
						<button class="tag-option-remove" type="button" on:click={() => deleteTagLocal(tag.id, tag.name)}>delete</button>
					</div>
				{/each}
			</div>
		</div>
		<div class="tag-footer">
			<span class="tag-pending-summary" id="tag-pending-summary">{pendingSummary}</span>
			<button class="btn active" id="tag-apply-btn" disabled={entries.length === 0} on:click={applyChanges}>Apply changes</button>
			<button class="btn" id="tag-cancel-btn" on:click={close}>Close</button>
		</div>
	</div>
</Modal>
