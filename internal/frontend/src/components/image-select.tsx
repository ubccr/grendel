import { getV1ImagesOptions } from "@/client/@tanstack/react-query.gen";
import { useQuery } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "./ui/select";

export default function ImageSelect() {
  const images = useQuery(getV1ImagesOptions());
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
