// if int64 timestamp is huge, treat as ms, otherwise seconds
export function toDate(ts: number): Date {
  return new Date(ts > 1e12 ? ts : ts * 1000);
}

export function formatDate(ts: number | null): string {
  if (ts === null) return "—";
  const d = toDate(ts);
  return new Intl.DateTimeFormat(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  }).format(d);
}

export function formatLocation(lat: number | null, lon: number | null): string {
  if (lat === null || lon === null) return "—";
  return `${lat.toFixed(5)}, ${lon.toFixed(5)}`;
}

export function formatCamera(make: string | null, model: string | null): string {
  const parts = [make, model].filter((x) => x && x.trim().length > 0) as string[];
  return parts.length ? parts.join(" ") : "—";
}