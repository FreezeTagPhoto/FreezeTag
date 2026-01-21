export const FIELD_KEYS = [
    "make",
    "model",
    "takenBefore",
    "takenAfter",
    "uploadedBefore",
    "uploadedAfter",
    "near",
] as const;

export type FieldKey = (typeof FIELD_KEYS)[number];

const FIELD_KEY_SET: ReadonlySet<string> = new Set(FIELD_KEYS);

export function isFieldKey(k: string): k is FieldKey {
    return FIELD_KEY_SET.has(k);
}

export function isDateKey(
    k: FieldKey,
): k is Extract<
    FieldKey,
    "takenBefore" | "takenAfter" | "uploadedBefore" | "uploadedAfter"
> {
    return (
        k === "takenBefore" ||
        k === "takenAfter" ||
        k === "uploadedBefore" ||
        k === "uploadedAfter"
    );
}
