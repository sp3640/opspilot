type HealthResponse = {
  status: string;
  service: string;
  environment: string;
  version: string;
  timestamp: string;
};

async function getHealth(): Promise<HealthResponse> {
  const res = await fetch("http://localhost:8080/health", {
    cache: "no-store",
  });

  if (!res.ok) {
    throw new Error("Failed to fetch backend health");
  }

  return res.json();
}

export default async function Home() {
  const health = await getHealth();

  return (
    <main
      style={{
        minHeight: "100vh",
        display: "flex",
        justifyContent: "center",
        alignItems: "center",
        background: "#111827",
        color: "white",
      }}
    >
      <div
        style={{
          width: 500,
          padding: 32,
          borderRadius: 12,
          background: "#1f2937",
        }}
      >
        <h1>🚀 OpsPilot AI</h1>

        <h2>Backend Status</h2>

        <p>🟢 {health.status}</p>

        <hr />

        <p>
          <strong>Service:</strong> {health.service}
        </p>

        <p>
          <strong>Environment:</strong> {health.environment}
        </p>

        <p>
          <strong>Version:</strong> {health.version}
        </p>

        <p>
          <strong>Time:</strong> {health.timestamp}
        </p>
      </div>
    </main>
  );
}