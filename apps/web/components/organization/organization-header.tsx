import { Building2 } from "lucide-react";

import { PageHeader } from "@/components/common";

export function OrganizationHeader() {
  return (
    <PageHeader
      title="Organization"
      description="Manage your organization identity and workspace details."
      breadcrumb={[{ label: "Overview", href: "/" }, { label: "Organization" }]}
      actions={<Building2 aria-hidden={true} className="h-5 w-5" />}
    />
  );
}
