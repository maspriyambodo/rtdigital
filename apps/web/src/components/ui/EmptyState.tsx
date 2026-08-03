import { type ReactNode } from "react";

export type EmptyStateProps = {
  title: ReactNode;
  description?: ReactNode;
  action?: ReactNode;
};

export function EmptyState({ title, description, action }: EmptyStateProps) {
  return (
    <section
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: "var(--space-3)",
        padding: "var(--space-10) var(--space-6)",
        textAlign: "center",
        border: "1px dashed var(--color-border-strong)",
        borderRadius: "var(--radius-lg)",
        background: "var(--color-bg-subtle)",
      }}
    >
      <h3
        style={{
          margin: 0,
          color: "var(--color-text)",
          fontSize: "1.125rem",
          fontWeight: 600,
          letterSpacing: "-0.01em",
        }}
      >
        {title}
      </h3>
      {description ? (
        <p
          style={{
            maxWidth: "32rem",
            margin: 0,
            color: "var(--color-text-secondary)",
            fontSize: "0.9375rem",
          }}
        >
          {description}
        </p>
      ) : null}
      {action ? <div style={{ marginTop: "var(--space-2)" }}>{action}</div> : null}
    </section>
  );
}