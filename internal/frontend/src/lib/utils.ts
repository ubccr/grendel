import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function severityColor(severity?: string) {
  switch (severity?.toLowerCase()) {
    case "error":
      return "text-red-600 border-red-400";
    case "warning":
      return "text-yellow-600 border-yellow-400";
    case "success":
      return "text-green-600 border-green-400";
    default:
      return "";
  }
}
