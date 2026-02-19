export function formatLocation(lat: number | null, lon: number | null): string {
    if (lat === null || lon === null) return "—";
    return `${lat.toFixed(5)}, ${lon.toFixed(5)}`;
}

export function formatCamera(
    make: string | null,
    model: string | null,
): string {
    const parts = [make, model]
        .map((x) => (x ?? "").trim())
        .filter((x) => x.length > 0);
    return parts.length ? parts.join(" ") : "—";
}

export function formatResultion(
    width: number | null,
    height: number | null,
): string {
    if (height === null || width === null) return "—";
    return `${width} × ${height}`;
}
