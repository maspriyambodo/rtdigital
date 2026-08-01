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
        padding: "var(--space-8) var(--space-4)",
        textAlign: "center",
        border: "1px dashed var(--color-border-strong)",
        borderRadius: "var(--radius-lg)",
        background: "var(--color-bg-subtle)",
      }}
    >
      <h3
        style={{
          margin: 0,
          marginBottom: description || action ? "var(--space-2)" : 0,
          color: "var(--color-text)",
          fontSize: "1.125rem",
          fontWeight: 600,
        }}
      >
        {title}
      </h3>
      {description ? (
        <p
          style={{
            maxWidth: "32rem",
            margin: 0,
            marginBottom: action ? "var(--space-6)" : 0,
            color: "var(--color-text-secondary)",
          }}
        >
          {description}
        </p>
      ) : null}
      {action ? <div>{action}</div> : null}
    </section>
  );
}