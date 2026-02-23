import { Field, FieldContent, FieldDescription, FieldError, FieldLabel } from "../ui/field";
import { Input } from "../ui/input";
import { useFieldContext } from "./form-context";

type props = {
  label: string;
  placeholder?: string;
  description?: string;
  autoComplete?: React.HTMLInputAutoCompleteAttribute;
};

export function TextField({ label, placeholder, description, autoComplete }: props) {
  const field = useFieldContext<string>();
  const isInvalid = field.state.meta.isTouched && !field.state.meta.isValid;

  return (
    <Field data-invalid={isInvalid} orientation="responsive">
      <FieldContent>
        <FieldLabel htmlFor={field.name}>{label}</FieldLabel>
        {description && <FieldDescription>{description}</FieldDescription>}
        {isInvalid && <FieldError errors={field.state.meta.errors} />}
      </FieldContent>
      <Input
        name={field.name}
        value={field.state.value}
        onChange={(e) => field.handleChange(e.target.value)}
        onBlur={field.handleBlur}
        aria-invalid={isInvalid}
        placeholder={placeholder}
        autoComplete={autoComplete}
      />
    </Field>
  );
}
