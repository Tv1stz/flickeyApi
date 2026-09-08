// src/lib/services/requestQueue.ts
//
// Решает проблему конкуренции между search и reverse за 1 req/s.
// Все запросы к Nominatim идут через единую очередь с соблюдением лимита.

type Task<T> = () => Promise<T>;

interface QueueItem {
	task: Task<unknown>;
	resolve: (v: unknown) => void;
	reject: (e: unknown) => void;
}

class RateLimitedQueue {
	private queue: QueueItem[] = [];
	private running = false;
	private lastCall = 0;

	constructor(private readonly minIntervalMs: number) {}

	enqueue<T>(task: Task<T>): Promise<T> {
		return new Promise<T>((resolve, reject) => {
			const resolveUnknown = (value: unknown) => resolve(value as T);
			this.queue.push({ task: task as Task<unknown>, resolve: resolveUnknown, reject });
			if (!this.running) this.drain();
		});
	}

	// Отменяет все ожидающие задачи (например при unmount компонента)
	cancelPending() {
		const pending = this.queue.splice(0);
		pending.forEach(({ reject }) => reject(new DOMException('Cancelled', 'AbortError')));
	}

	private async drain() {
		this.running = true;

		while (this.queue.length > 0) {
			const item = this.queue.shift()!;
			const wait = this.minIntervalMs - (Date.now() - this.lastCall);
			if (wait > 0) await new Promise((r) => setTimeout(r, wait));

			this.lastCall = Date.now();
			try {
				const result = await item.task();
				item.resolve(result);
			} catch (e) {
				item.reject(e);
			}
		}

		this.running = false;
	}
}

// Единственный экземпляр на всё приложение
export const nominatimQueue = new RateLimitedQueue(1100);
