<script lang="ts">
import { get } from 'svelte/store';
import Modal from './Modal.svelte';
import {
	allCollections,
	countCollectionResolvedVideos,
	refresh,
	refreshCollectionsOnly,
} from '../../lib/stores/library';
import {
	activeCollectionID,
	collectionPickerOpen,
	collectionPickerSelectedID,
	collectionPickerTargetHashes,
	watchCollectionOpen,
} from '../../lib/stores/ui';
import { bulkAddCollectionVideos, createCollection } from '../../lib/api';
import { exitSelectMode, selectMode } from '../../lib/stores/selection';
import { toast } from '../../lib/toast';

let newCollectionName = '';

$: if ($collectionPickerOpen && !get(collectionPickerSelectedID) && $allCollections.length) {
	collectionPickerSelectedID.set($allCollections[0].id);
}

function close() {
	collectionPickerOpen.set(false);
	collectionPickerTargetHashes.set([]);
	collectionPickerSelectedID.set('');
	newCollectionName = '';
}

async function createNew() {
	const name = newCollectionName.trim();
	if (!name) return;
	const created = await createCollection(name);
	if (!created) {
		toast('Create collection failed', 'error');
		return;
	}
	newCollectionName = '';
	try {
		await refreshCollectionsOnly();
	} catch {
		await refresh();
	}
	collectionPickerSelectedID.set(created.id);
}

async function confirm() {
	const id = get(collectionPickerSelectedID);
	const hashes = get(collectionPickerTargetHashes);
	if (!id || !hashes.length) return;
	const ok = await bulkAddCollectionVideos(id, hashes);
	if (!ok) {
		toast('Add to collection failed', 'error');
		return;
	}
	try {
		await refreshCollectionsOnly();
		if (get(activeCollectionID) === id) watchCollectionOpen.set(true);
	} catch {
		await refresh();
	}
	close();
	if (get(selectMode)) exitSelectMode();
	toast('Added to collection', 'success');
}
</script>

<Modal id="collection-picker-modal" open={$collectionPickerOpen} on:close={close}>
	<div class="collection-picker-inner">
		<div class="collection-picker-header">
			<div>
				<div class="collection-picker-title">Add to collection</div>
				<div class="collection-picker-subtitle" id="collection-picker-subtitle">
					{$collectionPickerTargetHashes.length} video{$collectionPickerTargetHashes.length === 1 ? '' : 's'} selected
				</div>
			</div>
			<button class="close-btn" id="collection-picker-close" on:click={close}>✕</button>
		</div>
		<div class="collection-picker-body">
			<div class="collection-new-row">
				<input id="collection-picker-new-input" placeholder="+ new collection name…" autocomplete="off" bind:value={newCollectionName} on:keydown={(e) => e.key === 'Enter' && createNew()} />
				<button class="btn" id="collection-picker-new-btn" on:click={createNew}>Create</button>
			</div>
			<div id="collection-picker-options">
				{#each $allCollections as collection (collection.id)}
					<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
					<div
						class="collection-picker-option"
						class:selected={$collectionPickerSelectedID === collection.id}
						on:click={() => collectionPickerSelectedID.set(collection.id)}
					>
						<span>{collection.name}</span>
						<span class="count">{countCollectionResolvedVideos(collection)}</span>
					</div>
				{/each}
			</div>
		</div>
		<div class="collection-picker-footer">
			<button class="btn" id="collection-picker-cancel-btn" on:click={close}>Cancel</button>
			<button class="btn active" id="collection-picker-confirm-btn" on:click={confirm}>Add selected</button>
		</div>
	</div>
</Modal>
