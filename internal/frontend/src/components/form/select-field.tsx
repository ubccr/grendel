import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Field, FieldContent, FieldDescription, FieldError, FieldLabel } from "../ui/field";
import { useFieldContext } from "./form-context";

type props = {
  label: string;
  placeholder?: string;
  description?: string;
  items: Map<string, string>;
};

export function SelectField({ label, placeholder, description, items }: props) {
  const field = useFieldContext<string>();
  const itemArray = Array.from(items.entries());
  const isInvalid = field.state.meta.isTouched && !field.state.meta.isValid;

  return (
    <Field data-invalid={isInvalid} orientation="responsive">
      <FieldContent>
        <FieldLabel htmlFor={field.name}>{label}</FieldLabel>
        {description && <FieldDescription>{description}</FieldDescription>}
        {isInvalid && <FieldError errors={field.state.meta.errors} />}
      </FieldContent>
      <Select name={field.name} value={field.state.value} onValueChange={field.handleChange}>
        <SelectTrigger aria-invalid={isInvalid}>
          <SelectValue placeholder={placeholder} />
        </SelectTrigger>
        <SelectContent>
          <SelectGroup>
            {itemArray.map(([k, v]) => (
              <SelectItem key={k} value={k}>
                {v}
              </SelectItem>
            ))}
          </SelectGroup>
        </SelectContent>
      </Select>
    </Field>
  );
}
