<script lang="ts">
import { get } from 'svelte/store';
import {
	allTags,
	allVideos,
	filteredVideos,
	filters,
	refresh,
	refreshAfterTagDataChange,
	refreshTagsOnly,
	tagCountByID,
} from '../../lib/stores/library';
import { createTag, deleteTag, fetchTagAssignmentsForHashes } from '../../lib/api';
import { toast } from '../../lib/toast';

export let collapsed = false;

let showForm = false;
let tagInput = '';

function toggleTagFilter(tagID: string) {
	filters.update((f) => {
		const next = new Set(f.activeTagFilters);
		if (next.has(tagID)) next.delete(tagID);
		else next.add(tagID);
		return { ...f, activeTagFilters: next };
	});
}

async function removeTag(tagID: string, name: string, e: MouseEvent) {
	e.stopPropagation();
	if (!confirm(`Delete tag "${name}"?`)) return;
	const ok = await deleteTag(tagID);
	if (!ok) {
		toast('Tag delete failed', 'error');
		return;
	}
	filters.update((f) => {
		const next = new Set(f.activeTagFilters);
		next.delete(tagID);
		return { ...f, activeTagFilters: next };
	});
	try {
		const prevFilteredPaths = get(filteredVideos).map((v) => v.path);
		const affectedHashes = [
			...new Set(
				get(allVideos)
					.filter((v) => Array.isArray(v.tags) && v.tags.includes(tagID))
					.map((v) => v.hash)
					.filter(Boolean),
			),
		];
		await refreshTagsOnly();
		const payload = await fetchTagAssignmentsForHashes(affectedHashes);
		const updates = Array.isArray(payload?.videos) ? payload.videos : [];
		const tagsByHash = Object.fromEntries(
			updates.map((u) => [u.hash, Array.isArray(u.tags) ? u.tags : []]),
		);
		for (const video of get(allVideos)) {
			if (video?.hash && Object.prototype.hasOwnProperty.call(tagsByHash, video.hash)) {
				video.tags = tagsByHash[video.hash];
			}
		}
		await refreshAfterTagDataChange(prevFilteredPaths, affectedHashes);
	} catch {
		await refresh();
	}
}

function showCreateForm() {
	collapsed = false;
	showForm = true;
	tagInput = '';
}

function hideForm() {
	showForm = false;
	tagInput = '';
}

async function handleCreateTag() {
	const name = tagInput.trim();
	if (!name) return;
	const created = await createTag(name);
	if (!created) {
		toast('Failed: tag already exists or request failed', 'error');
		return;
	}
	hideForm();
	toast('Created tag ' + name, 'success');
	try {
		await refreshTagsOnly();
	} catch {
		await refresh();
	}
}

</script>

<div class="sidebar-section" class:collapsed>
	<div class="sidebar-header" id="tags-header">
		<span class="sidebar-label">Tags</span>
		<div class="sidebar-actions">
			<button class="new-folder-btn" id="new-tag-btn" title="New tag" on:click={showCreateForm}>+</button>
			<button class="collapse-btn" id="tags-collapse-btn" title="Collapse tags" on:click={() => (collapsed = !collapsed)}>
				▾
			</button>
		</div>
	</div>
	<div class="sidebar-tag-form" class:visible={showForm} id="sidebar-tag-form">
		<input
			id="sidebar-tag-input"
			placeholder="tag name…"
			autocomplete="off"
			bind:value={tagInput}
			on:keydown={(e) => {
				if (e.key === 'Enter') handleCreateTag();
				if (e.key === 'Escape') hideForm();
			}}
		/>
		<button id="sidebar-tag-confirm" title="Create" on:click={handleCreateTag}>✓</button>
		<button id="sidebar-tag-cancel" title="Cancel" on:click={hideForm}>✕</button>
	</div>
	<div id="tag-list">
		{#each $allTags as tag (tag.id)}
			<button
				type="button"
				class="tag-btn"
				class:active={$filters.activeTagFilters.has(tag.id)}
				on:click={() => toggleTagFilter(tag.id)}
			>
				<span class="tag-chip" style="--tag-color: {tag.color || '#888888'}">{tag.name}</span>
				<span class="count">{tagCountByID(tag.id)}</span>
				<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
				<span class="tag-remove-btn" title="Delete tag" on:click={(e) => removeTag(tag.id, tag.name, e)}>✕</span>
			</button>
		{/each}
	</div>
</div>
