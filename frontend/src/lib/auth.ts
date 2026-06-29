export type PinLockHandler = () => void;

let pinGateMode = false;
let pinGateUnlocked = false;
let onLock: PinLockHandler | null = null;
let fetchPatched = false;

export function getPinGateMode(): boolean {
	return pinGateMode;
}

export function getPinGateUnlocked(): boolean {
	return pinGateUnlocked;
}

export function setPinGateMode(mode: boolean): void {
	pinGateMode = mode;
}

export function setPinGateUnlocked(unlocked: boolean): void {
	pinGateUnlocked = unlocked;
}

export function setOnPinLock(handler: PinLockHandler): void {
	onLock = handler;
}

function isAuthPinFetchArg(input: RequestInfo | URL): boolean {
	const raw = typeof input === 'string' ? input : input instanceof Request ? input.url : String(input);
	try {
		return new URL(raw, location.href).pathname === '/api/auth/pin';
	} catch {
		return false;
	}
}

export function installFetch401Handler(): void {
	if (fetchPatched) return;
	fetchPatched = true;
	const orig = window.fetch.bind(window);
	window.fetch = function (input: RequestInfo | URL, init?: RequestInit) {
		const merged: RequestInit = init ? { ...init } : {};
		if (!merged.credentials) merged.credentials = 'same-origin';
		return orig(input, merged).then((res) => {
			if (res.status === 401 && pinGateMode && !isAuthPinFetchArg(input)) {
				pinGateUnlocked = false;
				onLock?.();
			}
			return res;
		});
	};
}
