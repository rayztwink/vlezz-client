import { createI18n } from 'vue-i18n'
import en from './locales/en.json'
import ru from './locales/ru.json'

export type AppLocale = 'en' | 'ru'
export type LanguageSetting = 'system' | AppLocale

export function resolveLocale(language: LanguageSetting = 'system'): AppLocale {
  if (language === 'ru' || language === 'en') {
    return language
  }
  return navigator.language.toLowerCase().startsWith('ru') ? 'ru' : 'en'
}

const i18n = createI18n({
  legacy: false,
  locale: resolveLocale(),
  fallbackLocale: 'en',
  messages: { en, ru }
})

export function setLocale(language: LanguageSetting) {
  i18n.global.locale.value = resolveLocale(language)
}

export default i18n
