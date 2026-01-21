import { splitBySemicolonOutsideQuotes } from "./querysplit";
import type { Token } from "./tokens";
import { isFieldKey, isDateKey } from "./keys";

type StripResult = {
    text: string;
    exact: boolean;
    hadOpenQuote: boolean;
};

function stripOuterQuotes(s: string): StripResult {
    const t = s.trim();

    if (t.startsWith(`"`)) {
        const hadOpenQuote = true;

        if (t.endsWith(`"`) && t.length >= 2) {
            return { text: t.slice(1, -1), exact: true, hadOpenQuote };
        }

        return { text: t.slice(1), exact: true, hadOpenQuote };
    }

    return { text: t, exact: false, hadOpenQuote: false };
}


function parseDateOrUnixText(text: string): string | null {
    if (/^[0-9]+$/.test(text)) return text;

    const date = new Date(text);
    if (!Number.isNaN(date.getTime())) {
        return Math.floor(date.getTime() / 1000).toString();
    }

    return null;
}

function computeRangeFromChunk(
    chunk: { raw: string; start: number; end: number },
    trimmed: string,
): { start: number; end: number } {
    const leading = chunk.raw.length - chunk.raw.trimStart().length;
    const start = chunk.start + leading;
    return { start, end: start + trimmed.length };
}

function kmToAngularDegrees(km: number): number {
    const EARTH_RADIUS_KM = 6371;
    return (km / EARTH_RADIUS_KM) * (180 / Math.PI);
}

function formatDeg(deg: number): string {
    return deg.toFixed(6).replace(/\.?0+$/, "");
}

type CanonUnit = "km" | "m" | "mi" | "deg" | "ft" | "yd";

function parseDistanceToAngularDegrees(
    raw: string,
): { deg: number | null; error?: string } {
    const cleaned = raw
        .toLowerCase()
        .replace(/\s+/g, "")
        .replace(/[.)]+$/, "")
        .replace(/°/g, "deg");

    const match = cleaned.match(
        /^(\d+(?:\.\d+)?)(km|kms|kilometer|kilometers|m|meter|meters|mi|mile|miles|deg|degree|degrees|ft|foot|feet|yd|yard|yards)?$/,
    );

    if (!match) {
        return {
            deg: null,
            error: `near distance must look like 5km, 500m, 3mi, 20ft, 10yd, or 0.1°`,
        };
    }

    const distNum = Number(match[1]);
    const unitRaw = match[2] ?? "km"; // if no unit, km default

    if (!Number.isFinite(distNum) || distNum <= 0) {
        return { deg: null, error: `near distance must be a positive number` };
    }

    const unit: CanonUnit | null =
        unitRaw === "km" ||
        unitRaw === "kms" ||
        unitRaw === "kilometer" ||
        unitRaw === "kilometers"
            ? "km"
            : unitRaw === "m" || unitRaw === "meter" || unitRaw === "meters"
              ? "m"
              : unitRaw === "mi" || unitRaw === "mile" || unitRaw === "miles"
                ? "mi"
                : unitRaw === "deg" ||
                    unitRaw === "degree" ||
                    unitRaw === "degrees"
                  ? "deg"
                  : unitRaw === "ft" || unitRaw === "foot" || unitRaw === "feet"
                    ? "ft"
                    : unitRaw === "yd" ||
                        unitRaw === "yard" ||
                        unitRaw === "yards"
                      ? "yd"
                      : null;

    if (unit === null) {
        return { deg: null, error: `Unsupported near unit "${unitRaw}"` };
    }

    if (unit === "deg") {
        return { deg: distNum };
    }

    let km: number;
    if (unit === "km") km = distNum;
    else if (unit === "m") km = distNum / 1000;
    else if (unit === "mi") km = distNum * 1.609344;
    else if (unit === "ft") km = distNum * 0.0003048; // 1ft = 0.3048m
    else if (unit === "yd") km = distNum * 0.0009144; // 1yd = 0.9144m
    else return { deg: null, error: `Unsupported near unit "${unit}"` };

    return { deg: kmToAngularDegrees(km) };
}

