import type { HttpError } from "@/client/types.gen";

// authErrorMessage explains a failed sign-in or sign-up. fetch throws a TypeError when grendeld can't be reached at all; otherwise the API client throws the API's error body, whose status picks the caller's wording for the failures it expects
export function authErrorMessage(error: unknown, byStatus: Partial<Record<number, string>>) {
  if (error instanceof TypeError) {
    return "Grendel didn't respond. Check that grendeld is running, then try again.";
  }
  const e = typeof error === "object" && error !== null ? (error as HttpError) : undefined;
  return (
    (e?.status && byStatus[e.status]) || e?.detail || e?.title || "Something went wrong. Try again."
  );
}
