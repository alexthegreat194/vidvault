import type { ToastType } from './types';

export function toast(msg: string, type: ToastType = ''): void {
	const el = document.createElement('div');
	el.className = 'toast ' + type;
	el.textContent = msg;
	document.body.appendChild(el);
	setTimeout(() => el.remove(), 3000);
}
