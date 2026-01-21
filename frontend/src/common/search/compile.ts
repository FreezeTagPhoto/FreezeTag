import type { Token } from "./tokens";

const DATE_KEYS = new Set([
    "takenBefore",
    "takenAfter",
    "uploadedBefore",
    "uploadedAfter",
] as const);

function encode(v: string) {
    return encodeURIComponent(v);
}

function keyWithLike(base: "make" | "model", exact: boolean) {
    return exact ? base : (`${base}Like` as const);
}

export function compileTokensToApiQuery(tokens: Token[]): string {
    const parts: string[] = [];

    for (const t of tokens) {
        if (t.error) continue;

        if (t.kind === "tag") {
            parts.push(
                t.exact ? `tag=${encode(t.value)}` : `tagLike=${encode(t.value)}`,
            );
            continue;
        }

        if (t.key === "make" || t.key === "model") {
            const apiKey = keyWithLike(t.key, t.exact);
            parts.push(`${apiKey}=${encode(t.value)}`);
            continue;
        }

        if (DATE_KEYS.has(t.key as any)) {
            parts.push(`${t.key}=${encode(t.value)}`);
            continue;
        }

        if (t.key === "near") {
            parts.push(`near=${encode(t.value)}`);
            continue;
        }
    }

    return parts.join("&");
}