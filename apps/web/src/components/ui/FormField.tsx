import { useId, type ReactNode } from "react";

export type FormFieldProps = {
  id?: string;
  label: ReactNode;
  hint?: ReactNode;
  error?: ReactNode;
  required?: boolean;
  children: (props: {
    id: string;
    "aria-describedby"?: string;
    "aria-invalid"?: boolean;
  }) => ReactNode;
};

export function FormField({
  id: explicitId,
  label,
  hint,
  error,
  required = false,
  children,
}: FormFieldProps) {
  const generatedId = useId();
  const id = explicitId ?? generatedId;
  const hintId = `${id}-hint`;
  const errorId = `${id}-error`;
  const describedBy = [
    hint ? hintId : undefined,
    error ? errorId : undefined,
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        gap: "var(--space-2)",
      }}
    >
      <label htmlFor={id} style={{ color: "var(--color-text)", fontWeight: 600 }}>
        {label}
        {required ? (
          <span
            aria-hidden="true"
            style={{ marginLeft: "var(--space-1)", color: "var(--color-danger)" }}
          >
            *
          </span>
        ) : null}
      </label>

      {hint ? (
        <p
          id={hintId}
          style={{ color: "var(--color-text-secondary)", fontSize: "0.875rem" }}
        >
          {hint}
        </p>
      ) : null}

      {children({
        id,
        "aria-describedby": describedBy || undefined,
        "aria-invalid": error ? true : undefined,
      })}

      {error ? (
        <p
          id={errorId}
          role="alert"
          style={{
            color: "var(--color-danger)",
            fontSize: "0.875rem",
            fontWeight: 500,
          }}
        >
          {error}
        </p>
      ) : null}
    </div>
  );
}