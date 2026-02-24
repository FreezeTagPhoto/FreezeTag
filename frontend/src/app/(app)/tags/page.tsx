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
} from "lucide-react";

import TagTable from "@/components/Tags/TagTable/TagTable";
import Dialog from "@/components/UI/Dialog/Dialog";

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
    const filterActive = normalizedFilter.length > 0;

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
        filteredTags.length > 0 &&
        filteredSelectedCount === filteredTags.length;

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
            setCounts({});
            setCountsOk(false);
            setBanner((prev) => {
                const msg = `Failed to load tag counts (${countRes.error.status}): ${countRes.error.message}`;
                if (!prev) return { kind: "error", text: msg };
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
    }, []);

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

    const closeConfirm = () => {
        if (deleting) return;
        setConfirmOpen(false);
        setConfirmTags([]);
    };

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

    const sortedConfirmTags = useMemo(
        () => [...confirmTags].sort((a, b) => a.localeCompare(b)),
        [confirmTags],
    );

    const confirmShown = sortedConfirmTags.slice(0, 12);
    const confirmExtra = Math.max(
        0,
        sortedConfirmTags.length - confirmShown.length,
    );

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
                        disabled={
                            loading || deleting || filteredTags.length === 0
                        }
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
                            {allFilteredSelected
                                ? "Unselect all"
                                : "Select all"}
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
                            Delete{" "}
                            {selectedCount > 0 ? ` (${selectedCount})` : ""}
                        </span>
                    </button>
                </div>
            </section>

            <TagTable
                loading={loading}
                tags={tags}
                filteredTags={filteredTags}
                filterActive={filterActive}
                countsOk={countsOk}
                counts={counts}
                selected={selected}
                deleting={deleting}
                onToggleSelected={toggleSelected}
                onOpenGallery={goToGallery}
                onRequestDelete={requestDelete}
            />

            <Dialog
                open={confirmOpen}
                onClose={closeConfirm}
                ariaLabel="Confirm delete tags"
                title="Delete tag(s)?"
                icon={<AlertTriangle className={styles.dialogIcon} />}
                disableClose={deleting}
            >
                <p className={styles.dialogBody}>
                    This will permanently delete {confirmTags.length}{" "}
                    {confirmTags.length === 1 ? "tag" : "tags"} and remove{" "}
                    {confirmTags.length === 1 ? "it" : "them"} everywhere
                    {confirmTags.length === 1 ? " it appears" : " they appear"}.
                </p>

                <div className={styles.dialogTags}>
                    {confirmShown.map((t) => (
                        <span key={t} className={styles.dialogTagPill}>
                            {t}
                        </span>
                    ))}
                    {confirmExtra > 0 && (
                        <span className={styles.dialogMore}>
                            +{confirmExtra} more
                        </span>
                    )}
                </div>

                <div className={styles.dialogActions}>
                    <button
                        className={styles.button}
                        onClick={closeConfirm}
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
            </Dialog>
        </main>
    );
}
