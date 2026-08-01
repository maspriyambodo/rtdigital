import { forwardRef, type InputHTMLAttributes } from "react";

export type DatePickerProps = Omit<
  InputHTMLAttributes<HTMLInputElement>,
  "type"
> & {
  invalid?: boolean;
};

export const DatePicker = forwardRef<HTMLInputElement, DatePickerProps>(
  ({ invalid = false, style, ...props }, ref) => {
    const isInvalid = invalid || props["aria-invalid"] === true;

    return (
      <input
        ref={ref}
        type="date"
        style={{
          width: "100%",
          minHeight: 44,
          padding: "var(--space-2) var(--space-3)",
          border: `1px solid ${
            isInvalid ? "var(--color-danger)" : "var(--color-border-strong)"
          }`,
          borderRadius: "var(--radius-md)",
          background: "var(--color-surface)",
          color: "var(--color-text)",
          fontSize: "1rem",
          ...style,
        }}
        {...props}
      />
    );
  },
);

DatePicker.displayName = "DatePicker";