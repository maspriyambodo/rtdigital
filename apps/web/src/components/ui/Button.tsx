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
    boxShadow: "0 1px 2px rgb(37 99 235 / 0.2)",
  },
  secondary: {
    background: "var(--color-surface)",
    border: "1px solid var(--color-border-strong)",
    color: "var(--color-text)",
    boxShadow: "var(--shadow-sm)",
  },
  outline: {
    background: "transparent",
    border: "1px solid var(--color-border-strong)",
    color: "var(--color-primary-700)",
  },
  ghost: {
    background: "transparent",
    border: "1px solid transparent",
    color: "var(--color-text-secondary)",
  },
  danger: {
    background: "var(--color-danger-600)",
    border: "1px solid var(--color-danger-600)",
    color: "#ffffff",
    boxShadow: "0 1px 2px rgb(220 38 38 / 0.2)",
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
          gap: "var(--space-2)",
          minHeight: 44,
          padding: "var(--space-2) var(--space-4)",
          borderRadius: "var(--radius-md)",
          cursor: isDisabled ? "not-allowed" : "pointer",
          fontSize: "0.875rem",
          fontWeight: 600,
          letterSpacing: "0.01em",
          whiteSpace: "nowrap",
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