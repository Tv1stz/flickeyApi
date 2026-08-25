export interface ToastItem {
	id: string;
	title?: string;
	description?: string;
	type?: 'default' | 'success' | 'error' | 'warning' | 'info';
	duration?: number;
}

class ToastStore {
	toasts = $state<ToastItem[]>([]);

	show(toast: Omit<ToastItem, 'id'>): string {
		const id = Math.random().toString(36).substring(2, 9);
		const newToast: ToastItem = {
			id,
			duration: 4000,
			...toast
		};

		this.toasts = [...this.toasts, newToast];

		if (newToast.duration && newToast.duration > 0) {
			setTimeout(() => {
				this.dismiss(id);
			}, newToast.duration);
		}

		return id;
	}

	success(description: string, title: string = 'Успешно'): string {
		return this.show({ type: 'success', title, description });
	}

	error(description: string, title: string = 'Ошибка'): string {
		return this.show({ type: 'error', title, description });
	}

	info(description: string, title?: string): string {
		return this.show({ type: 'info', title, description });
	}

	warning(description: string, title: string = 'Внимание'): string {
		return this.show({ type: 'warning', title, description });
	}

	dismiss(id: string): void {
		this.toasts = this.toasts.filter((t) => t.id !== id);
	}
}

export const toast = new ToastStore();
