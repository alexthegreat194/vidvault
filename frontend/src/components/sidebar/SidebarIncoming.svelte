<script lang="ts">
import { get } from 'svelte/store';
import { allVideos, countIncomingVideos, filters } from '../../lib/stores/library';
import { clearNewStatus, refresh } from '../../lib/api';
import { toast } from '../../lib/toast';

$: incomingCount = countIncomingVideos();
$: disabled = incomingCount === 0;
$: active = $filters.activeIncomingFilter;

$: if (disabled && $filters.activeIncomingFilter) {
	filters.update((f) => ({ ...f, activeIncomingFilter: false }));
}

function toggleIncomingFilter() {
	if (disabled) return;
	filters.update((f) => ({
		...f,
		activeIncomingFilter: !f.activeIncomingFilter,
		activeFolder: !f.activeIncomingFilter ? '__all__' : f.activeFolder,
	}));
}

async function markAllResolved() {
	const paths = get(allVideos)
		.filter((v) => Boolean(v.is_new))
		.map((v) => v.path);
	if (!paths.length) return;
	const ok = await clearNewStatus(paths, false);
	if (!ok) {
		toast('Failed to mark incoming files as reviewed', 'error');
		return;
	}
	await refresh();
	toast('Marked all incoming as reviewed', 'success');
}
</script>

<div
	class="sidebar-section"
	class:incoming-sidebar--disabled={disabled}
	id="incoming-sidebar-section"
>
	<div class="sidebar-header">
		<span class="sidebar-label">Incoming</span>
	</div>
	<div id="incoming-list">
		<button
			type="button"
			class="tag-btn need-sorting-nav"
			class:active
			disabled={disabled}
			aria-disabled={disabled ? 'true' : 'false'}
			title={disabled
				? 'No incoming files need sorting'
				: active
					? 'Show all videos'
					: 'Show only incoming (need sorting)'}
			on:click={toggleIncomingFilter}
		>
			<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<path d="M12 2v20" /><path d="M2 12h20" />
			</svg>
			<span>Need sorting</span>
			<span class="count">{incomingCount}</span>
		</button>
	</div>
	<div class="sidebar-inline-actions">
		<button
			class="btn"
			id="incoming-mark-all-btn"
			style:display={incomingCount > 0 ? '' : 'none'}
			on:click={markAllResolved}
		>
			Mark as resolved
		</button>
	</div>
</div>
