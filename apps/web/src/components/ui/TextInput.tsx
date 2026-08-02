import { forwardRef, type InputHTMLAttributes } from "react";

export type TextInputProps = InputHTMLAttributes<HTMLInputElement> & {
  invalid?: boolean;
};

export const TextInput = forwardRef<HTMLInputElement, TextInputProps>(
  ({ invalid = false, style, ...props }, ref) => {
    const isInvalid = invalid || props["aria-invalid"] === true;

    return (
      <input
        ref={ref}
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
          boxShadow: isInvalid ? "0 0 0 1px var(--color-danger)" : "var(--shadow-sm)",
          ...style,
        }}
        {...props}
      />
    );
  },
);

TextInput.displayName = "TextInput";