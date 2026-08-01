import { forwardRef, type SelectHTMLAttributes } from "react";

export type SelectProps = SelectHTMLAttributes<HTMLSelectElement> & {
  invalid?: boolean;
};

export const Select = forwardRef<HTMLSelectElement, SelectProps>(
  ({ invalid = false, style, children, ...props }, ref) => {
    const isInvalid = invalid || props["aria-invalid"] === true;

    return (
      <div style={{ position: "relative" }}>
        <select
          ref={ref}
          style={{
            appearance: "none",
            width: "100%",
            minHeight: 44,
            padding: "var(--space-2) 2.5rem var(--space-2) var(--space-3)",
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
        >
          {children}
        </select>
        <div
          aria-hidden="true"
          style={{
            position: "absolute",
            top: 0,
            right: 0,
            bottom: 0,
            display: "flex",
            alignItems: "center",
            paddingRight: "var(--space-3)",
            pointerEvents: "none",
            color: "var(--color-text-muted)",
          }}
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <polyline points="6 9 12 15 18 9" />
          </svg>
        </div>
      </div>
    );
  },
);

Select.displayName = "Select";