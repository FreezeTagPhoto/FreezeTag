"use client";

import { startTransition, useEffect, useMemo, useRef, useState } from "react";
import styles from "./SearchBar.module.css";
import Pill from "@/components/UI/Pill/Pill";
import { parseUserQuery } from "@/common/search/parse";
import { FIELD_KEYS, isSearchValueKey } from "@/common/search/keys";
import { Search, X } from "lucide-react";

type Props = {
    value: string;
    onChange: (v: string) => void;
    enabled: boolean;
    placeholder?: string;
    allTags?: string[];
};

type Suggestion =
    | { kind: "key"; label: string; insert: string }
    | { kind: "tag"; label: string; insert: string };

function getActiveSegment(input: string, caret: number) {
    const left = input.lastIndexOf(";", Math.max(0, caret - 1));
    const right = input.indexOf(";", caret);

    const segStart = left === -1 ? 0 : left + 1;
    const segEnd = right === -1 ? input.length : right;

    const raw = input.slice(segStart, segEnd);
    const trimmed = raw.trim();

    const leadingSpaces = raw.length - raw.trimStart().length;
    const trimmedStart = segStart + leadingSpaces;

    return { segStart, segEnd, trimmedStart, raw, trimmed };
}

function buildSuggestions(
    input: string,
    caret: number,
    allTags: string[] = [],
): Suggestion[] {
    const { trimmed } = getActiveSegment(input, caret);

    if (!trimmed) return [];

    if (trimmed.includes("=") || trimmed.startsWith(`"`)) return [];

    const needle = trimmed.toLowerCase();

    const keyMatches = FIELD_KEYS.filter(
        (k) => !isSearchValueKey(k) && k.toLowerCase().startsWith(needle),
    ).map((k) => ({
        kind: "key" as const,
        label: `${k}=`,
        insert: `${k}=`,
    }));

    const prefixMatches = allTags.filter((t) =>
        t.toLowerCase().startsWith(needle),
    );
    const substringMatches = allTags.filter(
        (t) =>
            !t.toLowerCase().startsWith(needle) &&
            t.toLowerCase().includes(needle),
    );
    const tagMatches = [...prefixMatches, ...substringMatches]
        .slice(0, Math.max(0, 8 - keyMatches.length))
        .map((t) => ({
            kind: "tag" as const,
            label: t,
            insert: t,
        }));

    // const hasExactMatch = tagMatches.some(
    //     (t) => t.insert.toLowerCase() === needle,
    // );
    // const freeTextSuggestion =
    //     allTags.length === 0 || !hasExactMatch
    //         ? [{ kind: "tag" as const, label: `tag: ${trimmed}`, insert: trimmed }]
    //         : [];

    // return [...keyMatches, ...tagMatches, ...freeTextSuggestion];
    return [...keyMatches, ...tagMatches];
}

function buildTagSuggestionSuffix(after: string): {
    suffix: string;
    caretOffset: number;
} {
    if (after.trim().length === 0) {
        return { suffix: "; ", caretOffset: 2 };
    }

    if (/^\s*;\s*/.test(after)) {
        return {
            suffix: after.replace(/^\s*;\s*/, "; "),
            caretOffset: 2,
        };
    }

    return {
        suffix: `; ${after.replace(/^\s+/, "")}`,
        caretOffset: 2,
    };
}

function removeTokenFromQuery(
    input: string,
    start: number,
    end: number,
): string {
    let before = input.slice(0, start);
    let after = input.slice(end);

    // Prefer removing the trailing semicolon if it exists
    after = after.replace(/^\s*;\s*/, "");

    // Otherwise remove a preceding semicolon if present
    before = before.replace(/;\s*$/, "");

    // Join cleanly
    let out = before.replace(/\s+$/, "");
    if (out.length > 0 && after.trim().length > 0) out += "; ";
    out += after.replace(/^\s+/, "");

    // Final cleanup: preserve "; " when there is still at least one token.
    const cleaned = out.replace(/^\s*;\s*/, "").trim();
    if (!cleaned) return "";
    return `${cleaned.replace(/\s*;\s*$/, "").trim()}; `;
}

