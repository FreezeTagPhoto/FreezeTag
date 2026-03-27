"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { Undo2, MousePointer2 } from "lucide-react";
import styles from "./TopBar.module.css";
import SearchBar from "@/components/SearchBar/SearchBar";
import Pill from "@/components/UI/Pill/Pill";
import { addTagToQuery } from "@/common/search/addtagtoquery";
import { parseUserQuery } from "@/common/search/parse";

type TagInfo = { name: string; count?: number };

type TopBarProps = {
    searchTerm: string;
    onSearchTermChange: React.Dispatch<React.SetStateAction<string>>;

    sortBy: string;
    onSortByChange: (value: string) => void;

    sortOrder: string;
    onSortOrderChange: (value: string) => void;

    multiSelect: boolean;
    onMultiSelectChange: (value: boolean) => void;

    tags: TagInfo[];
};

const SORT_BY_OPTIONS = [
    { value: "DateCreated", label: "Date Created" },
    { value: "DateAdded", label: "Date Added" },
] as const;

const ORDER_OPTIONS = [
    { value: "DESC", label: "Newest first" },
    { value: "ASC", label: "Oldest first" },
] as const;

// Removes a token from the query string given its start and end indices
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

    out = out
        .replace(/^\s*;\s*/, "")
        .replace(/\s*;\s*$/, "")
        .trim();

    return out;
}

// Removes all instances of a tag from the query string
function tryRemoveTagFromQuery(input: string, tag: string): string {
    const tokens = parseUserQuery(input)
        .filter((t) => !t.error && t.kind === "tag" && t.value === tag)
        .sort((a, b) => b.range.start - a.range.start);

    if (tokens.length === 0) return input;

    let out = input;
    for (const tok of tokens) {
        out = removeTokenFromQuery(out, tok.range.start, tok.range.end);
    }

    const trimmed = out.trim();
    if (!trimmed) return "";

    if (trimmed.endsWith(";")) return `${trimmed} `;
    if (trimmed.endsWith("; ")) return trimmed;
    return `${trimmed}; `;
}

