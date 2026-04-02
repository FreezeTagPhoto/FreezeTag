"use client";

import { useEffect, useState, useMemo, useRef } from "react";
import { useRouter } from "next/navigation";
import { ArrowLeft, Check, ChevronDown, ChevronUp, Tags, Loader2 } from "lucide-react";

import SearchHandler from "@/api/query/searchhandler";
import AlbumImageAdder from "@/api/albums/albumimageadder";
import AlbumImagesGetter from "@/api/albums/albumimagesgetter";
import TagGetter from "@/api/tags/taggetter";
import TagCounter from "@/api/tags/tagcounter";

import MassTaggingGallery from "@/components/Gallery/MassTaggingGallery/MassTaggingGallery";
import SearchBar from "@/components/SearchBar/SearchBar";
import FreezeTagSet from "@/common/freezetag/freezetagset";
import { addTagToQuery } from "@/common/search/addtagtoquery";
import { parseUserQuery } from "@/common/search/parse";
import styles from "./AlbumUpdatePage.module.css";

// --- Query Helpers (Preserved from your original logic) ---
function normalizeUserTail(s: string): string {
    return s.trim().replace(/^\s*;\s*/, "").replace(/\s*;\s*$/, "").trim();
}

function buildQuery(sortBy: string, sortOrder: string, searchTerm: string): string {
    const tail = normalizeUserTail(searchTerm);
    return tail ? `sortBy=${sortBy};sortOrder=${sortOrder};${tail}` : `sortBy=${sortBy};sortOrder=${sortOrder}`;
}

function removeTokenFromQuery(input: string, start: number, end: number): string {
    let before = input.slice(0, start);
    let after = input.slice(end);
    after = after.replace(/^\s*;\s*/, "");
    before = before.replace(/;\s*$/, "");
    let out = before.replace(/\s+$/, "");
    if (out.length > 0 && after.trim().length > 0) out += "; ";
    out += after.replace(/^\s+/, "");
    return out.replace(/^\s*;\s*/, "").replace(/\s*;\s*$/, "").trim();
}

function tryRemoveTagFromQuery(input: string, tag: string): string {
    const tokens = parseUserQuery(input)
        .filter((t) => !t.error && t.kind === "tag" && t.value === tag)
        .sort((a, b) => b.range.start - a.range.start);

    if (tokens.length === 0) return input;
    let out = input;
    for (const tok of tokens) out = removeTokenFromQuery(out, tok.range.start, tok.range.end);
    const trimmed = out.trim();
    if (!trimmed) return "";
    if (trimmed.endsWith(";")) return `${trimmed} `;
    if (trimmed.endsWith("; ")) return trimmed;
    return `${trimmed}; `;
}

const SORT_BY_OPTIONS = [{ value: "DateCreated", label: "Date Created" }, { value: "DateAdded", label: "Date Added" }];
const ORDER_OPTIONS = [{ value: "DESC", label: "Newest first" }, { value: "ASC", label: "Oldest first" }];

