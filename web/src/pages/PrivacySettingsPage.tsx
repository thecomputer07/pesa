import { useTranslation } from "react-i18next";

import DeleteAccountPanel from "../components/privacy/DeleteAccountPanel";
import EmailChangePanel from "../components/privacy/EmailChangePanel";
import ExportPanel from "../components/privacy/ExportPanel";
import PrivacyConsents from "../components/privacy/PrivacyConsents";
import RetentionInfo from "../components/privacy/RetentionInfo";

export default function PrivacySettingsPage() {
  const { t } = useTranslation();
  return (
    <div className="space-y-8">
      <header>
        <h1 className="text-2xl font-semibold">{t("privacy.pageTitle")}</h1>
        <p className="mt-2 text-sm opacity-80">{t("privacy.pageIntro")}</p>
      </header>

      <PrivacyConsents />
      <RetentionInfo />
      <ExportPanel />
      <EmailChangePanel />
      <DeleteAccountPanel />

      <footer className="text-sm opacity-80 border-t pt-4">
        {t("privacy.contactLine")}{" "}
        <a href="mailto:privacy@pesadia.com">privacy@{"pesadia.com"}</a>
      </footer>
    </div>
  );
}
