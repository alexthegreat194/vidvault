<script lang="ts">
import { duplicateGroups, filters } from '../lib/stores/library';
import {
	autoOpenedDuplicates,
	dupModalOpen,
	duplicateBannerDismissed,
} from '../lib/stores/ui';

$: groups = $duplicateGroups.length;
$: files = $duplicateGroups.reduce((acc, g) => acc + g.members.length, 0);
$: shouldShow =
	groups > 0 && (!$duplicateBannerDismissed || $dupModalOpen);

function handleReview() {
	dupModalOpen.set(true);
}

function handleDismiss() {
	duplicateBannerDismissed.set(true);
}
</script>

<div id="dup-banner" class="dup-banner" class:visible={shouldShow}>
	<span class="dup-banner-text" id="dup-banner-text">
		{#if groups}
			{groups} duplicate group{groups === 1 ? '' : 's'} found ({files} files).
		{/if}
	</span>
	<div class="dup-banner-actions">
		<button class="btn active" id="dup-review-btn" on:click={handleReview}>
			Review duplicates
		</button>
		<button class="btn" id="dup-dismiss-btn" on:click={handleDismiss}>Dismiss</button>
	</div>
</div>
