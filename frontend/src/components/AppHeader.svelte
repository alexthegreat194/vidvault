<script lang="ts">
import {
	allVideos,
	filteredVideos,
	filters,
	isVideoDiscoveryLoading,
} from '../lib/stores/library';
import {
	galleryListView,
	showPinLockBtn,
	uploadModalOpen,
} from '../lib/stores/ui';
import {
	enterSelectMode,
	exitSelectMode,
	selectMode,
} from '../lib/stores/selection';
import { lockSession } from '../lib/api';
import {
	pinGateUnlocked,
	setPinOverlayVisible,
	unloadLibraryAfterPinLock,
} from '../lib/stores/library';
import { setPinGateUnlocked } from '../lib/auth';

function handleSearchInput(e: Event) {
	const value = (e.target as HTMLInputElement).value;
	filters.update((f) => ({ ...f, search: value }));
}

function handleSortChange(e: Event) {
	const value = (e.target as HTMLSelectElement).value;
	filters.update((f) => ({ ...f, sort: value }));
}

function handleFavoritesChange(e: Event) {
	const checked = (e.target as HTMLInputElement).checked;
	filters.update((f) => ({ ...f, showFavoritesOnly: checked }));
}

function handleGridView() {
	galleryListView.set(false);
}

function handleListView() {
	galleryListView.set(true);
}

function toggleSelect() {
	if ($selectMode) exitSelectMode();
	else enterSelectMode();
}

async function handleLock() {
	try {
		const ok = await lockSession();
		if (!ok) return;
		unloadLibraryAfterPinLock();
		pinGateUnlocked.set(false);
		setPinGateUnlocked(false);
		setPinOverlayVisible(true);
	} catch (e) {
		console.error(e);
	}
}
</script>

<header>
	<div class="logo">VID<span>VAULT</span></div>
	<div class="stats" id="stats">
		{#if $isVideoDiscoveryLoading}
			<div class="stats-loading">
				<span class="stats-loading-label">
					<b>{$filteredVideos.length}</b> / {$allVideos.length} videos · scanning library ({$allVideos.length} found)
				</span>
				<span class="stats-loading-bar" aria-hidden="true">
					<span class="stats-loading-fill"></span>
				</span>
			</div>
		{:else}
			<b>{$filteredVideos.length}</b> / {$allVideos.length} videos
		{/if}
	</div>
	<div class="controls">
		<label class="favorites-only-toggle" title="Show favorites only">
			<span class="favorites-switch">
				<input
					type="checkbox"
					id="favorites-only-toggle"
					class="favorites-switch-input"
					checked={$filters.showFavoritesOnly}
					on:change={handleFavoritesChange}
				/>
				<span class="favorites-switch-ui" aria-hidden="true">
					<span class="favorites-switch-track"></span>
					<span class="favorites-switch-thumb">
						<svg class="favorites-switch-star" xmlns="http://www.w3.org/2000/svg" width="11" height="11" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
							<path d="M12 2l2.4 7.4h7.6l-6 4.6 2.3 7-6.3-4.6-6.3 4.6 2.3-7-6-4.6h7.6z" />
						</svg>
					</span>
				</span>
			</span>
			<span class="favorites-only-label">Favorites only</span>
		</label>
		<div class="search-wrap">
			<svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<circle cx="11" cy="11" r="8" /><path d="m21 21-4.35-4.35" />
			</svg>
			<input
				id="search"
				type="search"
				placeholder="filter files…"
				autocomplete="off"
				value={$filters.search}
				on:input={handleSearchInput}
			/>
		</div>
		<select class="btn" id="sort-select" value={$filters.sort} on:change={handleSortChange}>
			<option value="name">name ↑</option>
			<option value="name-desc">name ↓</option>
			<option value="folder">folder</option>
			<option value="ext">type</option>
			<option value="modified-desc">modified (newest)</option>
			<option value="modified">modified (oldest)</option>
			<option value="size-desc">size (largest)</option>
			<option value="size">size (smallest)</option>
		</select>
		<div class="btn-group">
			<button
				class="btn"
				class:active={!$galleryListView}
				id="grid-btn"
				title="Grid view"
				on:click={handleGridView}
			>
				⊞
			</button>
			<button
				class="btn"
				class:active={$galleryListView}
				id="list-btn"
				title="List view"
				on:click={handleListView}
			>
				☰
			</button>
		</div>
		<button
			class="btn"
			class:active={$selectMode}
			id="select-btn"
			title="Select multiple"
			on:click={toggleSelect}
		>
			Select
		</button>
		{#if $showPinLockBtn}
			<button
				type="button"
				class="btn pin-lock-btn"
				id="pin-lock-btn"
				title="Lock session"
				on:click={handleLock}
			>
				<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
					<rect x="3" y="11" width="18" height="11" rx="2" ry="2" /><path d="M7 11V7a5 5 0 0 1 10 0v4" />
				</svg>
				Lock
			</button>
		{/if}
		<button class="btn upload-btn" id="open-upload-btn" on:click={() => uploadModalOpen.set(true)}>
			<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" /><polyline points="17 8 12 3 7 8" /><line x1="12" y1="3" x2="12" y2="15" />
			</svg>
			Upload
		</button>
	</div>
</header>
