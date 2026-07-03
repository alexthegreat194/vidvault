<script lang="ts">
import { get } from 'svelte/store';
import Modal from './Modal.svelte';
import {
	allVideos,
	collectionMap,
	refresh,
	resolveCollectionVideos,
} from '../../lib/stores/library';
import {
	activeCollectionID,
	watchCollectionMasterMuted,
	watchCollectionMasterVolume,
	watchCollectionOpen,
	watchGridCols,
} from '../../lib/stores/ui';
import { DEFAULT_LOOP_ALL_VIDEOS } from '../../lib/constants';
import { deleteCollection, renameCollection, setCollectionVideo } from '../../lib/api';
import { toast } from '../../lib/toast';

let tileVideos: HTMLVideoElement[] = [];

$: collection = $activeCollectionID ? $collectionMap[$activeCollectionID] : undefined;
$: resolved = collection ? resolveCollectionVideos(collection) : { ordered: [], missing: [] };

function close() {
	watchCollectionOpen.set(false);
	activeCollectionID.set('');
	tileVideos.forEach((v) => {
		v.pause();
		v.src = '';
	});
	tileVideos = [];
}

function applyAudio(video: HTMLVideoElement) {
	video.volume = get(watchCollectionMasterVolume);
	video.muted = get(watchCollectionMasterMuted);
}

function bindTileVideo(node: HTMLVideoElement) {
	node.loop = DEFAULT_LOOP_ALL_VIDEOS;
	applyAudio(node);
	tileVideos = [...tileVideos, node];
	return {
		destroy() {
			tileVideos = tileVideos.filter((v) => v !== node);
		},
	};
}

$: if ($watchCollectionMasterVolume !== undefined || $watchCollectionMasterMuted !== undefined) {
	for (const v of tileVideos) applyAudio(v);
}

function playAll() {
	for (const v of tileVideos) v.play().catch(() => {});
}

function pauseAll() {
	for (const v of tileVideos) tileVideos.forEach((v) => v.pause());
}

function toggleMute() {
	watchCollectionMasterMuted.update((m) => !m);
}

async function removeFromCollection(hash: string) {
	if (!collection) return;
	const ok = await setCollectionVideo(collection.id, hash, false);
	if (!ok) {
		toast('Remove failed', 'error');
		return;
	}
	await refresh();
}

async function handleDelete() {
	if (!collection) return;
	if (!confirm(`Delete collection "${collection.name}"?`)) return;
	const ok = await deleteCollection(collection.id);
	if (!ok) {
		toast('Collection delete failed', 'error');
		return;
	}
	close();
	await refresh();
}

async function handleRename() {
	if (!collection) return;
	const nextName = prompt('Rename collection', collection.name);
	if (!nextName?.trim()) return;
	const ok = await renameCollection(collection.id, nextName.trim());
	if (!ok) {
		toast('Collection rename failed', 'error');
		return;
	}
	await refresh();
}
</script>

<Modal id="watch-collection-modal" open={$watchCollectionOpen} on:close={close}>
	{#if collection}
		<div class="watch-collection-inner">
			<div class="watch-collection-header">
				<div>
					<div class="watch-collection-title" id="watch-collection-title">{collection.name}</div>
					<div class="watch-collection-subtitle" id="watch-collection-subtitle">
						{resolved.ordered.length} video{resolved.ordered.length === 1 ? '' : 's'}
						{resolved.missing.length ? ' • ' + resolved.missing.length + ' missing' : ''}
					</div>
				</div>
				<div class="watch-collection-actions">
					<button class="btn" id="watch-collection-play-all-btn" on:click={playAll}>Play all</button>
					<button class="btn" id="watch-collection-pause-all-btn" on:click={pauseAll}>Pause all</button>
					<button class="btn" id="watch-collection-mute-all-btn" on:click={toggleMute}>
						{$watchCollectionMasterMuted ? 'Unmute all' : 'Mute all'}
					</button>
					<label class="watch-volume-control" for="watch-collection-volume-input">
						Vol
						<input id="watch-collection-volume-input" type="range" min="0" max="1" step="0.05" bind:value={$watchCollectionMasterVolume} />
					</label>
					<label class="watch-grid-control" for="watch-collection-grid-cols">
						Grid
						<select class="btn" id="watch-collection-grid-cols" bind:value={$watchGridCols}>
							<option value={4}>4</option>
							<option value={3}>3</option>
							<option value={2}>2</option>
							<option value={1}>1</option>
						</select>
					</label>
					<button class="btn" id="watch-collection-rename-btn" on:click={handleRename}>Rename</button>
					<button class="btn" id="watch-collection-delete-btn" on:click={handleDelete}>Delete</button>
					<button class="close-btn" id="watch-collection-close" on:click={close}>✕</button>
				</div>
			</div>
			<div class="watch-collection-body" id="watch-collection-body">
				{#if !resolved.ordered.length}
					<div class="watch-collection-empty">No playable videos in this collection yet.</div>
				{:else}
					<div class="watch-grid cols-{$watchGridCols}">
						{#each resolved.ordered as v (v.path)}
							<div class="watch-tile">
								<!-- svelte-ignore a11y_media_has_caption -->
							<video
									use:bindTileVideo
									controls
									loop
									preload="metadata"
									src={'/video?path=' + encodeURIComponent(v.path)}
								></video>
								<div class="watch-tile-meta">
									<span class="watch-tile-name" title={v.name}>{v.name}</span>
									<span class="watch-tile-path">{v.folder || '/'}</span>
									<button class="watch-tile-remove" title="Remove from collection" on:click={() => removeFromCollection(v.hash)}>✕</button>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>
	{/if}
</Modal>
