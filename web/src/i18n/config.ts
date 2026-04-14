import i18n from "i18next";
import LanguageDetector from "i18next-browser-languagedetector";
import { initReactI18next } from "react-i18next";
import it from "./locales/it.json";
import en from "./locales/en.json";

if (!i18n.isInitialized) {
  void i18n
    .use(LanguageDetector)
    .use(initReactI18next)
    .init({
      resources: {
        it: { translation: it },
        en: { translation: en },
      },
      fallbackLng: "it",
      supportedLngs: ["it", "en"],
      load: "languageOnly",
      interpolation: { escapeValue: false },
      detection: {
        order: ["localStorage", "navigator", "htmlTag"],
        lookupLocalStorage: "pesa.lang",
        caches: ["localStorage"],
      },
      returnNull: false,
    });
}

export default i18n;
