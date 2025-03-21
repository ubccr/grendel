import { useGetV1Images } from "@/openapi/queries";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "./ui/select";
import { LoaderCircle } from "lucide-react";

export default function ImageSelect() {
  const images = useGetV1Images();
  return (
    <Select>
      <SelectTrigger className="w-[180px]">
        <SelectValue
          placeholder={
            images.isFetching ? (
              <>
                <LoaderCircle className="animate-spin" />
                <span className="sr-only">Loading</span>
              </>
            ) : (
              <span>Image</span>
            )
          }
        />
      </SelectTrigger>
      <SelectContent>
        {!images.isFetching &&
          images.data?.map((image, i) => (
            <SelectItem key={i} value={image.name}>
              {image.name}
            </SelectItem>
          ))}
      </SelectContent>
    </Select>
  );
}
