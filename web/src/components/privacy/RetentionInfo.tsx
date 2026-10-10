import { useTranslation } from "react-i18next";

const RETENTION = [
  "privacy.retention.meals",
  "privacy.retention.recipes",
  "privacy.retention.sessions",
  "privacy.retention.verifyToken",
  "privacy.retention.resetToken",
  "privacy.retention.export",
  "privacy.retention.audit",
  "privacy.retention.appLog",
  "privacy.retention.erasureLog",
];

export default function RetentionInfo() {
  const { t } = useTranslation();
  return (
    <section aria-labelledby="retention-title" className="space-y-3">
      <h2 id="retention-title" className="text-lg font-semibold">
        {t("privacy.retention.title")}
      </h2>
      <table className="w-full text-sm">
        <thead>
          <tr>
            <th className="text-left font-medium">
              {t("privacy.retention.categoryHeader")}
            </th>
            <th className="text-left font-medium">
              {t("privacy.retention.periodHeader")}
            </th>
          </tr>
        </thead>
        <tbody>
          {RETENTION.map((key) => (
            <tr key={key}>
              <td className="py-1">{t(`${key}.label`)}</td>
              <td className="py-1 opacity-80">{t(`${key}.period`)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}
