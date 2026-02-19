export function formatShortDate(
    ts: number | null,
    opts?: { timeZone?: string },
): string {
    if (ts === null) return "—";
    const d = new Date(ts > 1e12 ? ts : ts * 1000);
    return new Intl.DateTimeFormat(undefined, {
        timeZone: opts?.timeZone,
        year: "numeric",
        month: "short",
        day: "numeric",
    }).format(d);
}

export function formatLongDate(
    ts: number | null,
    opts?: { timeZone?: string },
): string {
    if (ts === null) return "—";
    const d = new Date(ts > 1e12 ? ts : ts * 1000);
    return new Intl.DateTimeFormat(undefined, {
        timeZone: opts?.timeZone,
        year: "numeric",
        month: "short",
        day: "numeric",
        hour: "numeric",
        minute: "2-digit",
    }).format(d);
}
