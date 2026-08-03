import { type CSSProperties, type HTMLAttributes } from "react";

export type StatusBadgeProps = HTMLAttributes<HTMLSpanElement> & {
  variant?: "success" | "warning" | "danger" | "info" | "neutral";
};

const styles: Record<
  NonNullable<StatusBadgeProps["variant"]>,
  CSSProperties
> = {
  success: {
    background: "var(--color-success-bg)",
    color: "var(--color-success)",
    borderColor: "rgb(22 101 52 / 0.15)",
  },
  warning: {
    background: "var(--color-warning-bg)",
    color: "var(--color-warning)",
    borderColor: "rgb(154 52 18 / 0.15)",
  },
  danger: {
    background: "var(--color-danger-bg)",
    color: "var(--color-danger)",
    borderColor: "rgb(185 28 28 / 0.15)",
  },
  info: {
    background: "var(--color-info-bg)",
    color: "var(--color-info)",
    borderColor: "rgb(7 89 133 / 0.15)",
  },
  neutral: {
    background: "var(--color-surface-muted)",
    color: "var(--color-text-secondary)",
    borderColor: "var(--color-border-strong)",
  },
};

export function StatusBadge({
  children,
  variant = "neutral",
  style,
  ...props
}: StatusBadgeProps) {
  return (
    <span
      style={{
        display: "inline-flex",
        alignItems: "center",
        padding: "4px var(--space-3)",
        border: "1px solid",
        borderRadius: "var(--radius-full)",
        fontSize: "0.75rem",
        fontWeight: 600,
        letterSpacing: "0.02em",
        lineHeight: 1.2,
        whiteSpace: "nowrap",
        ...styles[variant],
        ...style,
      }}
      {...props}
    >
      {children}
    </span>
  );
}