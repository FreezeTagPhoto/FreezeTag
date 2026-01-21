import type { Token } from "./tokens";

const DATE_KEYS: Record<
    "takenBefore" | "takenAfter" | "uploadedBefore" | "uploadedAfter",
    true
> = {
    takenBefore: true,
    takenAfter: true,
    uploadedBefore: true,
    uploadedAfter: true,
};

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
                t.exact
                    ? `tag=${encode(t.value)}`
                    : `tagLike=${encode(t.value)}`,
            );
            continue;
        }

        if (t.key === "make" || t.key === "model") {
            const apiKey = keyWithLike(t.key, t.exact);
            parts.push(`${apiKey}=${encode(t.value)}`);
            continue;
        }

        if (t.key in DATE_KEYS) {
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
