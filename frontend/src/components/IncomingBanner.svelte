<script lang="ts">
import { countIncomingVideos, filters } from '../lib/stores/library';
import {
	incomingBannerDismissed,
	incomingModalOpen,
} from '../lib/stores/ui';

$: incomingCount = countIncomingVideos();
$: shouldShow =
	incomingCount > 0 && (!$incomingBannerDismissed || $incomingModalOpen);

function handleReview() {
	filters.update((f) => ({
		...f,
		activeIncomingFilter: true,
		activeFolder: '__all__',
	}));
}

function handleMassSort() {
	incomingModalOpen.set(true);
}

function handleDismiss() {
	incomingBannerDismissed.set(true);
}
</script>

<div id="incoming-banner" class="incoming-banner" class:visible={shouldShow}>
	<span class="incoming-banner-text" id="incoming-banner-text">
		{#if incomingCount}
			{incomingCount} new file{incomingCount === 1 ? '' : 's'} need sorting.
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
