export const env = {
  API_URL:
    process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api/v1",

  APP_NAME: "OpsPilot Enterprise",
} as const;