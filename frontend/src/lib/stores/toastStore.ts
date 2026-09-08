// src/lib/stores/toastStore.ts
import { writable } from 'svelte/store';

export type ToastType = 'success' | 'error' | 'warning' | 'info';

export interface Toast {
	id: string;
	type: ToastType;
	title: string;
	message?: string;
	duration: number;
	dismissible: boolean;
	createdAt: number;
}

export interface ConfirmToast {
	id: string;
	title: string;
	message: string;
	confirmText: string;
	cancelText: string;
	onConfirm: () => void;
	onCancel: () => void;
	type: 'danger' | 'warning' | 'info';
}

const DEFAULT_DURATIONS: Record<ToastType, number> = {
	success: 4000,
	info: 5000,
	warning: 6000,
	error: 8000
};

function createToastStore() {
	const toasts = writable<Toast[]>([]);
	const confirmToast = writable<ConfirmToast | null>(null);

	function generateId(): string {
		return Math.random().toString(36).slice(2, 9);
	}

	function add(options: {
		type: ToastType;
		title: string;
		message?: string;
		duration?: number;
		dismissible?: boolean;
	}): string {
		const id = generateId();

		const toast: Toast = {
			id,
			type: options.type,
			title: options.title,
			message: options.message,
			duration: options.duration ?? DEFAULT_DURATIONS[options.type],
			dismissible: options.dismissible ?? true,
			createdAt: Date.now()
		};

		toasts.update((all) => [...all, toast]);

		// НЕ устанавливаем таймер здесь - управление таймерами в ToastContainer

		return id;
	}

	function dismiss(id: string) {
		toasts.update((all) => all.filter((t) => t.id !== id));
	}

	function dismissAll() {
		toasts.set([]);
	}

	function success(title: string, message?: string, duration?: number) {
		return add({ type: 'success', title, message, duration });
	}

	function error(title: string, message?: string, duration?: number) {
		return add({ type: 'error', title, message, duration });
	}

	function warning(title: string, message?: string, duration?: number) {
		return add({ type: 'warning', title, message, duration });
	}

	function info(title: string, message?: string, duration?: number) {
		return add({ type: 'info', title, message, duration });
	}

	function confirm(options: {
		title: string;
		message: string;
		confirmText?: string;
		cancelText?: string;
		type?: 'danger' | 'warning' | 'info';
		onConfirm?: () => void;
		onCancel?: () => void;
	}): Promise<boolean> {
		return new Promise((resolve) => {
			const id = generateId();

			confirmToast.set({
				id,
				title: options.title,
				message: options.message,
				confirmText: options.confirmText ?? 'Подтвердить',
				cancelText: options.cancelText ?? 'Отмена',
				type: options.type ?? 'warning',
				onConfirm: () => {
					options.onConfirm?.();
					confirmToast.set(null);
					resolve(true);
				},
				onCancel: () => {
					options.onCancel?.();
					confirmToast.set(null);
					resolve(false);
				}
			});
		});
	}

	function dismissConfirm() {
		confirmToast.set(null);
	}

	return {
		subscribe: toasts.subscribe,
		confirmToast: { subscribe: confirmToast.subscribe },
		add,
		dismiss,
		dismissAll,
		success,
		error,
		warning,
		info,
		confirm,
		dismissConfirm
	};
}

export const toast = createToastStore();
