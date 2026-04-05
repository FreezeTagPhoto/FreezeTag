"use client";

import {
    startTransition,
    useCallback,
    useEffect,
    useLayoutEffect,
    useMemo,
    useRef,
    useState,
} from "react";
import styles from "./SearchBar.module.css";
import { parseUserQuery } from "@/common/search/parse";
import { FIELD_KEYS, isFieldKey, isSearchValueKey } from "@/common/search/keys";
import { splitBySemicolonOutsideQuotes } from "@/common/search/querysplit";
import type { Token } from "@/common/search/tokens";
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

    after = after.replace(/^\s*;\s*/, "");
    before = before.replace(/;\s*$/, "");

    let out = before.replace(/\s+$/, "");
    if (out.length > 0 && after.trim().length > 0) out += "; ";
    out += after.replace(/^\s+/, "");

    const cleaned = out.replace(/^\s*;\s*/, "").trim();
    if (!cleaned) return "";
    return `${cleaned.replace(/\s*;\s*$/, "").trim()}; `;
}

function tokenToParts(
    token: Token,
):
    | { kind: "tag"; value: string }
    | { kind: "field"; key: string; value: string } {
    if (token.kind === "tag") {
        return { kind: "tag", value: token.value };
    }

    const showRawValue =
        token.key === "takenBefore" ||
        token.key === "takenAfter" ||
        token.key === "uploadedBefore" ||
        token.key === "uploadedAfter";

    return {
        kind: "field",
        key: token.key,
        value: showRawValue ? token.valueRaw.trim() : token.value,
    };
}

