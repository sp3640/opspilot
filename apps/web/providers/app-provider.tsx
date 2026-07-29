"use client";

import { QueryProvider } from "./query-provider";
import { ThemeProvider } from "./theme-provider";
import { Toaster } from "sonner";
import { useAuthBootstrap } from "@/hooks/use-auth-bootstrap";

function AuthBootstrap({ children }: { children: React.ReactNode }) {
  useAuthBootstrap();
  return children;
}

export function AppProvider({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <ThemeProvider>
      <QueryProvider>
        <AuthBootstrap>
          {children}
          <Toaster richColors position="top-right" />
        </AuthBootstrap>
      </QueryProvider>
    </ThemeProvider>
  );
}