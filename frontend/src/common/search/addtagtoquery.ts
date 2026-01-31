import { parseUserQuery } from "@/common/search/parse";

function formatTagToken(tag: string): string {
    const t = tag.trim();
    if (!t) return "";

    const safe = t.replaceAll(`"`, "");

    return `"${safe}"`;
}

export function addTagToQuery(input: string, tag: string): string {
    const t = tag.trim();
    if (!t) return input;

    const tokens = parseUserQuery(input);
    const alreadyHas = tokens.some(
        (tok) => tok.kind === "tag" && tok.value.trim() === t,
    );
    if (alreadyHas) {
        const trimmed = input.trimEnd();
        return trimmed.endsWith(";") ? `${trimmed} ` : `${trimmed}; `;
    }

    const tokenText = formatTagToken(t);

    const cleaned = input
        .trim()
        .replace(/\s*;\s*$/, "")
        .trim();

    if (cleaned.length === 0) return `${tokenText}; `;

    return `${cleaned}; ${tokenText}; `;
}
