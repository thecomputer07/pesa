import { Trans, useTranslation } from "react-i18next";

/**
 * Dichiarazione di accessibilità (Legge Stanca / EAA).
 * Template based on AgID self-declaration + EU-2016/2102 structure.
 */
export default function AccessibilityPage() {
  const { t } = useTranslation();
  return (
    <article className="prose max-w-none space-y-4">
      <h1 className="text-3xl font-bold">{t("accessibility.pageTitle")}</h1>

      <p>
        Pesa si impegna a rendere il proprio servizio accessibile, conformemente alla{" "}
        <strong>Legge 9 gennaio 2004, n. 4 (Legge Stanca)</strong> e al suo regolamento di
        attuazione, in recepimento della Direttiva (UE) 2016/2102 e della Direttiva (UE)
        2019/882 (European Accessibility Act).
      </p>

      <h2 className="text-xl font-semibold">Stato di conformità</h2>
      <p>
        Questo sito web è <strong>parzialmente conforme</strong> ai requisiti di cui
        all'allegato A della Legge 4/2004, per via dei casi di non conformità elencati
        di seguito. Lo standard tecnico di riferimento è <strong>WCAG 2.2 livello AA</strong>.
      </p>

      <h2 className="text-xl font-semibold">Contenuti non accessibili</h2>
      <ul>
        <li>
          Le immagini importate da Open Food Facts, quando presenti, possono avere testo
          alternativo sintetico (limite della fonte esterna).
        </li>
        <li>
          Le traduzioni in inglese sono fornite come ausilio: la versione italiana è la
          lingua di riferimento.
        </li>
      </ul>

      <h2 className="text-xl font-semibold">Redazione della dichiarazione</h2>
      <p>
        Dichiarazione redatta il <strong>{"2026-04-13"}</strong>. Il metodo
        utilizzato per redigere la dichiarazione è un'<strong>autovalutazione</strong>{" "}
        effettuata direttamente dal soggetto erogatore, con l'ausilio di <code>axe-core</code>{" "}
        integrato nella pipeline di integrazione continua.
      </p>
      <p>
        Ultima revisione: <strong>{"{{DATA_REVISIONE}}"}</strong>.
      </p>

      <h2 className="text-xl font-semibold">{t("accessibility.contact")}</h2>
      <p>
        <Trans
          i18nKey="accessibility.contactEmail"
          components={{
            a: (
              // eslint-disable-next-line jsx-a11y/anchor-has-content
              <a href={"mailto:accessibilita@pesadia.com"} />
            ),
          }}
          values={{ DOMINIO: "pesadia.com" }}
        />
        : <code>accessibilita@{"pesadia.com"}</code>
      </p>
      <p>
        In caso di risposta insoddisfacente o di mancata risposta entro{" "}
        <strong>30 giorni</strong>, l'utente può rivolgersi al difensore civico digitale:{" "}
        <a href="https://difensorecivico.agid.gov.it/" target="_blank" rel="noopener noreferrer">
          AgID — Difensore civico per il digitale
        </a>
        .
      </p>

      <h2 className="text-xl font-semibold">Informazioni sul sito</h2>
      <ul>
        <li>
          <strong>Data di pubblicazione del sito:</strong> {"2026-04-13"}
        </li>
        <li>
          <strong>Tecnologie assistive testate:</strong> NVDA 2024 + Firefox 128, VoiceOver
          macOS 14 + Safari 17, TalkBack Android 14 + Chrome 126.
        </li>
        <li>
          <strong>Strumenti di valutazione:</strong> axe-core 4.x, tastiera fisica, zoom
          400%, preferenza "reduce motion".
        </li>
      </ul>
    </article>
  );
}
