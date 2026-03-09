import { createContext, useContext, useState, useCallback } from 'react';
import type { ReactNode } from 'react';
import { setLocale as setLocaleUtil, getLocale, t as translate } from './i18n';

interface I18nContextType {
  locale: string;
  setLocale: (locale: string) => void;
  t: (key: string, params?: Record<string, unknown>) => string;
}

const I18nContext = createContext<I18nContextType | undefined>(undefined);

interface I18nProviderProps {
  children: ReactNode;
}

export function I18nProvider({ children }: I18nProviderProps) {
  const [locale, setLocaleState] = useState(() => {
    // 初始化时从 localStorage 读取
    const savedLocale = localStorage.getItem('locale');
    if (savedLocale === 'zh-CN' || savedLocale === 'en-US') {
      return savedLocale;
    }
    return getLocale();
  });

  const setLocale = useCallback((newLocale: string) => {
    if (newLocale === 'zh-CN' || newLocale === 'en-US') {
      setLocaleState(newLocale);
      setLocaleUtil(newLocale);
    }
  }, []);

  const t = useCallback((key: string, params?: Record<string, unknown>): string => {
    return translate(key, params);
  }, []);

  return (
    <I18nContext.Provider value={{ locale, setLocale, t }}>
      {children}
    </I18nContext.Provider>
  );
}

export function useTranslation() {
  const context = useContext(I18nContext);
  if (context === undefined) {
    throw new Error('useTranslation must be used within an I18nProvider');
  }
  return context;
}
