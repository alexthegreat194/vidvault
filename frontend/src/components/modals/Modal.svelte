<script lang="ts">
import { createEventDispatcher, onMount } from 'svelte';

export let open = false;
export let id = '';

const dispatch = createEventDispatcher<{ close: void }>();

function handleOverlayClick(e: MouseEvent) {
	if (e.target === e.currentTarget) dispatch('close');
}

function handleKeydown(e: KeyboardEvent) {
	if (e.key === 'Escape' && open) dispatch('close');
}

onMount(() => {
	document.addEventListener('keydown', handleKeydown);
	return () => document.removeEventListener('keydown', handleKeydown);
});
</script>

<div {id} class:open on:click={handleOverlayClick} role="presentation">
	<slot />
</div>