export default function AlbumEditPage({ albumName }: { albumName: string }) {
    const router = useRouter();

    const [existingIds, setExistingIds] = useState<Set<number>>(new Set());
    const [availableIds, setAvailableIds] = useState<number[]>([]);
    const [selectedIds, setSelectedIds] = useState(new FreezeTagSet<number>());
    
    const [searchTerm, setSearchTerm] = useState("");
    const [sortBy, setSortBy] = useState("DateAdded");
    const [sortOrder, setSortOrder] = useState("DESC");
    const [allTags, setAllTags] = useState<string[]>([]);
    const [tagCounts, setTagCounts] = useState<Record<string, number>>({});
    
    const [loadingImages, setLoadingImages] = useState(true);
    const [busy, setBusy] = useState(false);
    const [sortOpen, setSortOpen] = useState(false);
    const [tagsOpen, setTagsOpen] = useState(false);
    const sortWrapRef = useRef<HTMLDivElement | null>(null);

    useEffect(() => {
        AlbumImagesGetter(albumName).then(res => {
            if (res.ok) setExistingIds(new Set(res.value));
        });
        TagGetter().then(res => {
            if (res.ok) setAllTags(res.value);
        });
    }, [albumName]);

    useEffect(() => {
        if (!sortOpen) return;
        const onMouseDown = (e: MouseEvent) => {
            if (!sortWrapRef.current?.contains(e.target as Node)) setSortOpen(false);
        };
        document.addEventListener("mousedown", onMouseDown);
        return () => document.removeEventListener("mousedown", onMouseDown);
    }, [sortOpen]);

    const fullQuery = useMemo(() => buildQuery(sortBy, sortOrder, searchTerm), [sortBy, sortOrder, searchTerm]);
    
    const activeTags = useMemo(() => {
        const set = new Set<string>();
        for (const tok of parseUserQuery(searchTerm)) {
            if (!tok.error && tok.kind === "tag") set.add(tok.value);
        }
        return set;
    }, [searchTerm]);

    const visibleTags = useMemo(() => {
        return [...allTags]
            .filter((t) => (tagCounts[t] ?? 0) > 0 || activeTags.has(t))
            .sort((a, b) => a.localeCompare(b));
    }, [allTags, tagCounts, activeTags]);

    const filteredIds = useMemo(() => availableIds.filter(id => !existingIds.has(id)), [availableIds, existingIds]);

    useEffect(() => {
        let active = true;
        setLoadingImages(true);

        Promise.all([SearchHandler(fullQuery), TagCounter(fullQuery)]).then(([searchRes, countRes]) => {
            if (!active) return;
            if (searchRes.ok) setAvailableIds(searchRes.value);
            if (countRes.ok) setTagCounts(countRes.value);
            setLoadingImages(false);
        });

        return () => { active = false; };
    }, [fullQuery]);

    const toggleTag = (tag: string) => {
        setSearchTerm(prev => {
            const removed = tryRemoveTagFromQuery(prev, tag);
            return removed === prev ? addTagToQuery(prev, tag) : removed;
        });
    };

    const handleAdd = async () => {
        const ids = selectedIds.toArray();
        if (ids.length === 0 || busy) return;
        setBusy(true);
        await Promise.all(ids.map(id => AlbumImageAdder(id, albumName)));
        router.push(`/albums/${encodeURIComponent(albumName)}`);
        router.refresh();
    };

    return (
        <section className={styles.container}>
            <header className={styles.header}>
                <div className={styles.nav}>
                    <button onClick={() => router.back()} className={styles.backBtn}><ArrowLeft size={20} /></button>
                    <div>
                        <h1>Add to {albumName}</h1>
                        <p className={styles.subtle}>{filteredIds.length} available &bull; {selectedIds.size()} selected</p>
                    </div>
                </div>
                <div className={styles.actions}>
                    <button className={styles.button} onClick={() => setSelectedIds(new FreezeTagSet(filteredIds))} disabled={busy || loadingImages}>Select All</button>
                    <button className={styles.button} onClick={() => setSelectedIds(new FreezeTagSet())} disabled={busy}>Clear</button>
                    <button onClick={handleAdd} disabled={busy || selectedIds.size() === 0} className={styles.saveBtn}>
                        {busy ? <Loader2 className={styles.spin} size={18} /> : <Check size={18} />} Add Selected
                    </button>
                </div>
            </header>

            <div className={styles.toolbar}>
                <div className={styles.searchWrap}>
                    <SearchBar 
                        enabled={!busy} 
                        value={searchTerm} 
                        onChange={setSearchTerm} 
                        allTags={allTags} 
                        placeholder="Search library..." 
                    />
                </div>

                <div className={styles.sortMenuWrap} ref={sortWrapRef}>
                    <button className={styles.button} onClick={() => setSortOpen(!sortOpen)} disabled={busy}>
                        <span>Sort</span> {sortOpen ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
                    </button>
                    {sortOpen && (
                        <div className={styles.dropdown}>
                            <div className={styles.dropdownSectionTitle}>Sort by</div>
                            <div className={styles.dropdownGrid}>
                                {SORT_BY_OPTIONS.map(opt => (
                                    <button key={opt.value} className={`${styles.dropdownItem} ${sortBy === opt.value ? styles.dropdownItemActive : ""}`} onClick={() => setSortBy(opt.value)}>{opt.label}</button>
                                ))}
                            </div>
                            <div className={styles.dropdownSectionTitle}>Order</div>
                            <div className={styles.dropdownGrid}>
                                {ORDER_OPTIONS.map(opt => (
                                    <button key={opt.value} className={`${styles.dropdownItem} ${sortOrder === opt.value ? styles.dropdownItemActive : ""}`} onClick={() => setSortOrder(opt.value)}>{opt.label}</button>
                                ))}
                            </div>
                        </div>
                    )}
                </div>

                <button className={styles.button} onClick={() => setTagsOpen(!tagsOpen)} disabled={busy}>
                    <Tags size={16} /> <span>Tags</span> {tagsOpen ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
                </button>
            </div>

            {tagsOpen && (
                <div className={styles.tagPanel}>
                    {visibleTags.map(tag => {
                        const active = activeTags.has(tag);
                        return (
                            <button key={tag} className={`${styles.tagPill} ${active ? styles.tagPillActive : ""}`} onClick={() => toggleTag(tag)} disabled={busy}>
                                <span>{tag}</span> <span className={styles.tagCount}>{tagCounts[tag] ?? 0}</span>
                            </button>
                        );
                    })}
                </div>
            )}

            <main className={styles.content}>
                {loadingImages ? (
                    <div className={styles.center}><Loader2 className={styles.spin} size={40} /></div>
                ) : filteredIds.length === 0 ? (
                    <p className={styles.empty}>No images match this search.</p>
                ) : (
                    <MassTaggingGallery image_ids={filteredIds} selectedIds={selectedIds} onChange={(id) => setSelectedIds(prev => prev.toggle(id))} />
                )}
            </main>
        </section>
    );
}