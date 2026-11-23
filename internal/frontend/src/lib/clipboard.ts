import { toast } from "sonner";

// copyText puts text on the clipboard and says whether it got there. The browser only has a clipboard on a page served over https, so a refusal is toasted rather than thrown
export async function copyText(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text);
    return true;
  } catch {
    toast.error("Could not copy", { description: "The browser only allows copying over https." });
    return false;
  }
}
