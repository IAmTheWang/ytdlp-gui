import { createContext, ReactNode, useCallback, useContext, useState } from "react";
import { Language, TranslationKey, translations } from "./translations";

export type { Language, TranslationKey } from "./translations";
export { LANGUAGES, isLanguage } from "./translations";

interface I18nContextValue {
    lang: Language;
    setLang: (lang: Language) => void;
    t: (key: TranslationKey, vars?: Record<string, string | number>) => string;
}

const I18nContext = createContext<I18nContextValue | null>(null);

export function I18nProvider({ children }: { children: ReactNode }) {
    const [lang, setLang] = useState<Language>("en");

    const t = useCallback(
        (key: TranslationKey, vars?: Record<string, string | number>) => {
            let str = translations[lang][key] ?? translations.en[key] ?? key;
            if (vars) {
                for (const [name, value] of Object.entries(vars)) {
                    str = str.replaceAll(`{${name}}`, String(value));
                }
            }
            return str;
        },
        [lang]
    );

    return <I18nContext.Provider value={{ lang, setLang, t }}>{children}</I18nContext.Provider>;
}

export function useI18n(): I18nContextValue {
    const ctx = useContext(I18nContext);
    if (!ctx) {
        throw new Error("useI18n must be used within an I18nProvider");
    }
    return ctx;
}
