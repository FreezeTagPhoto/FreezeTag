import { splitBySemicolonOutsideQuotes } from "./querysplit";
import type { Token } from "./tokens";
import { isFieldKey, isDateKey } from "./keys";
import { UnitsNearDefault } from "@/common/units/UnitManager";

type StripResult = {
    text: string;
    exact: boolean;
    hadOpenQuote: boolean;
};

function stripOuterQuotes(
    s: string,
    opts?: { trimEnd?: boolean },
): StripResult {
    const trimEnd = opts?.trimEnd ?? true;
    const t0 = s.trimStart();
    const t = trimEnd ? t0.trimEnd() : t0;

    if (!t.startsWith(`"`)) {
        return { text: t, exact: false, hadOpenQuote: false };
    }

    const hadOpenQuote = true;
    const exact = true;
    const text = t.endsWith(`"`) && t.length >= 2 ? t.slice(1, -1) : t.slice(1);

    return { text, exact, hadOpenQuote };
}

function missingClosingQuote(hadOpenQuote: boolean, raw: string): boolean {
    return hadOpenQuote && !raw.trim().endsWith(`"`);
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

const UNIT_ALIASES: Readonly<Record<string, CanonUnit>> = {
    km: "km",
    kms: "km",
    kilometer: "km",
    kilometers: "km",

    m: "m",
    meter: "m",
    meters: "m",

    mi: "mi",
    mile: "mi",
    miles: "mi",

    deg: "deg",
    degree: "deg",
    degrees: "deg",

    ft: "ft",
    foot: "ft",
    feet: "ft",

    yd: "yd",
    yard: "yd",
    yards: "yd",
};

function parseDistanceToAngularDegrees(raw: string): {
    deg: number | null;
    error?: string;
} {
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
    const unitRaw = match[2] ?? UnitsNearDefault(); // unitless defaults to user choice

    if (!Number.isFinite(distNum) || distNum <= 0) {
        return { deg: null, error: `near distance must be a positive number` };
    }

    const unit = UNIT_ALIASES[unitRaw];
    if (!unit) {
        return { deg: null, error: `Unsupported near unit "${unitRaw}"` };
    }

    if (unit === "deg") return { deg: distNum };

    const km =
        unit === "km"
            ? distNum
            : unit === "m"
              ? distNum / 1000
              : unit === "mi"
                ? distNum * 1.609344
                : unit === "ft"
                  ? distNum * 0.0003048 // 1ft = 0.3048m
                  : distNum * 0.0009144; // unit === "yd", 1yd = 0.9144m

    return { deg: kmToAngularDegrees(km) };
}

function parseNear(text: string): {
    normalized: string | null;
    error?: string;
} {
    const fail = (error: string) => ({ normalized: null, error });

    const parts = text
        .split(",")
        .map((p) => p.trim())
        .filter((p) => p.length > 0);

    if (parts.length !== 3) {
        return fail(
            `near expects 3 comma-separated values: lat, lon, distance (e.g. near=40,-110,5km)`,
        );
    }

    const lat = Number(parts[0]);
    const lon = Number(parts[1]);

    if (!Number.isFinite(lat) || !Number.isFinite(lon)) {
        return fail(`near lat/lon must be numbers (e.g. 40,-110,5km)`);
    }
    if (lat < -90 || lat > 90) {
        return fail(`near latitude must be between -90 and 90`);
    }
    if (lon < -180 || lon > 180) {
        return fail(`near longitude must be between -180 and 180`);
    }

    const dist = parseDistanceToAngularDegrees(parts[2]);
    if (dist.deg === null) {
        return fail(dist.error ?? "Invalid near distance");
    }

    return { normalized: `${lat},${lon},${formatDeg(dist.deg)}` };
}

export function parseUserQuery(input: string): Token[] {
    const chunks = splitBySemicolonOutsideQuotes(input);
    const tokens: Token[] = [];

    for (const chunk of chunks) {
        const isClosedBySemicolon =
            chunk.end < input.length && input[chunk.end] === ";";

        const rawNoLeading = chunk.raw.trimStart();
        const trimmed = isClosedBySemicolon
            ? rawNoLeading.trimEnd()
            : rawNoLeading;

        if (trimmed.trim().length === 0) continue;

        const range = computeRangeFromChunk(chunk, trimmed.trimEnd());
        const equalsAt = trimmed.indexOf("=");

        // tag (no '=')
        if (equalsAt === -1) {
            const { text, exact, hadOpenQuote } = stripOuterQuotes(trimmed, {
                trimEnd: isClosedBySemicolon,
            });
            tokens.push({
                kind: "tag",
                valueRaw: trimmed,
                value: text,
                exact,
                range,
                error: missingClosingQuote(hadOpenQuote, trimmed)
                    ? "Missing closing quote"
                    : undefined,
            });
            continue;
        }

        const keyRaw = trimmed.slice(0, equalsAt).trim();
        const valueRaw = trimmed.slice(equalsAt + 1);

        // if unknown key, tag token with error
        if (!isFieldKey(keyRaw)) {
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

        const { text, exact, hadOpenQuote } = stripOuterQuotes(valueRaw, {
            trimEnd: isClosedBySemicolon,
        });
        const quoteError = missingClosingQuote(hadOpenQuote, valueRaw)
            ? "Missing closing quote"
            : undefined;

        // near normalization
        if (keyRaw === "near") {
            const parsed = parseNear(text);
            tokens.push({
                kind: "field",
                key: keyRaw,
                valueRaw,
                value: parsed.normalized ?? text,
                exact,
                range,
                error: quoteError ?? parsed.error,
            });
            continue;
        }

        // date normalization
        if (isDateKey(keyRaw)) {
            const parsed = parseDateOrUnixText(text);
            tokens.push({
                kind: "field",
                key: keyRaw,
                valueRaw,
                value: parsed ?? text,
                exact,
                range,
                error:
                    quoteError ??
                    (parsed === null
                        ? "Invalid date (use YYYY-MM-DD or unix seconds)"
                        : undefined),
            });
            continue;
        }

        // sorting normalization
        if (keyRaw === "sortBy") {
            tokens.push({
                kind: "field",
                key: keyRaw,
                valueRaw,
                value: text,
                exact,
                range,
                error:
                    quoteError ??
                    (text !== "DateAdded" && text !== "DateCreated"
                        ? "Invalid Sorting Strategy"
                        : undefined),
            });
            continue;
        }
        if (keyRaw === "sortOrder") {
            tokens.push({
                kind: "field",
                key: keyRaw,
                valueRaw,
                value: text,
                exact,
                range,
                error:
                    quoteError ??
                    (text !== "ASC" && text !== "DESC"
                        ? "Invalid Sorting Order"
                        : undefined),
            });
            continue;
        }

        // regular field
        tokens.push({
            kind: "field",
            key: keyRaw,
            valueRaw,
            value: text,
            exact,
            range,
            error: quoteError,
        });
    }

    return tokens;
}
