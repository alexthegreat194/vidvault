<script lang="ts">
import { batchesNeedingSort, filters } from '../lib/stores/library';
import {
	incomingBannerDismissed,
	incomingModalBatchId,
	incomingModalOpen,
} from '../lib/stores/ui';

$: topBatch = $batchesNeedingSort[0] ?? null;
$: shouldShow =
	!!topBatch && (!$incomingBannerDismissed || $incomingModalOpen);

function handleReview() {
	if (!topBatch) return;
	filters.update((f) => ({
		...f,
		activeBatchId: topBatch.id,
		activeFolder: '__all__',
	}));
}

function handleMassSort() {
	if (!topBatch) return;
	incomingModalBatchId.set(topBatch.id);
	incomingModalOpen.set(true);
}

function handleDismiss() {
	incomingBannerDismissed.set(true);
}
</script>

<div id="incoming-banner" class="incoming-banner" class:visible={shouldShow}>
	<span class="incoming-banner-text" id="incoming-banner-text">
		{#if topBatch}
			{topBatch.new_count} new file{topBatch.new_count === 1 ? '' : 's'} in {topBatch.name} need sorting.
		{/if}
	</span>
	<div class="incoming-banner-actions">
		<button class="btn" id="incoming-review-btn" on:click={handleReview}>View</button>
		<button class="btn active" id="incoming-mass-sort-btn" on:click={handleMassSort}>
			Mass sort
		</button>
		<button class="btn" id="incoming-dismiss-btn" on:click={handleDismiss}>Dismiss</button>
	</div>
</div>