function parseNear(text: string): { normalized: string | null; error?: string } {
    const parts = text
        .split(",")
        .map((p) => p.trim())
        .filter((p) => p.length > 0);

    if (parts.length !== 3) {
        return {
            normalized: null,
            error: `near expects 3 comma-separated values: lat, lon, distance (e.g. near=40,-110,5km)`,
        };
    }

    const lat = Number(parts[0]);
    const lon = Number(parts[1]);

    if (!Number.isFinite(lat) || !Number.isFinite(lon)) {
        return {
            normalized: null,
            error: `near lat/lon must be numbers (e.g. 40,-110,5km)`,
        };
    }
    if (lat < -90 || lat > 90) {
        return { normalized: null, error: `near latitude must be between -90 and 90` };
    }
    if (lon < -180 || lon > 180) {
        return { normalized: null, error: `near longitude must be between -180 and 180` };
    }

    const dist = parseDistanceToAngularDegrees(parts[2]);
    if (dist.deg === null) {
        return { normalized: null, error: dist.error };
    }

    const normalized = `${lat},${lon},${formatDeg(dist.deg)}`;
    return { normalized };
}

export function parseUserQuery(input: string): Token[] {
    const chunks = splitBySemicolonOutsideQuotes(input);
    const tokens: Token[] = [];

    for (const chunk of chunks) {
        const trimmed = chunk.raw.trim();
        if (!trimmed) continue;

        const range = computeRangeFromChunk(chunk, trimmed);

        const equalsAt = trimmed.indexOf("=");
        if (equalsAt !== -1) {
            const keyRaw = trimmed.slice(0, equalsAt).trim();
            const valueRaw = trimmed.slice(equalsAt + 1);

            if (isFieldKey(keyRaw)) {
                const { text, exact, hadOpenQuote } = stripOuterQuotes(valueRaw);

                const missingClosingQuote =
                    hadOpenQuote && !valueRaw.trim().endsWith(`"`);

                // normalize
                if (keyRaw === "near") {
                    const parsed = parseNear(text);
                    tokens.push({
                        kind: "field",
                        key: keyRaw,
                        valueRaw,
                        value: parsed.normalized ?? text,
                        exact,
                        range,
                        error: missingClosingQuote
                            ? "Missing closing quote"
                            : parsed.error,
                    });
                    continue;
                }

                if (isDateKey(keyRaw)) {
                    const parsed = parseDateOrUnixText(text);

                    tokens.push({
                        kind: "field",
                        key: keyRaw,
                        valueRaw,
                        value: parsed ?? text,
                        exact,
                        range,
                        error: missingClosingQuote
                            ? "Missing closing quote"
                            : parsed === null
                              ? "Invalid date (use YYYY-MM-DD or unix seconds)"
                              : undefined,
                    });
                } else {
                    tokens.push({
                        kind: "field",
                        key: keyRaw,
                        valueRaw,
                        value: text,
                        exact,
                        range,
                        error: missingClosingQuote ? "Missing closing quote" : undefined,
                    });
                }

                continue;
            }

            // unknown key, treat as tag with error
            tokens.push({
                kind: "tag",
                valueRaw: trimmed,
                value: trimmed,
                exact: false,
                range,
                error: `Unknown filter "${keyRaw}"`,
            });

            continue;
        }

        const { text, exact, hadOpenQuote } = stripOuterQuotes(trimmed);
        const missingClosingQuote = hadOpenQuote && !trimmed.endsWith(`"`);

        tokens.push({
            kind: "tag",
            valueRaw: trimmed,
            value: text,
            exact,
            range,
            error: missingClosingQuote ? "Missing closing quote" : undefined,
        });
    }

    return tokens;
}