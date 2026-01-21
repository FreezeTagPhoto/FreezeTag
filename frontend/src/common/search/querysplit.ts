export type QueryChunk = { raw: string; start: number; end: number };

export function splitBySemicolonOutsideQuotes(input: string): QueryChunk[] {
    const chunks: QueryChunk[] = [];
    let start = 0;
    let inQuotes = false;

    for (let i = 0; i < input.length; i++) {
        const ch = input[i];

        const isEscapedQuote = ch === `"` && i > 0 && input[i - 1] === "\\";

        if (ch === `"` && !isEscapedQuote) {
            inQuotes = !inQuotes;
        }

        if (ch === ";" && !inQuotes) {
            chunks.push({ raw: input.slice(start, i), start, end: i });
            start = i + 1;
        }
    }

    chunks.push({ raw: input.slice(start), start, end: input.length });
    return chunks;
}