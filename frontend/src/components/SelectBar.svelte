<script lang="ts">
import { get } from 'svelte/store';
import {
	collectionPickerOpen,
	collectionPickerTargetHashes,
	moveModalOpen,
	moveTarget,
	tagModalOpen,
	tagTargetPaths,
} from '../lib/stores/ui';
import {
	exitSelectMode,
	selectedPaths,
	selectAllPaths,
	clearSelection,
} from '../lib/stores/selection';
import { filteredVideos, refresh } from '../lib/stores/library';
import { forgetVideos } from '../lib/api';
import { toast } from '../lib/toast';

$: count = $selectedPaths.size;
$: visible = count > 0;

function openMove() {
	if (!count) return;
	const first = get(filteredVideos).find((v) => $selectedPaths.has(v.path));
	if (first) {
		moveTarget.set(first);
		moveModalOpen.set(true);
	}
}

async function handleForget() {
	if (!count) return;
	const paths = [...$selectedPaths];
	const ok = await forgetVideos(paths, [], false);
	if (!ok) {
		toast('Forget operation failed', 'error');
		return;
	}
	await refresh();
	exitSelectMode();
	toast('Forgot ' + paths.length + ' file(s)', 'success');
}

function openTags() {
	if (!count) return;
	const targets = get(filteredVideos).filter((v) => $selectedPaths.has(v.path));
	if (!targets.length) return;
	tagTargetPaths.set(targets.map((v) => v.path));
	tagModalOpen.set(true);
}

function openCollectionPicker() {
	if (!count) return;
	const targets = get(filteredVideos).filter((v) => $selectedPaths.has(v.path));
	const hashes = [...new Set(targets.map((v) => v.hash).filter(Boolean))];
	if (!hashes.length) {
		toast('No valid hashes found in selection', 'error');
		return;
	}
	collectionPickerTargetHashes.set(hashes);
	collectionPickerOpen.set(true);
}

function handleSelectAll() {
	selectAllPaths(get(filteredVideos).map((v) => v.path));
}

function handleClear() {
	clearSelection();
}
</script>

<div class="select-bar" id="select-bar" class:visible={visible}>
	<span class="select-count" id="select-count">{count} selected</span>
	<button class="btn" id="select-tag-btn" on:click={openTags}>Tags…</button>
	<button class="btn" id="select-collection-btn" on:click={openCollectionPicker}>Add to collection…</button>
	<button class="btn active" id="select-move-btn" on:click={openMove}>Move…</button>
	<button class="btn" id="select-forget-btn" on:click={handleForget}>Forget…</button>
	<button class="btn" id="select-all-btn" on:click={handleSelectAll}>All</button>
	<button class="btn" id="select-clear-btn" on:click={handleClear}>Clear</button>
</div>
