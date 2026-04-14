import { useTranslation } from "react-i18next";

// Static consent info. Values pulled from Privacy Policy §3 and cookie policy.
export default function PrivacyConsents() {
  const { t } = useTranslation();
  return (
    <section aria-labelledby="privacy-consents-title" className="space-y-3">
      <h2 id="privacy-consents-title" className="text-lg font-semibold">
        {t("privacy.consents.title")}
      </h2>
      <dl className="grid grid-cols-1 gap-2 text-sm">
        <div>
          <dt className="font-medium">{t("privacy.consents.policyLabel")}</dt>
          <dd className="opacity-80">{t("privacy.consents.policyValue")}</dd>
        </div>
        <div>
          <dt className="font-medium">{t("privacy.consents.cookiesLabel")}</dt>
          <dd className="opacity-80">{t("privacy.consents.cookiesValue")}</dd>
        </div>
        <div>
          <dt className="font-medium">{t("privacy.consents.analyticsLabel")}</dt>
          <dd className="opacity-80">{t("privacy.consents.analyticsValue")}</dd>
        </div>
      </dl>
    </section>
  );
}
