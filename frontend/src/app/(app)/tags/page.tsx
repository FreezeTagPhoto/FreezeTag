"use client";

import { useEffect, useMemo, useState } from "react";
import styles from "./page.module.css";
import TagGetter from "@/api/tags/taggetter";
import TagCounter from "@/api/tags/tagcounter";
import TagDeleter from "@/api/tags/tagdeleter";
import { useRouter } from "next/navigation";
import {
    AlertTriangle,
    CheckSquare,
    RefreshCcw,
    Search,
    Square,
    Trash2,
    X,
    ArrowRight,
} from "lucide-react";

type Banner =
    | { kind: "success"; text: string }
    | { kind: "error"; text: string }
    | null;

function tagToGalleryQuery(tag: string): string {
    const escaped = tag.replace(/"/g, '\\"');
    return `"${escaped}";`;
}

export default function TagsPage() {
    const router = useRouter();

    const [tags, setTags] = useState<string[]>([]);
    const [counts, setCounts] = useState<Record<string, number>>({});
    const [countsOk, setCountsOk] = useState<boolean>(false);

    const [loading, setLoading] = useState<boolean>(true);
    const [refreshing, setRefreshing] = useState<boolean>(false);
    const [deleting, setDeleting] = useState<boolean>(false);

    const [filter, setFilter] = useState<string>("");
    const [selected, setSelected] = useState<Set<string>>(new Set());
    const [banner, setBanner] = useState<Banner>(null);

    const [confirmOpen, setConfirmOpen] = useState<boolean>(false);
    const [confirmTags, setConfirmTags] = useState<string[]>([]);

    const normalizedFilter = filter.trim().toLowerCase();

    const filteredTags = useMemo(() => {
        const base = [...tags].sort((a, b) => a.localeCompare(b));
        if (!normalizedFilter) return base;
        return base.filter((t) => t.toLowerCase().includes(normalizedFilter));
    }, [tags, normalizedFilter]);

    const selectedCount = selected.size;

    const filteredSelectedCount = useMemo(() => {
        if (selectedCount === 0) return 0;
        let count = 0;
        for (const t of filteredTags) if (selected.has(t)) count++;
        return count;
    }, [filteredTags, selected, selectedCount]);

    const allFilteredSelected =
        filteredTags.length > 0 && filteredSelectedCount === filteredTags.length;

    async function loadTags(opts?: { silent?: boolean }) {
        const silent = opts?.silent ?? false;
        if (!silent) setLoading(true);
        else setRefreshing(true);

        setBanner(null);

        const [listRes, countRes] = await Promise.all([
            TagGetter(),
            TagCounter(""),
        ]);

        const nextTagsSet = new Set<string>();

        if (listRes.ok) {
            for (const t of listRes.value) {
                if (typeof t === "string" && t.length > 0) nextTagsSet.add(t);
            }
        } else {
            setBanner({
                kind: "error",
                text: `Failed to load tag list (${listRes.error.status}): ${listRes.error.message}`,
            });
        }

        if (countRes.ok) {
            const map = countRes.value ?? {};
            const safe: Record<string, number> = {};
            for (const [k, v] of Object.entries(map)) {
                safe[k] = typeof v === "number" ? v : 0;
                if (k && typeof k === "string") nextTagsSet.add(k);
            }
            setCounts(safe);
            setCountsOk(true);
        } else {
            // Keep list usable even if counts fail.
            setCounts({});
            setCountsOk(false);
            setBanner((prev) => {
                const msg = `Failed to load tag counts (${countRes.error.status}): ${countRes.error.message}`;
                if (!prev) return { kind: "error", text: msg };
                // avoid duplicating errors; keep the first if it exists
                return prev;
            });
        }

        const nextTags = [...nextTagsSet].sort((a, b) => a.localeCompare(b));

        setTags(nextTags);
        setSelected((prev) => {
            const next = new Set<string>();
            const existing = new Set(nextTags);
            for (const t of prev) if (existing.has(t)) next.add(t);
            return next;
        });

        if (!silent) setLoading(false);
        else setRefreshing(false);
    }

    useEffect(() => {
        void loadTags();
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    useEffect(() => {
        if (!confirmOpen) return;

        const onKeyDown = (e: KeyboardEvent) => {
            if (e.key === "Escape") {
                setConfirmOpen(false);
                setConfirmTags([]);
            }
        };

        window.addEventListener("keydown", onKeyDown);
        return () => window.removeEventListener("keydown", onKeyDown);
    }, [confirmOpen]);

    function toggleSelected(tag: string) {
        setSelected((prev) => {
            const next = new Set(prev);
            if (next.has(tag)) next.delete(tag);
            else next.add(tag);
            return next;
        });
    }

    function toggleSelectAllFiltered() {
        setSelected((prev) => {
            const next = new Set(prev);
            if (allFilteredSelected) {
                for (const t of filteredTags) next.delete(t);
            } else {
                for (const t of filteredTags) next.add(t);
            }
            return next;
        });
    }

    function requestDelete(tagsToDelete: string[]) {
        if (tagsToDelete.length === 0) return;
        setBanner(null);
        setConfirmTags(tagsToDelete);
        setConfirmOpen(true);
    }

    async function performDelete() {
        if (confirmTags.length === 0) return;

        setDeleting(true);
        setBanner(null);

        const toDelete = [...confirmTags];

        const result = await TagDeleter(toDelete);
        if (!result.ok) {
            setBanner({
                kind: "error",
                text: `Failed to delete tag(s) (${result.error.status}): ${result.error.message}`,
            });
            setDeleting(false);
            return;
        }

        setTags((prev) => prev.filter((t) => !toDelete.includes(t)));
        setSelected((prev) => {
            const next = new Set(prev);
            for (const t of toDelete) next.delete(t);
            return next;
        });
        setCounts((prev) => {
            const next = { ...prev };
            for (const t of toDelete) delete next[t];
            return next;
        });

        setBanner({
            kind: "success",
            text: `Deleted ${result.value} tag record(s).`,
        });

        setDeleting(false);
        setConfirmOpen(false);
        setConfirmTags([]);
    }

    function goToGallery(tag: string) {
        const q = tagToGalleryQuery(tag);
        router.push(`/?q=${encodeURIComponent(q)}`);
    }

    const totalUsed = useMemo(() => {
        if (!countsOk) return null;
        let sum = 0;
        for (const t of tags) sum += counts[t] ?? 0;
        return sum;
    }, [countsOk, counts, tags]);

    return (
        <main className={styles.main}>
            <header className={styles.header}>
                <div className={styles.headerText}>
                    <h1 className={styles.h1}>Tags</h1>
                    {/* <p className={styles.subtitle}>
                        View all tags in the database, jump to the Gallery by
                        clicking a tag, or delete ones you no longer want.
                    </p> */}
                </div>

                <button
                    className={styles.iconButton}
                    onClick={() => void loadTags({ silent: true })}
                    disabled={loading || refreshing || deleting}
                    title="Refresh"
                    aria-label="Refresh"
                >
                    <RefreshCcw className={styles.icon} />
                    <span className={styles.iconButtonText}>
                        {refreshing ? "Refreshing..." : "Refresh"}
                    </span>
                </button>
            </header>

            {banner && (
                <div
                    className={
                        banner.kind === "error"
                            ? styles.bannerError
                            : styles.bannerSuccess
                    }
                    role="status"
                >
                    <div className={styles.bannerLeft}>
                        {banner.kind === "error" ? (
                            <AlertTriangle className={styles.bannerIcon} />
                        ) : (
                            <CheckSquare className={styles.bannerIcon} />
                        )}
                        <span className={styles.bannerText}>{banner.text}</span>
                    </div>

                    <button
                        className={styles.bannerClose}
                        onClick={() => setBanner(null)}
                        aria-label="Dismiss message"
                        title="Dismiss"
                    >
                        <X className={styles.icon} />
                    </button>
                </div>
            )}

            <section className={styles.toolbar} aria-label="Tag controls">
                <div className={styles.searchWrap}>
                    <Search className={styles.searchIcon} />
                    <input
                        className={styles.searchInput}
                        value={filter}
                        onChange={(e) => setFilter(e.target.value)}
                        placeholder="Search tags..."
                        inputMode="search"
                        autoComplete="off"
                        aria-label="Search tags"
                    />
                    {filter.length > 0 && (
                        <button
                            className={styles.clearButton}
                            onClick={() => setFilter("")}
                            aria-label="Clear search"
                            title="Clear"
                        >
                            <X className={styles.icon} />
                        </button>
                    )}
                </div>

                <div className={styles.toolbarRight}>
                    <button
                        className={styles.button}
                        onClick={toggleSelectAllFiltered}
                        disabled={loading || deleting || filteredTags.length === 0}
                        title={
                            allFilteredSelected
                                ? "Unselect all (filtered)"
                                : "Select all (filtered)"
                        }
                    >
                        {allFilteredSelected ? (
                            <CheckSquare className={styles.icon} />
                        ) : (
                            <Square className={styles.icon} />
                        )}
                        <span>
                            {allFilteredSelected ? "Unselect all" : "Select all"}
                        </span>
                    </button>

                    <button
                        className={styles.dangerButton}
                        onClick={() => requestDelete([...selected])}
                        disabled={loading || deleting || selectedCount === 0}
                        title={
                            selectedCount === 0
                                ? "Select tags to delete"
                                : `Delete ${selectedCount} selected tag(s)`
                        }
                    >
                        <Trash2 className={styles.icon} />
                        <span>
                            Delete {selectedCount > 0 ? ` (${selectedCount})` : ""}
                        </span>
                    </button>
                </div>
            </section>

            <section className={styles.listSection} aria-label="Tag list">
                <div className={styles.listHeader}>
                    <div className={styles.listHeaderLeft}>
                        <span className={styles.listTitle}>All tags</span>
                        <span className={styles.listMeta}>
                            {loading ? "Loading..." : `${tags.length} total`}
                            {normalizedFilter ? ` • ${filteredTags.length} matching` : ""}
                            {countsOk && totalUsed !== null ? ` • ${totalUsed} uses` : ""}
                            {!countsOk && !loading ? " • counts unavailable" : ""}
                        </span>
                    </div>
                </div>

                <div className={styles.list} role="list">
                    {loading ? (
                        <div className={styles.skeletonWrap}>
                            {Array.from({ length: 8 }).map((_, i) => (
                                <div key={i} className={styles.skeletonRow} />
                            ))}
                        </div>
                    ) : filteredTags.length === 0 ? (
                        <div className={styles.emptyState}>
                            <p className={styles.emptyTitle}>No tags found</p>
                            <p className={styles.emptySubtitle}>
                                {normalizedFilter
                                    ? "Try a different search term."
                                    : "Your database doesn't have any tags yet."}
                            </p>
                        </div>
                    ) : (
                        filteredTags.map((tag) => {
                            const isSelected = selected.has(tag);
                            const count = countsOk ? (counts[tag] ?? 0) : null;

                            return (
                                <div
                                    key={tag}
                                    className={`${styles.row} ${
                                        isSelected ? styles.rowSelected : ""
                                    }`}
                                    role="listitem"
                                >
                                    <button
                                        className={styles.checkbox}
                                        onClick={() => toggleSelected(tag)}
                                        aria-label={
                                            isSelected
                                                ? `Unselect ${tag}`
                                                : `Select ${tag}`
                                        }
                                        title={
                                            isSelected
                                                ? `Unselect ${tag}`
                                                : `Select ${tag}`
                                        }
                                    >
                                        {isSelected ? (
                                            <CheckSquare className={styles.icon} />
                                        ) : (
                                            <Square className={styles.icon} />
                                        )}
                                    </button>

                                    <button
                                        type="button"
                                        className={styles.tagLink}
                                        onClick={() => goToGallery(tag)}
                                        title={`Open Gallery with ${tagToGalleryQuery(tag)}`}
                                        aria-label={`Open Gallery filtered by ${tag}`}
                                    >
                                        <span className={styles.tagName}>{tag}</span>
                                        <span className={styles.countBadge}>
                                            {count === null ? "—" : count}
                                        </span>
                                        <ArrowRight className={styles.tagArrow} />
                                    </button>

                                    <div className={styles.actions}>
                                        <button
                                            className={styles.rowDelete}
                                            onClick={() => requestDelete([tag])}
                                            disabled={deleting}
                                            aria-label={`Delete ${tag}`}
                                            title={`Delete ${tag}`}
                                        >
                                            <Trash2 className={styles.icon} />
                                        </button>
                                    </div>
                                </div>
                            );
                        })
                    )}
                </div>
            </section>

            {confirmOpen && (
                <div
                    className={styles.modalOverlay}
                    role="dialog"
                    aria-modal="true"
                    aria-label="Confirm delete tags"
                    onMouseDown={(e) => {
                        // click outside to close
                        if (e.target === e.currentTarget) {
                            setConfirmOpen(false);
                            setConfirmTags([]);
                        }
                    }}
                >
                    <div className={styles.modal}>
                        <div className={styles.modalHeader}>
                            <div className={styles.modalTitleRow}>
                                <AlertTriangle className={styles.modalIcon} />
                                <h2 className={styles.modalTitle}>Delete tag(s)?</h2>
                            </div>
                            <button
                                className={styles.modalClose}
                                onClick={() => {
                                    setConfirmOpen(false);
                                    setConfirmTags([]);
                                }}
                                aria-label="Close"
                                title="Close"
                                disabled={deleting}
                            >
                                <X className={styles.icon} />
                            </button>
                        </div>

                        <p className={styles.modalBody}>
                            This will permanently delete{" "}
                            <strong>{confirmTags.length}</strong>{" "}
                            {confirmTags.length === 1 ? "tag" : "tags"} and remove{" "}
                            {confirmTags.length === 1 ? "it" : "them"} everywhere
                            {confirmTags.length === 1 ? " it" : " they"} appear.
                        </p>

                        <div className={styles.modalTags}>
                            {confirmTags
                                .slice()
                                .sort((a, b) => a.localeCompare(b))
                                .slice(0, 12)
                                .map((t) => (
                                    <span key={t} className={styles.modalTagPill}>
                                        {t}
                                    </span>
                                ))}
                            {confirmTags.length > 12 && (
                                <span className={styles.modalMore}>
                                    +{confirmTags.length - 12} more
                                </span>
                            )}
                        </div>

                        <div className={styles.modalActions}>
                            <button
                                className={styles.button}
                                onClick={() => {
                                    setConfirmOpen(false);
                                    setConfirmTags([]);
                                }}
                                disabled={deleting}
                            >
                                Cancel
                            </button>

                            <button
                                className={styles.dangerButton}
                                onClick={() => void performDelete()}
                                disabled={deleting}
                            >
                                <Trash2 className={styles.icon} />
                                <span>{deleting ? "Deleting..." : "Delete"}</span>
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </main>
    );
}