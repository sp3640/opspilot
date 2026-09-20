// NEXT_PUBLIC_* variables are inlined into the client bundle at build time,
// so this module's checks run during `next build` itself (static
// generation evaluates it), not just at runtime.
//
// NEXT_PUBLIC_APP_ENV is a separate, explicit opt-in from Next's own
// NODE_ENV - `next build` always sets NODE_ENV=production internally (even
// for a local/test build), so gating on NODE_ENV would make every build
// require a real API URL and break local development. Requiring a distinct,
// explicitly-set variable (mirroring the backend's own APP_ENV) means only
// a deployment that actually declares itself production is held to this
// check; an ordinary `next build` run without it keeps working exactly as
// before.
const isProductionBuild = process.env.NEXT_PUBLIC_APP_ENV === "production";
const apiUrl = process.env.NEXT_PUBLIC_API_URL;
const isLocalhostUrl = (url: string) => /^https?:\/\/(localhost|127\.0\.0\.1)(:\d+)?(\/|$)/.test(url);

if (isProductionBuild && (!apiUrl || isLocalhostUrl(apiUrl))) {
  throw new Error(
    "NEXT_PUBLIC_API_URL must be set to the real production backend URL when NEXT_PUBLIC_APP_ENV=production " +
      "(it is currently unset or pointing at localhost). See apps/web/.env.example."
  );
}

export const env = {
  API_URL: apiUrl ?? "http://localhost:8080/api/v1",

  APP_NAME: "OpsPilot Enterprise",
} as const;
