import { forwardRef, type TextareaHTMLAttributes } from "react";

export type TextAreaProps = TextareaHTMLAttributes<HTMLTextAreaElement> & {
  invalid?: boolean;
};

export const TextArea = forwardRef<HTMLTextAreaElement, TextAreaProps>(
  ({ invalid = false, style, ...props }, ref) => {
    const isInvalid = invalid || props["aria-invalid"] === true;

    return (
      <textarea
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
          fontSize: "0.9375rem",
          fontFamily: "inherit",
          boxShadow: isInvalid
            ? "0 0 0 1px var(--color-danger)"
            : "var(--shadow-sm)",
          ...style,
        }}
        {...props}
      />
    );
  },
);

TextArea.displayName = "TextArea";