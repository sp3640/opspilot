'use client';

export function LoadingScreen() {
  return (
    <div
      className="flex min-h-screen items-center justify-center"
      style={{ backgroundColor: 'var(--background)', color: 'var(--foreground)' }}
    >
      <div className="text-center">
        <div
          className="mb-4 inline-flex h-12 w-12 animate-spin rounded-full border-4"
          style={{ borderColor: 'var(--border)', borderTopColor: 'var(--primary)' }}
          role="status"
          aria-label="Loading"
        />
        <p className="text-sm" style={{ color: 'var(--muted-foreground)' }}>
          Loading...
        </p>
      </div>
    </div>
  );
}
