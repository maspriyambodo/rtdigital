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
    borderColor: "var(--color-success)",
  },
  warning: {
    background: "var(--color-warning-bg)",
    color: "var(--color-warning)",
    borderColor: "var(--color-warning)",
  },
  danger: {
    background: "var(--color-danger-bg)",
    color: "var(--color-danger)",
    borderColor: "var(--color-danger)",
  },
  info: {
    background: "var(--color-info-bg)",
    color: "var(--color-info)",
    borderColor: "var(--color-info)",
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
        padding: "2px var(--space-2)",
        border: "1px solid",
        borderRadius: "var(--radius-full)",
        fontSize: "0.75rem",
        fontWeight: 600,
        lineHeight: 1.4,
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