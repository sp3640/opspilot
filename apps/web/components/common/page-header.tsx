import type { ReactNode } from "react";
import Link from "next/link";
import { ArrowLeft, ChevronRight } from "lucide-react";

import { BackLink } from "./back-link";

type BreadcrumbItem = {
  label: string;
  href?: string;
};

type PageHeaderProps = {
  title: string;
  description?: string;
  actions?: ReactNode;
  breadcrumb?: BreadcrumbItem[];
  /**
   * Shared "Back" affordance for pages reached by drilling into a detail/
   * deep workflow from a list (application/deployment/incident/alert/
   * integration details, settings subpages) - where relying on the
   * browser's own back button alone is not a reliable or sufficiently
   * discoverable way back to the parent list. Opt-in: pass `backHref` for
   * a plain link back to a known parent route, or `onBack` for custom
   * behavior (e.g. `router.back()`). Omit both on pages reached directly
   * from primary navigation, where a Back control would be redundant.
   */
  backHref?: string;
  onBack?: () => void;
  backLabel?: string;
};

/** A shared page-level title block for all product modules. */
export function PageHeader({ title, description, actions, breadcrumb, backHref, onBack, backLabel = "Back" }: PageHeaderProps) {
  return (
    <header className="flex flex-col gap-5 sm:flex-row sm:items-end sm:justify-between">
      <div className="min-w-0">
        {(backHref || onBack) && (
          <div className="mb-3">
            {onBack ? (
              <button
                type="button"
                onClick={onBack}
                className="inline-flex items-center gap-1.5 text-sm font-medium transition-colors hover:text-[var(--foreground)]"
                style={{ color: "var(--muted-foreground)" }}
              >
                <ArrowLeft aria-hidden="true" className="h-4 w-4" />
                {backLabel}
              </button>
            ) : (
              <BackLink fallbackHref={backHref as string} label={backLabel} />
            )}
          </div>
        )}
        {breadcrumb && breadcrumb.length > 0 && (
          <nav aria-label="Breadcrumb" className="mb-3">
            <ol className="flex flex-wrap items-center gap-1.5 text-sm" style={{ color: "var(--muted-foreground)" }}>
              {breadcrumb.map((item, index) => (
                <li key={`${item.label}-${index}`} className="flex items-center gap-1.5">
                  {index > 0 && <ChevronRight aria-hidden="true" className="h-3.5 w-3.5" />}
                  {item.href ? <Link href={item.href} className="rounded-sm transition-colors hover:text-[var(--foreground)]">{item.label}</Link> : <span aria-current="page">{item.label}</span>}
                </li>
              ))}
            </ol>
          </nav>
        )}
        <h1 className="text-3xl font-semibold tracking-tight sm:text-4xl">{title}</h1>
        {description && <p className="mt-2 max-w-2xl text-base leading-7" style={{ color: "var(--muted-foreground)" }}>{description}</p>}
      </div>
      {actions && <div className="flex shrink-0 flex-wrap items-center gap-3">{actions}</div>}
    </header>
  );
}
