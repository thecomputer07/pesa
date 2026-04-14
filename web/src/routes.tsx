import { lazy, Suspense } from "react";
import { Route, Routes } from "react-router-dom";
import Layout from "./components/Layout";
import RequireAuth from "./lib/auth/RequireAuth";

const HomePage = lazy(() => import("./pages/HomePage"));
const SearchPage = lazy(() => import("./pages/SearchPage"));
const FoodDetailPage = lazy(() => import("./pages/FoodDetailPage"));
const ScopePage = lazy(() => import("./pages/ScopePage"));
const AccessibilityPage = lazy(() => import("./pages/AccessibilityPage"));
const AboutPage = lazy(() => import("./pages/AboutPage"));
const DiaryPage = lazy(() => import("./pages/DiaryPage"));
const RecipesPage = lazy(() => import("./pages/RecipesPage"));
const RecipeEditorPage = lazy(() => import("./pages/RecipeEditorPage"));
const FavoritesPage = lazy(() => import("./pages/FavoritesPage"));
const MealTemplatesPage = lazy(() => import("./pages/MealTemplatesPage"));
const ScanPage = lazy(() => import("./pages/ScanPage"));
const SourcesDataPage = lazy(() => import("./pages/SourcesDataPage"));
const SourcesImagesPage = lazy(() => import("./pages/SourcesImagesPage"));
const PrivacySettingsPage = lazy(() => import("./pages/PrivacySettingsPage"));
const LoginPage = lazy(() => import("./pages/LoginPage"));
const RegisterPage = lazy(() => import("./pages/RegisterPage"));
const PasswordResetRequestPage = lazy(() => import("./pages/PasswordResetRequestPage"));
const PasswordResetConfirmPage = lazy(() => import("./pages/PasswordResetConfirmPage"));
const VerifyEmailPage = lazy(() => import("./pages/VerifyEmailPage"));
const CustomFoodPage = lazy(() => import("./pages/CustomFoodPage"));
const MyFoodsPage = lazy(() => import("./pages/MyFoodsPage"));

function PageFallback() {
  return (
    <div role="status" aria-live="polite" className="py-12 text-center text-sm opacity-70">
      …
    </div>
  );
}

export default function AppRoutes() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route
          index
          element={
            <Suspense fallback={<PageFallback />}>
              <HomePage />
            </Suspense>
          }
        />
        <Route
          path="cerca"
          element={
            <Suspense fallback={<PageFallback />}>
              <SearchPage />
            </Suspense>
          }
        />
        <Route
          path="alimento/:id"
          element={
            <Suspense fallback={<PageFallback />}>
              <FoodDetailPage />
            </Suspense>
          }
        />
        <Route
          path="scopo"
          element={
            <Suspense fallback={<PageFallback />}>
              <ScopePage />
            </Suspense>
          }
        />
        <Route
          path="accessibilita"
          element={
            <Suspense fallback={<PageFallback />}>
              <AccessibilityPage />
            </Suspense>
          }
        />
        <Route
          path="chi-siamo"
          element={
            <Suspense fallback={<PageFallback />}>
              <AboutPage />
            </Suspense>
          }
        />
        <Route
          path="diario"
          element={
            <RequireAuth>
              <Suspense fallback={<PageFallback />}>
                <DiaryPage />
              </Suspense>
            </RequireAuth>
          }
        />
        <Route
          path="diario/:date"
          element={
            <RequireAuth>
              <Suspense fallback={<PageFallback />}>
                <DiaryPage />
              </Suspense>
            </RequireAuth>
          }
        />
        <Route
          path="ricette"
          element={
            <RequireAuth>
              <Suspense fallback={<PageFallback />}>
                <RecipesPage />
              </Suspense>
            </RequireAuth>
          }
        />
        <Route
          path="ricette/nuova"
          element={
            <RequireAuth>
              <Suspense fallback={<PageFallback />}>
                <RecipeEditorPage />
              </Suspense>
            </RequireAuth>
          }
        />
        <Route
          path="ricette/:clientId"
          element={
            <RequireAuth>
              <Suspense fallback={<PageFallback />}>
                <RecipeEditorPage />
              </Suspense>
            </RequireAuth>
          }
        />
        <Route
          path="preferiti"
          element={
            <RequireAuth>
              <Suspense fallback={<PageFallback />}>
                <FavoritesPage />
              </Suspense>
            </RequireAuth>
          }
        />
        <Route
          path="template"
          element={
            <RequireAuth>
              <Suspense fallback={<PageFallback />}>
                <MealTemplatesPage />
              </Suspense>
            </RequireAuth>
          }
        />
        <Route
          path="scan"
          element={
            <Suspense fallback={<PageFallback />}>
              <ScanPage />
            </Suspense>
          }
        />
        <Route
          path="fonti-dati"
          element={
            <Suspense fallback={<PageFallback />}>
              <SourcesDataPage />
            </Suspense>
          }
        />
        <Route
          path="fonti-immagini"
          element={
            <Suspense fallback={<PageFallback />}>
              <SourcesImagesPage />
            </Suspense>
          }
        />
        <Route
          path="io/privacy"
          element={
            <RequireAuth>
              <Suspense fallback={<PageFallback />}>
                <PrivacySettingsPage />
              </Suspense>
            </RequireAuth>
          }
        />
        <Route
          path="accedi"
          element={
            <Suspense fallback={<PageFallback />}>
              <LoginPage />
            </Suspense>
          }
        />
        <Route
          path="registrati"
          element={
            <Suspense fallback={<PageFallback />}>
              <RegisterPage />
            </Suspense>
          }
        />
        <Route
          path="password-reset/richiedi"
          element={
            <Suspense fallback={<PageFallback />}>
              <PasswordResetRequestPage />
            </Suspense>
          }
        />
        <Route
          path="password-reset/conferma"
          element={
            <Suspense fallback={<PageFallback />}>
              <PasswordResetConfirmPage />
            </Suspense>
          }
        />
        <Route
          path="verifica-email"
          element={
            <Suspense fallback={<PageFallback />}>
              <VerifyEmailPage />
            </Suspense>
          }
        />
        <Route
          path="alimento/nuovo"
          element={
            <RequireAuth>
              <Suspense fallback={<PageFallback />}>
                <CustomFoodPage />
              </Suspense>
            </RequireAuth>
          }
        />
        <Route
          path="alimento/:id/modifica"
          element={
            <RequireAuth>
              <Suspense fallback={<PageFallback />}>
                <CustomFoodPage />
              </Suspense>
            </RequireAuth>
          }
        />
        <Route
          path="miei-alimenti"
          element={
            <RequireAuth>
              <Suspense fallback={<PageFallback />}>
                <MyFoodsPage />
              </Suspense>
            </RequireAuth>
          }
        />
        <Route path="*" element={<NotFound />} />
      </Route>
    </Routes>
  );
}

function NotFound() {
  return (
    <div className="py-16 text-center">
      <h1 className="text-2xl font-semibold">404</h1>
      <p className="mt-2 text-[color:var(--color-ink-muted)]">
        <a href="/">→ Home</a>
      </p>
    </div>
  );
}