export default function SearchBar({
    value,
    onChange,
    enabled,
    placeholder = "Search...",
    allTags = [],
}: Props) {
    const wrapRef = useRef<HTMLDivElement>(null);
    const inputRef = useRef<HTMLInputElement>(null);
    const [draftValue, setDraftValue] = useState(value);

    useEffect(() => {
        setDraftValue(value);
    }, [value]);

    const tokens = useMemo(() => parseUserQuery(draftValue), [draftValue]);

    const [caret, setCaret] = useState(0);

    const [suggestionsEnabled, setSuggestionsEnabled] = useState(true);

    const [dropdownOpen, setDropdownOpen] = useState(false);

    const [manualClosed, setManualClosed] = useState(false);

    const hasValue = draftValue.length > 0;

    const rawSuggestions = useMemo(() => {
        if (!suggestionsEnabled) return [];
        if (manualClosed) return [];
        return buildSuggestions(draftValue, caret, allTags);
    }, [suggestionsEnabled, manualClosed, draftValue, caret, allTags]);

    const suggestions = dropdownOpen ? rawSuggestions : [];

    const maybeOpenDropdown = (nextInput = draftValue, nextCaret = caret) => {
        if (!suggestionsEnabled) return;
        if (manualClosed) return;
        setDropdownOpen(
            buildSuggestions(nextInput, nextCaret, allTags).length > 0,
        );
    };

    const commitValue = (next: string) => {
        setDraftValue(next);
        startTransition(() => {
            onChange(next);
        });
    };

    useEffect(() => {
        const onMouseDown = (e: MouseEvent) => {
            const root = wrapRef.current;
            if (!root) return;
            if (!root.contains(e.target as Node)) {
                setDropdownOpen(false);
                setManualClosed(false);
            }
        };
        document.addEventListener("mousedown", onMouseDown);
        return () => document.removeEventListener("mousedown", onMouseDown);
    }, []);

    useEffect(() => {
        const onKeyDown = (e: KeyboardEvent) => {
            if (e.key !== "Escape") return;
            if (!wrapRef.current?.contains(document.activeElement)) return;

            setDropdownOpen(false);
            setManualClosed(true);
        };
        document.addEventListener("keydown", onKeyDown);
        return () => document.removeEventListener("keydown", onKeyDown);
    }, []);

    const applySuggestion = (s: Suggestion) => {
        const el = inputRef.current;
        if (!el) return;

        const c = el.selectionStart ?? 0;
        const seg = getActiveSegment(draftValue, c);

        const before = draftValue.slice(0, seg.trimmedStart);
        const after = draftValue.slice(seg.segEnd);

        let next = `${before}${s.insert}${after}`;
        let pos = (before + s.insert).length;

        if (s.kind === "tag") {
            const { suffix, caretOffset } = buildTagSuggestionSuffix(after);
            next = `${before}${s.insert}${suffix}`;
            pos = (before + s.insert).length + caretOffset;
        }

        commitValue(next);
        requestAnimationFrame(() => {
            el.focus();
            el.setSelectionRange(pos, pos);
            setCaret(pos);
            setDropdownOpen(false);
            setManualClosed(false);
        });
    };

    const updateCaret = () => {
        const el = inputRef.current;
        if (!el) return 0;
        const nextCaret = el.selectionStart ?? 0;
        setCaret(nextCaret);
        return nextCaret;
    };

    return (
        <div ref={wrapRef} className={styles.wrap}>
            <div className={styles.searchRow}>
                <button
                    className={`${styles.searchIconBtn} ${
                        suggestionsEnabled ? styles.iconOn : styles.iconDisabled
                    }`}
                    disabled={!enabled}
                    type="button"
                    aria-label={
                        suggestionsEnabled
                            ? "Disable suggestions"
                            : "Enable suggestions"
                    }
                    aria-pressed={suggestionsEnabled}
                    onMouseDown={(e) => e.preventDefault()}
                    onClick={() => {
                        setSuggestionsEnabled((v) => !v);
                        setDropdownOpen(false);
                        setManualClosed(false);
                        inputRef.current?.focus();
                    }}
                    title={
                        suggestionsEnabled
                            ? "Suggestions on"
                            : "Suggestions off"
                    }
                >
                    <Search className={styles.btnIcon} aria-hidden="true" />
                </button>

                <input
                    ref={inputRef}
                    className={styles.search}
                    placeholder={placeholder}
                    aria-label="Search"
                    value={draftValue}
                    disabled={!enabled}
                    onChange={(e) => {
                        const next = e.target.value;
                        const nextCaret =
                            e.target.selectionStart ?? next.length;
                        commitValue(next);
                        setCaret(nextCaret);
                        setManualClosed(false);
                        maybeOpenDropdown(next, nextCaret);
                    }}
                    onFocus={() => {
                        const el = inputRef.current;
                        const nextCaret = el?.selectionStart ?? 0;
                        setCaret(nextCaret);
                        setManualClosed(false);
                        maybeOpenDropdown(el?.value ?? draftValue, nextCaret);
                    }}
                    onClick={() => {
                        const el = inputRef.current;
                        const nextCaret = el?.selectionStart ?? 0;
                        setCaret(nextCaret);
                        setManualClosed(false);
                        maybeOpenDropdown(el?.value ?? draftValue, nextCaret);
                    }}
                    onKeyUp={() => {
                        const el = inputRef.current;
                        const nextCaret = updateCaret();
                        maybeOpenDropdown(el?.value ?? draftValue, nextCaret);
                    }}
                />
                {hasValue ? (
                    <button
                        className={`${styles.clear} ${styles.iconOn}`}
                        aria-label="Clear"
                        onClick={() => {
                            commitValue("");
                            setCaret(0);
                            setDropdownOpen(false);
                            setManualClosed(false);
                        }}
                        type="button"
                        disabled={!enabled}
                        title="Clear"
                    >
                        <X className={styles.btnIcon} aria-hidden="true" />
                    </button>
                ) : null}

                {dropdownOpen && suggestions.length > 0 && (
                    <div
                        className={styles.dropdown}
                        role="listbox"
                        aria-label="Search suggestions"
                    >
                        {suggestions.map((s, idx) => (
                            <button
                                key={`${s.kind}-${s.label}-${idx}`}
                                className={styles.dropdownItem}
                                type="button"
                                onMouseDown={(e) => e.preventDefault()}
                                onClick={() => applySuggestion(s)}
                            >
                                <span className={styles.dropdownLabel}>
                                    {s.label}
                                </span>
                                <span className={styles.dropdownMeta}>
                                    {s.kind === "key" ? "filter" : "tag"}
                                </span>
                            </button>
                        ))}
                    </div>
                )}
            </div>

            {/* {tokens.length > 0 ? (
                <div className={styles.tokenRow} aria-label="Parsed filters">
                    {tokens.map((t, idx) => {
                        const label =
                            t.kind === "tag"
                                ? `tag: ${t.value}`
                                : `${t.key}: ${
                                      t.key === "takenBefore" ||
                                      t.key === "takenAfter" ||
                                      t.key === "uploadedBefore" ||
                                      t.key === "uploadedAfter"
                                          ? (t.valueRaw ?? t.value)
                                          : t.value
                                  }`;

                        return (
                            <span
                                key={idx}
                                className={`${styles.tokenWrap} ${
                                    t.error ? styles.tokenError : ""
                                }`}
                                title={t.error ?? ""}
                                onMouseDown={(e) => {
                                    e.preventDefault();
                                    const el = inputRef.current;
                                    if (!el) return;
                                    el.focus();
                                    el.setSelectionRange(
                                        t.range.start,
                                        t.range.end,
                                    );
                                }}
                            >
                                <Pill
                                    label={label}
                                    caret={false}
                                    variant={t.error ? "error" : "token"}
                                />
                                <button
                                    className={styles.tokenClose}
                                    type="button"
                                    aria-label="Remove filter"
                                    onMouseDown={(e) => e.preventDefault()}
                                    onClick={(e) => {
                                        e.stopPropagation();
                                        const next = removeTokenFromQuery(
                                            draftValue,
                                            t.range.start,
                                            t.range.end,
                                        );
                                        commitValue(next);
                                    }}
                                    // title="Remove"
                                >
                                    <X
                                        className={styles.tokenCloseIcon}
                                        aria-hidden="true"
                                    />
                                </button>
                            </span>
                        );
                    })}
                </div>
            ) : (
                <div className={styles.hintRow}>
                    <span>Try:</span>
                    <code>make=&quot;Toyota&quot;;</code>
                    <code>model=Camry;</code>
                </div>
            )} */}
            {tokens.length > 0 ? (
                <div className={styles.tokenRow} aria-label="Parsed filters">
                    {tokens.map((t, idx) => {
                        const label =
                            t.kind === "tag"
                                ? `tag: ${t.value}`
                                : `${t.key}: ${
                                      t.key === "takenBefore" ||
                                      t.key === "takenAfter" ||
                                      t.key === "uploadedBefore" ||
                                      t.key === "uploadedAfter"
                                          ? (t.valueRaw ?? t.value)
                                          : t.value
                                  }`;

                        return (
                            <span
                                key={idx}
                                className={`${styles.tokenWrap} ${
                                    t.error ? styles.tokenError : ""
                                }`}
                                title={t.error ?? ""}
                                onMouseDown={(e) => {
                                    e.preventDefault();
                                    const el = inputRef.current;
                                    if (!el) return;
                                    el.focus();
                                    el.setSelectionRange(
                                        t.range.start,
                                        t.range.end,
                                    );
                                }}
                            >
                                <Pill
                                    label={label}
                                    caret={false}
                                    variant={t.error ? "error" : "token"}
                                />
                                <button
                                    className={styles.tokenClose}
                                    type="button"
                                    aria-label="Remove filter"
                                    onMouseDown={(e) => e.preventDefault()}
                                    onClick={(e) => {
                                        e.stopPropagation();
                                        const next = removeTokenFromQuery(
                                            draftValue,
                                            t.range.start,
                                            t.range.end,
                                        );
                                        commitValue(next);
                                    }}
                                    // title="Remove"
                                >
                                    <X
                                        className={styles.tokenCloseIcon}
                                        aria-hidden="true"
                                    />
                                </button>
                            </span>
                        );
                    })}
                </div>
            ) : null}
        </div>
    );
}