export default function TopBar({
    searchTerm,
    onSearchTermChange,
    sortBy,
    onSortByChange,
    sortOrder,
    onSortOrderChange,
    multiSelect,
    onMultiSelectChange,
    tags,
}: TopBarProps) {
    const [open, setOpen] = useState<null | "tags" | "sort">(null);

    const tagsWrapRef = useRef<HTMLDivElement>(null);
    const sortWrapRef = useRef<HTMLDivElement>(null);

    const tagScrollRef = useRef<HTMLDivElement>(null);
    const [tagFade, setTagFade] = useState<{ top: boolean; bottom: boolean }>({
        top: false,
        bottom: false,
    });

    const activeTags = useMemo(() => {
        const set = new Set<string>();
        for (const tok of parseUserQuery(searchTerm)) {
            if (tok.error) continue;
            if (tok.kind === "tag") set.add(tok.value);
        }
        return set;
    }, [searchTerm]);

    const alphaTags = useMemo(() => {
        return (
            [...tags]
                .filter((t) => (t.count ?? 0) > 0) // hides all 0-count tags
                // .filter((t) => (t.count ?? 0) > 0 || activeTags.has(t.name)) // shows active tags even if they are 0-count
                .sort((a, b) => a.name.localeCompare(b.name))
        );
    }, [tags]);

    useEffect(() => {
        const onMouseDown = (e: MouseEvent) => {
            const node = e.target as Node;

            const inTags = tagsWrapRef.current?.contains(node) ?? false;
            const inSort = sortWrapRef.current?.contains(node) ?? false;

            if (!inTags && !inSort) setOpen(null);
            else if (!inTags && open === "tags") setOpen(null);
            else if (!inSort && open === "sort") setOpen(null);
        };

        document.addEventListener("mousedown", onMouseDown);
        return () => document.removeEventListener("mousedown", onMouseDown);
    }, [open]);

    // Close menus on Escape key
    useEffect(() => {
        const onKeyDown = (e: KeyboardEvent) => {
            if (e.key === "Escape") setOpen(null);
        };
        document.addEventListener("keydown", onKeyDown);
        return () => document.removeEventListener("keydown", onKeyDown);
    }, []);

    const syncTagFade = () => {
        const el = tagScrollRef.current;
        if (!el) {
            setTagFade((prev) =>
                prev.top === false && prev.bottom === false
                    ? prev
                    : { top: false, bottom: false },
            );
            return;
        }

        const overflow = el.scrollHeight > el.clientHeight + 1;
        if (!overflow) {
            setTagFade((prev) =>
                prev.top === false && prev.bottom === false
                    ? prev
                    : { top: false, bottom: false },
            );
            return;
        }

        const atTop = el.scrollTop <= 1;
        const atBottom = el.scrollTop + el.clientHeight >= el.scrollHeight - 1;

        const next = { top: !atTop, bottom: !atBottom };
        setTagFade((prev) =>
            prev.top === next.top && prev.bottom === next.bottom ? prev : next,
        );
    };

    useEffect(() => {
        if (open !== "tags") return;

        const el = tagScrollRef.current;
        if (!el) return;

        const raf = requestAnimationFrame(syncTagFade);

        const onResize = () => syncTagFade();
        window.addEventListener("resize", onResize);

        const ro = new ResizeObserver(() => syncTagFade());
        ro.observe(el);

        return () => {
            cancelAnimationFrame(raf);
            window.removeEventListener("resize", onResize);
            ro.disconnect();
        };
    }, [open, alphaTags.length]);

    // Toggle a tag in the search query
    const toggleTag = (tag: string) => {
        onSearchTermChange((prev) => {
            const removed = tryRemoveTagFromQuery(prev, tag);
            return removed === prev ? addTagToQuery(prev, tag) : removed;
        });
    };

    return (
        <div className={styles.bar}>
            <SearchBar
                enabled={!multiSelect}
                value={searchTerm}
                onChange={onSearchTermChange}
                allTags={tags.map((t) => t.name)}
            />

            <div className={styles.pills}>
                <div className={styles.menuContainer} ref={tagsWrapRef}>
                    <Pill
                        label="Tags"
                        caret
                        invertCaret={open === "tags"}
                        variant="menu"
                        onClick={() =>
                            setOpen((v) => (v === "tags" ? null : "tags"))
                        }
                        enabled={!multiSelect}
                    />

                    {open === "tags" && (
                        <div
                            className={`${styles.menuDropdown}`}
                            role="menu"
                            aria-label="Tags"
                        >
                            {alphaTags.length === 0 ? (
                                <div className={styles.menuEmpty}>
                                    <div className={styles.menuEmptyTitle}>
                                        No tags
                                    </div>
                                    <div className={styles.menuEmptySub}>
                                        Nothing to filter.
                                    </div>
                                </div>
                            ) : (
                                <div
                                    className={styles.pillMosaicWrap}
                                    data-fade-top={tagFade.top ? "1" : "0"}
                                    data-fade-bottom={
                                        tagFade.bottom ? "1" : "0"
                                    }
                                >
                                    <div
                                        ref={tagScrollRef}
                                        className={styles.pillMosaic}
                                        role="group"
                                        aria-label="Tag filters"
                                        onScroll={syncTagFade}
                                    >
                                        {alphaTags.map((t) => {
                                            const isActive = activeTags.has(
                                                t.name,
                                            );
                                            return (
                                                <button
                                                    key={t.name}
                                                    type="button"
                                                    className={`${styles.pillChoice} ${
                                                        isActive
                                                            ? styles.pillChoiceActive
                                                            : ""
                                                    }`}
                                                    onMouseDown={(e) =>
                                                        e.preventDefault()
                                                    }
                                                    onClick={() =>
                                                        toggleTag(t.name)
                                                    }
                                                >
                                                    <span
                                                        className={
                                                            styles.pillChoiceLabel
                                                        }
                                                    >
                                                        {t.name}
                                                    </span>
                                                    <span
                                                        className={
                                                            styles.pillChoiceBadge
                                                        }
                                                    >
                                                        {t.count ?? 0}
                                                    </span>
                                                </button>
                                            );
                                        })}
                                    </div>
                                </div>
                            )}
                        </div>
                    )}
                </div>

                <div className={styles.menuContainer} ref={sortWrapRef}>
                    <Pill
                        label="Sort"
                        caret
                        invertCaret={open === "sort"}
                        variant="menu"
                        onClick={() =>
                            setOpen((v) => (v === "sort" ? null : "sort"))
                        }
                        enabled={!multiSelect}
                    />

                    {open === "sort" && (
                        <div
                            className={`${styles.menuDropdown} ${styles.menuDropdownActive}`}
                            role="menu"
                            aria-label="Sort"
                        >
                            <div className={styles.sortGridDropdown}>
                                <div className={styles.sectionTitleWide}>
                                    Sort by
                                </div>

                                {SORT_BY_OPTIONS.map((opt) => {
                                    const on = sortBy === opt.value;
                                    return (
                                        <button
                                            key={opt.value}
                                            type="button"
                                            className={`${styles.sortCell} ${
                                                on ? styles.sortCellActive : ""
                                            }`}
                                            role="menuitemradio"
                                            aria-checked={on}
                                            onMouseDown={(e) =>
                                                e.preventDefault()
                                            }
                                            onClick={() =>
                                                onSortByChange(opt.value)
                                            }
                                        >
                                            {opt.label}
                                        </button>
                                    );
                                })}

                                <div className={styles.sectionTitleWide}>
                                    Order
                                </div>

                                {ORDER_OPTIONS.map((opt) => {
                                    const on = sortOrder === opt.value;
                                    return (
                                        <button
                                            key={opt.value}
                                            type="button"
                                            className={`${styles.sortCell} ${
                                                on ? styles.sortCellActive : ""
                                            }`}
                                            role="menuitemradio"
                                            aria-checked={on}
                                            onMouseDown={(e) =>
                                                e.preventDefault()
                                            }
                                            onClick={() =>
                                                onSortOrderChange(opt.value)
                                            }
                                        >
                                            {opt.label}
                                        </button>
                                    );
                                })}
                            </div>
                        </div>
                    )}
                </div>
                <div className={styles.menuContainer}>
                    <Pill
                        label={
                            multiSelect ? (
                                <>
                                    <Undo2
                                        className={styles.pillIcon}
                                        aria-hidden="true"
                                    />
                                    Return
                                </>
                            ) : (
                                <>
                                    <MousePointer2
                                        className={styles.pillIcon}
                                        aria-hidden="true"
                                    />
                                    Select
                                </>
                            )
                        }
                        variant="menu"
                        onClick={() => onMultiSelectChange(!multiSelect)}
                    />
                </div>
            </div>
        </div>
    );
}