function parseActiveField(
    activeRaw: string,
): { key: string; value: string; valueOffset: number } | null {
    const leadingSpaces = activeRaw.length - activeRaw.trimStart().length;
    const withoutLeading = activeRaw.slice(leadingSpaces);
    const equalsAt = withoutLeading.indexOf("=");
    if (equalsAt === -1) return null;

    const keyRaw = withoutLeading.slice(0, equalsAt).trim();
    if (!isFieldKey(keyRaw)) return null;

    const valueOffset = leadingSpaces + equalsAt + 1;
    return {
        key: keyRaw,
        value: activeRaw.slice(valueOffset),
        valueOffset,
    };
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
    const chipsScrollRef = useRef<HTMLDivElement>(null);
    const chipsFadeRafRef = useRef<number | null>(null);
    const [draftValue, setDraftValue] = useState(value);

    useEffect(() => {
        setDraftValue(value);
    }, [value]);

    const tokens = useMemo(() => parseUserQuery(draftValue), [draftValue]);
    const chunks = useMemo(
        () => splitBySemicolonOutsideQuotes(draftValue),
        [draftValue],
    );

    const activeChunk = chunks[chunks.length - 1] ?? {
        raw: "",
        start: 0,
        end: 0,
    };

    const activeStart = activeChunk.start;
    const activeInputValue = draftValue.slice(activeStart);
    const committedPrefix = draftValue.slice(0, activeStart);

    const activeField = useMemo(
        () => parseActiveField(activeInputValue),
        [activeInputValue],
    );

    const activeInputOffset = activeField?.valueOffset ?? 0;
    const activeInputPrefix = activeInputValue.slice(0, activeInputOffset);
    const editableValue = activeField ? activeField.value : activeInputValue;

    const committedTokens = useMemo(
        () => tokens.filter((t) => t.range.start < activeStart),
        [tokens, activeStart],
    );

    const [caret, setCaret] = useState(0);
    const [suggestionsEnabled, setSuggestionsEnabled] = useState(true);
    const [dropdownOpen, setDropdownOpen] = useState(false);
    const [manualClosed, setManualClosed] = useState(false);
    const [chipsFade, setChipsFade] = useState({
        top: false,
        bottom: false,
    });

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

    const syncChipsFade = useCallback(() => {
        if (chipsFadeRafRef.current !== null) {
            cancelAnimationFrame(chipsFadeRafRef.current);
        }

        chipsFadeRafRef.current = requestAnimationFrame(() => {
            chipsFadeRafRef.current = null;
            const el = chipsScrollRef.current;

            if (!el) {
                setChipsFade({ top: false, bottom: false });
                return;
            }

            const overflow = el.scrollHeight > el.clientHeight + 1;
            if (!overflow) {
                setChipsFade({ top: false, bottom: false });
                return;
            }

            const atTop = el.scrollTop <= 1;
            const atBottom =
                el.scrollTop + el.clientHeight >= el.scrollHeight - 1;
            const next = { top: !atTop, bottom: !atBottom };
            setChipsFade((prev) =>
                prev.top === next.top && prev.bottom === next.bottom
                    ? prev
                    : next,
            );
        });
    }, []);

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

    useEffect(() => {
        const el = chipsScrollRef.current;
        if (!el) return;

        syncChipsFade();

        const ro = new ResizeObserver(() => syncChipsFade());
        ro.observe(el);

        const onResize = () => syncChipsFade();
        window.addEventListener("resize", onResize);

        return () => {
            window.removeEventListener("resize", onResize);
            ro.disconnect();
            if (chipsFadeRafRef.current !== null) {
                cancelAnimationFrame(chipsFadeRafRef.current);
                chipsFadeRafRef.current = null;
            }
        };
    }, [syncChipsFade]);

    useEffect(() => {
        syncChipsFade();
    }, [draftValue, syncChipsFade]);

    const applySuggestion = (s: Suggestion) => {
        const el = inputRef.current;
        if (!el) return;

        const c = activeStart + activeInputOffset + (el.selectionStart ?? 0);
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
            const nextChunks = splitBySemicolonOutsideQuotes(next);
            const nextActiveStart =
                nextChunks[nextChunks.length - 1]?.start ?? 0;
            const nextActiveRaw = next.slice(nextActiveStart);
            const nextActiveOffset =
                parseActiveField(nextActiveRaw)?.valueOffset ?? 0;
            const tailPos = Math.max(
                0,
                pos - (nextActiveStart + nextActiveOffset),
            );

            el.focus();
            el.setSelectionRange(tailPos, tailPos);
            setCaret(pos);
            setDropdownOpen(false);
            setManualClosed(false);
        });
    };

    const updateCaret = () => {
        const el = inputRef.current;
        if (!el) return 0;
        const nextCaret =
            activeStart + activeInputOffset + (el.selectionStart ?? 0);
        setCaret(nextCaret);
        return nextCaret;
    };

    useLayoutEffect(() => {
        const el = inputRef.current;
        if (!el) return;
        if (document.activeElement !== el) return;

        const localCaret = Math.max(
            0,
            Math.min(
                editableValue.length,
                caret - (activeStart + activeInputOffset),
            ),
        );

        if (
            el.selectionStart !== localCaret ||
            el.selectionEnd !== localCaret
        ) {
            el.setSelectionRange(localCaret, localCaret);
        }
    }, [caret, activeStart, activeInputOffset, editableValue]);

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

                <div
                    className={`${styles.searchField} ${
                        !enabled ? styles.searchFieldDisabled : ""
                    }`}
                    onMouseDown={() => {
                        if (enabled) inputRef.current?.focus();
                    }}
                >
                    <div
                        className={styles.inputInnerFadeWrap}
                        data-fade-top={chipsFade.top ? "1" : "0"}
                        data-fade-bottom={chipsFade.bottom ? "1" : "0"}
                    >
                        <div
                            ref={chipsScrollRef}
                            className={styles.inputInner}
                            onScroll={syncChipsFade}
                        >
                            {committedTokens.map((token, idx) => {
                                const parts = tokenToParts(token);

                                return (
                                    <span
                                        key={`${token.range.start}-${token.range.end}-${idx}`}
                                        className={`${styles.inlineToken} ${
                                            parts.kind === "tag"
                                                ? styles.inlineTagToken
                                                : ""
                                        } ${
                                            token.error
                                                ? styles.inlineTokenError
                                                : ""
                                        }`}
                                        title={token.error ?? ""}
                                    >
                                        {parts.kind === "field" ? (
                                            <span
                                                className={
                                                    styles.inlineTokenKey
                                                }
                                            >{`${parts.key}:`}</span>
                                        ) : null}

                                        <span
                                            className={styles.inlineTokenValue}
                                        >
                                            {parts.value}
                                        </span>

                                        <button
                                            className={styles.inlineTokenClose}
                                            type="button"
                                            aria-label={
                                                parts.kind === "field"
                                                    ? `Remove ${parts.key} filter`
                                                    : `Remove ${parts.value} tag`
                                            }
                                            disabled={!enabled}
                                            onMouseDown={(e) =>
                                                e.preventDefault()
                                            }
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                const next =
                                                    removeTokenFromQuery(
                                                        draftValue,
                                                        token.range.start,
                                                        token.range.end,
                                                    );
                                                commitValue(next);
                                            }}
                                        >
                                            <X
                                                className={
                                                    styles.tokenCloseIcon
                                                }
                                                aria-hidden="true"
                                            />
                                        </button>
                                    </span>
                                );
                            })}

                            <span
                                className={`${styles.activeInputShell} ${
                                    activeField
                                        ? `${styles.inlineToken} ${styles.inlineTokenActive} ${styles.activeInputShellActive}`
                                        : ""
                                }`}
                            >
                                {activeField ? (
                                    <span className={styles.inlineTokenKey}>
                                        {activeField.key}:
                                    </span>
                                ) : null}

                                <input
                                    ref={inputRef}
                                    className={
                                        activeField
                                            ? styles.inlineTokenInput
                                            : styles.inlineInput
                                    }
                                    style={
                                        activeField
                                            ? {
                                                  width: `${Math.max(
                                                      1,
                                                      editableValue.length + 1,
                                                  )}ch`,
                                              }
                                            : undefined
                                    }
                                    placeholder={
                                        !activeField &&
                                        committedTokens.length === 0
                                            ? placeholder
                                            : undefined
                                    }
                                    aria-label={
                                        activeField
                                            ? `${activeField.key} filter`
                                            : "Search"
                                    }
                                    value={editableValue}
                                    disabled={!enabled}
                                    onChange={(e) => {
                                        const nextEditable = e.target.value;
                                        const nextCaretInEditable =
                                            e.target.selectionStart ??
                                            nextEditable.length;
                                        const nextActive = `${activeInputPrefix}${nextEditable}`;
                                        const next = `${committedPrefix}${nextActive}`;
                                        const nextCaret =
                                            committedPrefix.length +
                                            activeInputOffset +
                                            nextCaretInEditable;

                                        commitValue(next);
                                        setCaret(nextCaret);
                                        setManualClosed(false);
                                        maybeOpenDropdown(next, nextCaret);
                                    }}
                                    onFocus={() => {
                                        const el = inputRef.current;
                                        const nextCaret =
                                            activeStart +
                                            activeInputOffset +
                                            (el?.selectionStart ?? 0);
                                        setCaret(nextCaret);
                                        setManualClosed(false);
                                        maybeOpenDropdown(
                                            draftValue,
                                            nextCaret,
                                        );
                                    }}
                                    onClick={() => {
                                        const el = inputRef.current;
                                        const nextCaret =
                                            activeStart +
                                            activeInputOffset +
                                            (el?.selectionStart ?? 0);
                                        setCaret(nextCaret);
                                        setManualClosed(false);
                                        maybeOpenDropdown(
                                            draftValue,
                                            nextCaret,
                                        );
                                    }}
                                    onKeyDown={(e) => {
                                        if (e.key !== "Backspace") return;

                                        const selStart =
                                            e.currentTarget.selectionStart ?? 0;
                                        const selEnd =
                                            e.currentTarget.selectionEnd ?? 0;
                                        if (selStart !== 0 || selEnd !== 0)
                                            return;

                                        if (activeInputOffset > 0) {
                                            const removeAt =
                                                activeStart +
                                                activeInputOffset -
                                                1;
                                            if (removeAt < 0) return;

                                            const next =
                                                draftValue.slice(0, removeAt) +
                                                draftValue.slice(removeAt + 1);
                                            const nextCaret = removeAt;
                                            commitValue(next);
                                            setCaret(nextCaret);
                                            setManualClosed(false);
                                            maybeOpenDropdown(next, nextCaret);
                                            e.preventDefault();
                                            return;
                                        }

                                        if (
                                            committedTokens.length === 0 ||
                                            editableValue.trim().length > 0
                                        ) {
                                            return;
                                        }

                                        const lastToken =
                                            committedTokens[
                                                committedTokens.length - 1
                                            ];
                                        const next = removeTokenFromQuery(
                                            draftValue,
                                            lastToken.range.start,
                                            lastToken.range.end,
                                        );
                                        commitValue(next);
                                        setDropdownOpen(false);
                                        setManualClosed(false);
                                        e.preventDefault();

                                        requestAnimationFrame(() => {
                                            const nextChunks =
                                                splitBySemicolonOutsideQuotes(
                                                    next,
                                                );
                                            const nextActiveStart =
                                                nextChunks[
                                                    nextChunks.length - 1
                                                ]?.start ?? 0;
                                            setCaret(nextActiveStart);
                                            inputRef.current?.focus();
                                            inputRef.current?.setSelectionRange(
                                                0,
                                                0,
                                            );
                                        });
                                    }}
                                    onKeyUp={() => {
                                        const nextCaret = updateCaret();
                                        maybeOpenDropdown(
                                            draftValue,
                                            nextCaret,
                                        );
                                    }}
                                />
                            </span>
                        </div>
                    </div>
                </div>

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
        </div>
    );
}
