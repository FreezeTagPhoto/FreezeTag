export function formatLocation(lat: number | null, lon: number | null): string {
    if (lat === null || lon === null) return "—";
    const latDir = lat >= 0 ? "N" : "S";
    const lonDir = lon >= 0 ? "E" : "W";
    return `${Math.abs(lat).toFixed(5)}°${latDir}, ${Math.abs(lon).toFixed(5)}°${lonDir}`;
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

export function formatResolution(
    width: number | null,
    height: number | null,
): string {
    if (height === null || width === null) return "—";
    return `${width} × ${height}`;
}

export function formatFile(filename: string | null): string {
    if (!filename) return "—";

    const dot = filename.lastIndexOf(".");
    if (dot === -1) return filename;

    return `${filename.slice(dot + 1).toUpperCase()}`;
}
