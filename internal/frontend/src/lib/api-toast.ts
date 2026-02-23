import { toast } from "sonner";
import type { GenericResponse, HttpError } from "@/client";

type props =
  | {
      data: GenericResponse;
      error: undefined;
    }
  | {
      data: undefined;
      error: HttpError;
    };
export default function ApiToast({ data, error }: props) {
  if (data) {
    toast.success(data.title, {
      description: data.detail,
    });
  } else if (error) {
    toast.error(error.title, {
      description: error.detail,
    });
  }
}
