"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import styles from "./SearchBar.module.css";
import Pill from "@/components/UI/Pill/Pill";
import { parseUserQuery } from "@/common/search/parse";
import { FIELD_KEYS } from "@/common/search/keys";

type Props = {
    value: string;
    onChange: (v: string) => void;
    placeholder?: string;
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

function buildSuggestions(input: string, caret: number): Suggestion[] {
    const { trimmed } = getActiveSegment(input, caret);

    // Keep the UI clean: no giant list when segment is empty
    if (!trimmed) return [];

    // If they’re typing a value (key=...), don’t suggest keys
    if (trimmed.includes("=") || trimmed.startsWith(`"`)) return [];

    const needle = trimmed.toLowerCase();

    const keyMatches = FIELD_KEYS.filter((k) =>
        k.toLowerCase().startsWith(needle),
    ).map((k) => ({
        kind: "key" as const,
        label: `${k}=`,
        insert: `${k}=`,
    }));

    // Always provide “treat as tag” fallback last
    return [
        ...keyMatches,
        { kind: "tag", label: `tag: ${trimmed}`, insert: trimmed },
    ];
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

    // Final cleanup
    out = out
        .replace(/^\s*;\s*/, "")
        .replace(/\s*;\s*$/, "")
        .trim();

    return out;
}

export default function SearchBar({
    value,
    onChange,
    placeholder = "Search...",
}: Props) {
    const wrapRef = useRef<HTMLDivElement>(null);
    const inputRef = useRef<HTMLInputElement>(null);

    const tokens = useMemo(() => parseUserQuery(value), [value]);

    const [caret, setCaret] = useState(0);

    // Suggestions master toggle (now the left 🔍 button)
    const [suggestionsEnabled, setSuggestionsEnabled] = useState(true);

    // Dropdown open state
    const [dropdownOpen, setDropdownOpen] = useState(false);

    // If user pressed Esc, don’t auto-reopen until they type/click again
    const [manualClosed, setManualClosed] = useState(false);

    const updateCaret = () => {
        const el = inputRef.current;
        if (!el) return;
        setCaret(el.selectionStart ?? 0);
    };

    const suggestions = useMemo(() => {
        if (!suggestionsEnabled) return [];
        if (!dropdownOpen) return [];
        return buildSuggestions(value, caret);
    }, [suggestionsEnabled, dropdownOpen, value, caret]);

    const maybeOpenDropdown = () => {
        if (!suggestionsEnabled) return;
        if (manualClosed) return;

        const next = buildSuggestions(value, caret);
        setDropdownOpen(next.length > 0);
    };

    // Click-outside closes dropdown
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

    // Escape closes dropdown and stays closed
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
        const seg = getActiveSegment(value, c);

        const before = value.slice(0, seg.trimmedStart);
        const after = value.slice(seg.segEnd);

        const next = `${before}${s.insert}${after}`;

        onChange(next);
        requestAnimationFrame(() => {
            el.focus();
            const pos = (before + s.insert).length;
            el.setSelectionRange(pos, pos);
            setCaret(pos);
            setDropdownOpen(false);
            setManualClosed(false);
        });
    };

    return (
        <div ref={wrapRef} className={styles.wrap}>
            <div className={styles.searchRow}>
                {/* Left icon is now the suggestions toggle */}
                <button
                    className={`${styles.searchIconBtn} ${
                        suggestionsEnabled ? styles.iconOn : styles.iconDisabled
                    }`}
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
                    🔍
                </button>

                <input
                    ref={inputRef}
                    className={styles.search}
                    placeholder={placeholder}
                    aria-label="Search"
                    value={value}
                    onChange={(e) => {
                        onChange(e.target.value);
                        setManualClosed(false);
                    }}
                    onFocus={() => {
                        updateCaret();
                        setManualClosed(false);
                        maybeOpenDropdown();
                    }}
                    onClick={() => {
                        updateCaret();
                        setManualClosed(false);
                        maybeOpenDropdown();
                    }}
                    onKeyUp={() => {
                        updateCaret();
                        maybeOpenDropdown();
                    }}
                />

                {/* Clear is always visible; disabled when empty */}
                <button
                    className={`${styles.clear} ${
                        value.length === 0 ? styles.iconDisabled : styles.iconOn
                    }`}
                    aria-label="Clear"
                    onClick={() => onChange("")}
                    type="button"
                    disabled={value.length === 0}
                    title={value.length === 0 ? "Nothing to clear" : "Clear"}
                >
                    ✕
                </button>

                {suggestionsEnabled &&
                    dropdownOpen &&
                    suggestions.length > 0 && (
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
                                        {s.kind === "key" ? "filter" : "text"}
                                    </span>
                                </button>
                            ))}
                        </div>
                    )}
            </div>

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
                                        onChange(
                                            removeTokenFromQuery(
                                                value,
                                                t.range.start,
                                                t.range.end,
                                            ),
                                        );
                                    }}
                                    title="Remove"
                                >
                                    ✕
                                </button>
                            </span>
                        );
                    })}
                </div>
            ) : (
                <div className={styles.hintRow}>
                    Try:{" "}
                    <code>
                        make=&quot;Toyota&quot;; model=Camry;
                        takenAfter=2025-01-01; &quot;beach&quot;
                    </code>
                </div>
            )}
        </div>
    );
}
