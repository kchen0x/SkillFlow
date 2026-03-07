import { zh } from './zh'
import { en } from './en'

export type Lang = 'zh' | 'en'
export type TranslationKey = keyof typeof zh
export type Translations = Record<TranslationKey, string>

// Compile-time guard: errors if en and zh key sets diverge
type _EnHasAllZhKeys = { [K in keyof typeof zh]: (typeof en)[K] }
type _ZhHasAllEnKeys = { [K in keyof typeof en]: (typeof zh)[K] }

export const locales: Record<Lang, Translations> = { zh, en }
