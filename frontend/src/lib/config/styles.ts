// src/lib/config/styles.ts
export const cardStyles = {
	// Секция без границ
	section: 'py-8 sm:py-10 md:py-12',

	// Секция с разделителем снизу
	sectionDivided: 'py-8 sm:py-10 md:py-12 border-b border-zinc-200',

	// Заголовок секции
	title: 'text-2xl font-semibold text-zinc-900 md:text-3xl',

	// Подзаголовок
	subtitle: 'text-base text-zinc-500 md:text-lg',

	// Внутренние отступы контента
	contentGap: 'mt-6 sm:mt-8 space-y-6 sm:space-y-8',

	// Legacy — для обратной совместимости
	base: 'sm:rounded-3xl sm:border sm:border-zinc-200 bg-white',
	padding: 'p-4 sm:p-5 md:p-8 lg:p-10',
	card: 'bg-white p-4 sm:p-5 md:p-8 lg:p-10 sm:rounded-3xl sm:border sm:border-zinc-200',
	cardsGap: 'space-y-2 sm:space-y-3 lg:space-y-4'
} as const;

export const formStyles = {
	// Группы полей
	fieldGroup: 'space-y-5 sm:space-y-6',

	// Ряд полей
	fieldRow: 'grid gap-4 md:gap-5',

	// Блок ошибки
	errorBlock: 'flex items-start gap-2.5 rounded-xl bg-red-50 border border-red-100 px-4 py-3',

	// Информационный блок — легче, без границы
	infoBlock: 'rounded-2xl bg-zinc-50 px-4 py-4 flex items-start gap-3',

	// Иконка в блоке
	blockIcon: 'flex h-10 w-10 shrink-0 items-center justify-center rounded-xl'
} as const;

export const mobileStyles = {
	inputText: 'text-base',
	touchTarget: 'min-h-[44px] min-w-[44px]',
	touch: 'touch-manipulation'
} as const;

export const mapOverlayStyles = {
	iconButton:
		'flex h-11 w-11 touch-manipulation items-center justify-center rounded-full border border-zinc-200 bg-white/95 text-zinc-700 shadow-lg backdrop-blur-sm transition-all hover:border-zinc-300 hover:bg-white hover:shadow-xl disabled:cursor-not-allowed disabled:opacity-60',
	infoBubble:
		'max-w-[220px] rounded-xl bg-white/95 px-3 py-2 text-right text-xs font-medium text-zinc-700 shadow-lg backdrop-blur-sm'
} as const;
