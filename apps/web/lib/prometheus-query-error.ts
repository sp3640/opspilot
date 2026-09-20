import axios from "axios";

// Turns a failed Prometheus query call into a short, human-readable message
// - never the raw backend error string (which can echo upstream Prometheus
// wording) verbatim except for the one genuinely safe case (a plain
// validation message from our own backend, e.g. "query is required").
// Never includes a stack trace or any credential value; the backend itself
// never puts either into an error message, but this function does not rely
// on that - it only ever emits fixed, curated copy or that one safe field.
export function getPrometheusQueryErrorMessage(error: unknown): string {
  if (!axios.isAxiosError<{ message?: string }>(error)) {
    if (error instanceof Error && error.message) return error.message;
    return "Something went wrong running this query. Please try again.";
  }

  const status = error.response?.status;
  const backendMessage = error.response?.data?.message ?? "";

  if (status === 401 || status === 403 || /unauthorized|forbidden|status 401|status 403/i.test(backendMessage)) {
    return "Prometheus rejected the request as unauthorized. Check the integration's credentials.";
  }
  if (/prometheus query error/i.test(backendMessage)) {
    return "Invalid PromQL query. Check the query syntax and try again.";
  }
  if (status === 502 || /prometheus api returned status|call prometheus api|prometheus returned no data/i.test(backendMessage)) {
    return "Prometheus is unavailable right now. Please try again later.";
  }
  if (status === 400) {
    return backendMessage || "This query is invalid. Check the query syntax and try again.";
  }
  if (status === 404) {
    return "This integration could not be found.";
  }
  return "Something went wrong running this query. Please try again.";
}
