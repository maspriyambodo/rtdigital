import { forwardRef, type ButtonHTMLAttributes } from "react";

export type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: "primary" | "secondary" | "outline" | "ghost" | "danger";
  loading?: boolean;
};

const styles = {
  primary: {
    background: "var(--color-primary-600)",
    border: "1px solid var(--color-primary-600)",
    color: "#ffffff",
  },
  secondary: {
    background: "var(--color-surface-muted)",
    border: "1px solid var(--color-border)",
    color: "var(--color-text)",
  },
  outline: {
    background: "transparent",
    border: "1px solid var(--color-primary-600)",
    color: "var(--color-primary-600)",
  },
  ghost: {
    background: "transparent",
    border: "1px solid transparent",
    color: "var(--color-primary-600)",
  },
  danger: {
    background: "var(--color-danger)",
    border: "1px solid var(--color-danger)",
    color: "#ffffff",
  },
} as const;

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  (
    { children, variant = "primary", loading = false, disabled, style, ...props },
    ref,
  ) => {
    const isDisabled = loading || disabled;

    return (
      <button
        ref={ref}
        disabled={isDisabled}
        aria-busy={loading || undefined}
        style={{
          display: "inline-flex",
          alignItems: "center",
          justifyContent: "center",
          minHeight: 44,
          padding: "var(--space-2) var(--space-4)",
          borderRadius: "var(--radius-md)",
          cursor: isDisabled ? "not-allowed" : "pointer",
          fontWeight: 600,
          opacity: isDisabled ? 0.6 : 1,
          ...styles[variant],
          ...style,
        }}
        {...props}
      >
        {loading ? "Memuat…" : children}
      </button>
    );
  },
);

Button.displayName = "Button";