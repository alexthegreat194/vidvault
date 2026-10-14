<script lang="ts">
import { activeBatch, allVideos, filters, refreshBatches } from '../../lib/stores/library';
import { incomingModalBatchId, incomingModalOpen } from '../../lib/stores/ui';
import { updateBatch } from '../../lib/api';
import { toast } from '../../lib/toast';

$: batch = $activeBatch;
$: batchVideos = batch
	? $allVideos.filter((v) => v.batch_id === batch.id)
	: [];
$: newCount = batchVideos.filter((v) => v.is_new).length;

function clearFilter() {
	filters.update((f) => ({ ...f, activeBatchId: null }));
}

function openMassSort() {
	if (!batch) return;
	incomingModalBatchId.set(batch.id);
	incomingModalOpen.set(true);
}

async function markDone() {
	if (!batch) return;
	try {
		await updateBatch(batch.id, { needs_sorting: false });
		await refreshBatches();
		toast('Batch marked done', 'success');
	} catch {
		toast('Failed to mark batch done', 'error');
	}
}
</script>

{#if batch}
	<div class="batch-header" id="batch-header">
		<div class="batch-header-info">
			<div class="batch-header-name">{batch.name}</div>
			<div class="batch-header-subtitle">
				{batchVideos.length} video{batchVideos.length === 1 ? '' : 's'}
				{#if newCount > 0}
					· {newCount} new
				{/if}
			</div>
		</div>
		<div class="batch-header-actions">
			<button class="btn" on:click={openMassSort}>Mass sort</button>
			{#if batch.needs_sorting}
				<button class="btn" on:click={markDone}>Mark batch done</button>
			{/if}
			<button class="btn" title="Clear filter" on:click={clearFilter}>×</button>
		</div>
	</div>
{/if}
