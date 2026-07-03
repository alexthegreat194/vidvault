<script lang="ts">
import { get } from 'svelte/store';
import {
	allCollections,
	allVideos,
	countCollectionResolvedVideos,
	refresh,
	refreshCollectionsOnly,
} from '../../lib/stores/library';
import {
	activeCollectionID,
	watchCollectionOpen,
} from '../../lib/stores/ui';
import { createCollection, deleteCollection, renameCollection } from '../../lib/api';
import { toast } from '../../lib/toast';
import type { Collection } from '../../lib/types';

export let collapsed = false;

let showForm = false;
let collectionInput = '';

function showCreateForm() {
	collapsed = false;
	showForm = true;
	collectionInput = '';
}

function hideForm() {
	showForm = false;
	collectionInput = '';
}

async function createCollectionLocal() {
	const name = collectionInput.trim();
	if (!name) return;
	const created = await createCollection(name);
	if (!created) {
		toast('Failed: collection already exists or request failed', 'error');
		return;
	}
	hideForm();
	toast('Created collection ' + name, 'success');
	try {
		await refreshCollectionsOnly();
	} catch {
		await refresh();
	}
}

function openWatch(collection: Collection) {
	activeCollectionID.set(collection.id);
	watchCollectionOpen.set(true);
}

async function handleDelete(collection: Collection, e: MouseEvent) {
	e.stopPropagation();
	if (!confirm(`Delete collection "${collection.name}"?`)) return;
	const ok = await deleteCollection(collection.id);
	if (!ok) {
		toast('Collection delete failed', 'error');
		return;
	}
	if (get(activeCollectionID) === collection.id) {
		watchCollectionOpen.set(false);
		activeCollectionID.set('');
	}
	await refresh();
	toast('Deleted ' + collection.name, 'success');
}

async function handleRename(collection: Collection, e: MouseEvent) {
	e.stopPropagation();
	const nextName = prompt('Rename collection', collection.name);
	if (!nextName?.trim()) return;
	const ok = await renameCollection(collection.id, nextName.trim());
	if (!ok) {
		toast('Collection rename failed', 'error');
		return;
	}
	await refresh();
	if (get(activeCollectionID) === collection.id) {
		watchCollectionOpen.set(true);
	}
}
</script>

<div class="sidebar-section" class:collapsed>
	<div class="sidebar-header" id="collections-header">
		<span class="sidebar-label">Collections</span>
		<div class="sidebar-actions">
			<button class="new-folder-btn" id="new-collection-btn" title="New collection" on:click={showCreateForm}>+</button>
			<button class="collapse-btn" id="collections-collapse-btn" title="Collapse collections" on:click={() => (collapsed = !collapsed)}>
				▾
			</button>
		</div>
	</div>
	<div class="sidebar-collection-form" class:visible={showForm} id="sidebar-collection-form">
		<input
			id="sidebar-collection-input"
			placeholder="collection name…"
			autocomplete="off"
			bind:value={collectionInput}
			on:keydown={(e) => {
				if (e.key === 'Enter') createCollectionLocal();
				if (e.key === 'Escape') hideForm();
			}}
		/>
		<button id="sidebar-collection-confirm" title="Create" on:click={createCollectionLocal}>✓</button>
		<button id="sidebar-collection-cancel" title="Cancel" on:click={hideForm}>✕</button>
	</div>
	<div id="collection-list">
		{#each $allCollections as collection (collection.id)}
			<div
				class="collection-btn"
				class:active={$activeCollectionID === collection.id}
				role="button"
				tabindex="0"
				on:click={() => openWatch(collection)}
				on:keydown={(e) => e.key === 'Enter' && openWatch(collection)}
			>
				<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<rect x="3" y="3" width="7" height="7" /><rect x="14" y="3" width="7" height="7" /><rect x="14" y="14" width="7" height="7" /><rect x="3" y="14" width="7" height="7" />
				</svg>
				<span>{collection.name}</span>
				<span class="count">{countCollectionResolvedVideos(collection)}</span>
				<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
				<button class="collection-rename-btn" title="Rename collection" on:click|stopPropagation={(e) => handleRename(collection, e)}>✎</button>
				<!-- svelte-ignore a11y_click_events_have_key_events -->
				<button class="collection-remove-btn" title="Delete collection" on:click|stopPropagation={(e) => handleDelete(collection, e)}>✕</button>
			</div>
		{/each}
	</div>
</div>
