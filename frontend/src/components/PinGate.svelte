<script lang="ts">
import { submitPin } from '../lib/api';
import {
	init,
	pinGateError,
	pinGateMode,
	pinGateUnlocked,
	pinGateVisible,
	setPinOverlayVisible,
	showPinLockBtn,
} from '../lib/stores/library';
import { setPinGateUnlocked } from '../lib/auth';

let pinInput = '';
let inputEl: HTMLInputElement | undefined;

$: if ($pinGateVisible && inputEl) {
	inputEl.readOnly = true;
	queueMicrotask(() => {
		if (!inputEl) return;
		inputEl.focus();
		inputEl.readOnly = false;
	});
}

async function handleSubmit(e: Event) {
	e.preventDefault();
	pinGateError.set('');
	try {
		const res = await submitPin(pinInput);
		if (!res.ok) {
			setPinOverlayVisible(true, 'Incorrect PIN.');
			if (inputEl) {
				inputEl.readOnly = false;
				inputEl.focus();
				inputEl.select();
			}
			return;
		}
		pinGateUnlocked.set(true);
		setPinGateUnlocked(true);
		setPinOverlayVisible(false);
		showPinLockBtn.set(true);
		await init();
	} catch (err) {
		console.error(err);
		setPinOverlayVisible(true, 'Could not reach server.');
		if (inputEl) inputEl.readOnly = false;
	}
}
</script>

{#if $pinGateMode && $pinGateVisible}
	<div
		id="pin-gate-overlay"
		class="pin-gate-overlay"
		aria-hidden="false"
	>
		<div class="pin-gate-dialog" role="dialog" aria-labelledby="pin-gate-title">
			<div class="logo pin-gate-logo">VID<span>VAULT</span></div>
			<p id="pin-gate-title" class="pin-gate-title">Enter PIN to continue</p>
			<form
				id="pin-gate-form"
				class="pin-gate-form"
				autocomplete="off"
				data-lpignore="true"
				data-form-type="other"
				on:submit={handleSubmit}
			>
				<input
					bind:this={inputEl}
					type="password"
					id="pin-gate-input"
					class="pin-gate-input"
					name="vv_pin"
					bind:value={pinInput}
					autocomplete="off"
					autocorrect="off"
					autocapitalize="off"
					spellcheck="false"
					inputmode="numeric"
					aria-label="PIN"
					data-lpignore="true"
					data-1p-ignore
					data-bwignore
					data-form-type="other"
					disabled={!$pinGateVisible}
				/>
				{#if $pinGateError}
					<p id="pin-gate-error" class="pin-gate-error">{$pinGateError}</p>
				{/if}
				<button type="submit" class="btn pin-gate-submit">Unlock</button>
			</form>
		</div>
	</div>
{/if}
