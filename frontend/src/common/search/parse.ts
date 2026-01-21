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