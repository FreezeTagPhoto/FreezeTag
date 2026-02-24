"use client";

import { useMemo } from "react";
import styles from "./TagTable.module.css";
import { ArrowRight, CheckSquare, Square, Trash2 } from "lucide-react";

export type TagTableProps = {
    loading: boolean;

    // Full list (for header meta), and filtered list (for rows)
    tags: string[];
    filteredTags: string[];
    filterActive: boolean;

    countsOk: boolean;
    counts: Record<string, number>;

    selected: Set<string>;
    deleting: boolean;

    onToggleSelected: (tag: string) => void;
    onOpenGallery: (tag: string) => void;
    onRequestDelete: (tags: string[]) => void;
};

export default function TagTable({
    loading,
    tags,
    filteredTags,
    filterActive,
    countsOk,
    counts,
    selected,
    deleting,
    onToggleSelected,
    onOpenGallery,
    onRequestDelete,
}: TagTableProps) {
    const selectedCount = selected.size;

    const totalUsed = useMemo(() => {
        if (!countsOk) return null;
        let sum = 0;
        for (const t of tags) sum += counts[t] ?? 0;
        return sum;
    }, [countsOk, counts, tags]);

    return (
        <section className={styles.listSection} aria-label="Tag list">
            <div className={styles.listHeader}>
                <div className={styles.listHeaderLeft}>
                    <span className={styles.listTitle}>All tags</span>
                    <span className={styles.listMeta}>
                        {loading ? "Loading..." : `${tags.length} total`}
                        {filterActive ? ` • ${filteredTags.length} matching` : ""}
                        {countsOk && totalUsed !== null ? ` • ${totalUsed} uses` : ""}
                        {!countsOk && !loading ? " • counts unavailable" : ""}
                        {selectedCount > 0 ? ` • ${selectedCount} selected` : ""}
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
                            {filterActive
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
                                    onClick={() => onToggleSelected(tag)}
                                    aria-label={
                                        isSelected ? `Unselect ${tag}` : `Select ${tag}`
                                    }
                                    title={isSelected ? `Unselect ${tag}` : `Select ${tag}`}
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
                                    onClick={() => onOpenGallery(tag)}
                                    aria-label={`Open Gallery filtered by ${tag}`}
                                    title={`Open Gallery filtered by ${tag}`}
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
                                        onClick={() => onRequestDelete([tag])}
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
    );
}