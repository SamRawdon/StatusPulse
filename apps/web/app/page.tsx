"use client";

import { useEffect, useMemo, useState } from "react";

type Monitor = {
  id: string;
  name: string;
  url: string;
  method: string;
  expectedStatus: number;
  intervalSeconds: number;
  timeoutMs: number;
  enabled: boolean;
  nextRunAt: string;
  createdAt: string;
};

const API_URL = process.env.NEXT_PUBLIC_API_URL!;

export default function Home() {
  const [monitors, setMonitors] = useState<Monitor[]>([]);
  const [name, setName] = useState("");
  const [url, setUrl] = useState("");
  const [intervalSeconds, setIntervalSeconds] = useState(60);
  const [loading, setLoading] = useState(false);
  const [err, setErr] = useState<string | null>(null);

  const canSubmit = useMemo(() => name.trim() !== "" && url.trim() !== "", [name, url]);

  async function refresh() {
    const res = await fetch(`${API_URL}/v1/monitors`, { cache: "no-store" });
    if (!res.ok) throw new Error(await res.text());
    const data = (await res.json()) as Monitor[];
    setMonitors(data);
  }

  useEffect(() => {
    refresh().catch((e) => setErr(String(e?.message ?? e)));
  }, []);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!canSubmit) return;

    setLoading(true);
    setErr(null);

    try {
      const res = await fetch(`${API_URL}/v1/monitors`, {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify({
          name,
          url,
          intervalSeconds,
          timeoutMs: 5000,
          expectedStatus: 200,
          method: "GET",
        }),
      });

      if (!res.ok) throw new Error(await res.text());

      setName("");
      setUrl("");
      setIntervalSeconds(60);
      await refresh();
    } catch (e: any) {
      setErr(String(e?.message ?? e));
    } finally {
      setLoading(false);
    }
  }

  return (
    <main style={{ padding: 24, maxWidth: 900, margin: "0 auto" }}>
      <h1 style={{ marginTop: 0 }}>StatusPulse</h1>

      <section style={{ padding: 16, border: "1px solid #ddd", borderRadius: 12 }}>
        <h2 style={{ marginTop: 0 }}>Create monitor</h2>

        <form onSubmit={onSubmit} style={{ display: "grid", gap: 12 }}>
          <label style={{ display: "grid", gap: 6 }}>
            <span>Name</span>
            <input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="My API"
              style={{ padding: 10, borderRadius: 10, border: "1px solid #ccc" }}
            />
          </label>

          <label style={{ display: "grid", gap: 6 }}>
            <span>URL</span>
            <input
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              placeholder="https://example.com/health"
              style={{ padding: 10, borderRadius: 10, border: "1px solid #ccc" }}
            />
          </label>

          <label style={{ display: "grid", gap: 6 }}>
            <span>Interval (seconds)</span>
            <input
              type="number"
              value={intervalSeconds}
              min={10}
              max={3600}
              onChange={(e) => setIntervalSeconds(Number(e.target.value))}
              style={{ padding: 10, borderRadius: 10, border: "1px solid #ccc" }}
            />
          </label>

          <button
            type="submit"
            disabled={!canSubmit || loading}
            style={{
              padding: 10,
              borderRadius: 10,
              border: "1px solid #111",
              background: loading ? "#eee" : "#111",
              color: loading ? "#111" : "#fff",
              cursor: loading ? "not-allowed" : "pointer",
              width: 180,
            }}
          >
            {loading ? "Creating..." : "Create"}
          </button>

          {err && <pre style={{ margin: 0, color: "crimson", whiteSpace: "pre-wrap" }}>{err}</pre>}
        </form>
      </section>

      <section style={{ marginTop: 24 }}>
        <h2>Monitors</h2>

        {monitors === null || monitors.length === 0 ? (
          <p>No monitors yet.</p>
        ) : (
          <div style={{ display: "grid", gap: 10 }}>
            {monitors.map((m) => (
              <div
                key={m.id}
                style={{
                  padding: 14,
                  border: "1px solid #ddd",
                  borderRadius: 12,
                  display: "grid",
                  gap: 6,
                }}
              >
                <div style={{ display: "flex", justifyContent: "space-between", gap: 12 }}>
                  <strong>{m.name}</strong>
                  <span style={{ opacity: 0.7 }}>{m.intervalSeconds}s</span>
                </div>
                <div style={{ fontFamily: "ui-monospace, SFMono-Regular, Menlo, monospace" }}>{m.url}</div>
                <div style={{ opacity: 0.75, fontSize: 14 }}>
                  expected {m.expectedStatus} · timeout {m.timeoutMs}ms
                </div>
              </div>
            ))}
          </div>
        )}
      </section>
    </main>
  );
}
