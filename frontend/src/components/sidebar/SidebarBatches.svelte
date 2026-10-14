<script lang="ts">
import { filters, refreshBatches, uploadBatches } from '../../lib/stores/library';
import { updateBatch } from '../../lib/api';
import { toast } from '../../lib/toast';
import type { UploadBatch } from '../../lib/types';

export let collapsed = false;

let showDone = true;
let editingId: string | null = null;
let editName = '';

$: visibleBatches = showDone
	? $uploadBatches
	: $uploadBatches.filter((b) => b.needs_sorting);

function selectBatch(batch: UploadBatch) {
	if (editingId) return;
	filters.update((f) => ({
		...f,
		activeBatchId: f.activeBatchId === batch.id ? null : batch.id,
		activeFolder: '__all__',
	}));
}

function startRename(batch: UploadBatch, e?: MouseEvent) {
	e?.stopPropagation();
	editingId = batch.id;
	editName = batch.name;
}

async function saveRename(batch: UploadBatch) {
	const trimmed = editName.trim();
	editingId = null;
	try {
		await updateBatch(batch.id, { name: trimmed });
		await refreshBatches();
	} catch {
		toast('Failed to rename batch', 'error');
	}
}

function cancelRename() {
	editingId = null;
	editName = '';
}

function compactBatchLabel(batch: UploadBatch): string {
	if (batch.name && !/^\w{3} \d{1,2}, \d{1,2}:\d{2} [AP]M$/.test(batch.name)) {
		return batch.name;
	}
	const d = new Date(batch.created_at);
	if (Number.isNaN(d.getTime())) return batch.name;
	const mon = d.toLocaleString('en-US', { month: 'short' });
	const day = d.getDate();
	const h = d.getHours() % 12 || 12;
	const m = String(d.getMinutes()).padStart(2, '0');
	const ap = d.getHours() >= 12 ? 'p' : 'a';
	return `${mon} ${day} ${h}:${m}${ap}`;
}
</script>

<div class="sidebar-section" class:collapsed>
	<div class="sidebar-header" id="batches-header">
		<span class="sidebar-label">Batches</span>
		<div class="sidebar-actions">
			<label class="batch-show-done" title="Show completed batches">
				<input type="checkbox" bind:checked={showDone} />
			</label>
			<button
				class="collapse-btn"
				id="batches-collapse-btn"
				title="Collapse upload batches"
				on:click={() => (collapsed = !collapsed)}
			>
				▾
			</button>
		</div>
	</div>
	<div id="batch-list">
		{#each visibleBatches as batch (batch.id)}
			<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
			<div
				class="batch-btn"
				class:active={$filters.activeBatchId === batch.id}
				class:batch-done={!batch.needs_sorting}
				role="button"
				tabindex="0"
				on:click={() => selectBatch(batch)}
				on:dblclick={() => startRename(batch)}
				on:keydown={(e) => e.key === 'Enter' && selectBatch(batch)}
			>
				<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<polyline points="22 12 16 12 14 15 10 15 8 12 2 12" />
					<path d="M5.45 5.11L2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.45-6.89A2 2 0 0 0 16.76 4H7.24a2 2 0 0 0-1.79 1.11z" />
				</svg>
				{#if editingId === batch.id}
					<input
						class="batch-rename-input"
						bind:value={editName}
						on:click|stopPropagation
						on:keydown={(e) => {
							if (e.key === 'Enter') saveRename(batch);
							if (e.key === 'Escape') cancelRename();
						}}
						on:blur={() => saveRename(batch)}
					/>
				{:else}
					<span class="batch-name" title={batch.name}>{compactBatchLabel(batch)}</span>
				{/if}
				<span
					class="count"
					class:has-new={batch.new_count > 0}
					title={batch.new_count > 0 ? `${batch.new_count} new` : undefined}
				>
					{batch.new_count > 0 ? batch.new_count : batch.video_count}
				</span>
				<button class="batch-rename-btn" title="Rename batch" on:click|stopPropagation={(e) => startRename(batch, e)}>✎</button>
			</div>
		{/each}
	</div>
</div>
