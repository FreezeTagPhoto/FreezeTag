"use client";

import { useEffect, useState, useMemo } from "react";
import { useParams, useRouter } from "next/navigation";
import { ArrowLeft, Check, ChevronDown, ChevronUp, Tags } from "lucide-react";

import SearchHandler from "@/api/query/searchhandler";
import AlbumImageAdder from "@/api/albums/albumimageadder";
import AlbumImagesGetter from "@/api/albums/albumimagesgetter";
import AlbumGetter from "@/api/albums/albumgetter";
import TagGetter from "@/api/tags/taggetter";
import TagCounter from "@/api/tags/tagcounter";

import MassTaggingGallery from "@/components/Gallery/MassTaggingGallery/MassTaggingGallery";
import SearchBar from "@/components/SearchBar/SearchBar";
import FreezeTagSet from "@/common/freezetag/freezetagset";
import { addTagToQuery } from "@/common/search/addtagtoquery";
import { parseUserQuery } from "@/common/search/parse";
import styles from "./AlbumEditPage.module.css";

// --- Query Helpers ---
function normalizeUserTail(s: string) {
    return s
        .trim()
        .replace(/^\s*;\s*/, "")
        .replace(/\s*;\s*$/, "")
        .trim();
}
function buildQuery(sortBy: string, sortOrder: string, searchTerm: string) {
    const tail = normalizeUserTail(searchTerm);
    return tail
        ? `sortBy=${sortBy};sortOrder=${sortOrder};${tail}`
        : `sortBy=${sortBy};sortOrder=${sortOrder}`;
}
function removeTokenFromQuery(input: string, start: number, end: number) {
    const before = input.slice(0, start).replace(/;\s*$/, "");
    const after = input.slice(end).replace(/^\s*;\s*/, "");
    let out = before.replace(/\s+$/, "");
    if (out.length > 0 && after.trim().length > 0) out += "; ";
    out += after.replace(/^\s+/, "");
    return out
        .replace(/^\s*;\s*/, "")
        .replace(/\s*;\s*$/, "")
        .trim();
}
function tryRemoveTagFromQuery(input: string, tag: string) {
    const tokens = parseUserQuery(input)
        .filter((t) => !t.error && t.kind === "tag" && t.value === tag)
        .sort((a, b) => b.range.start - a.range.start);
    if (tokens.length === 0) return input;
    let out = input;
    for (const tok of tokens)
        out = removeTokenFromQuery(out, tok.range.start, tok.range.end);
    const trimmed = out.trim();
    if (!trimmed) return "";
    return trimmed.endsWith(";") || trimmed.endsWith("; ")
        ? trimmed
        : `${trimmed}; `;
}

export default function AlbumEditPage() {
    const router = useRouter();
    const albumId = parseInt(useParams().id as string, 10);

    const [albumName, setAlbumName] = useState("");
    const [existingIds, setExistingIds] = useState<Set<number>>(new Set());
    const [availableIds, setAvailableIds] = useState<number[]>([]);
    const [selectedIds, setSelectedIds] = useState(new FreezeTagSet<number>());

    const [searchTerm, setSearchTerm] = useState("");
    const [sortBy, setSortBy] = useState("DateAdded");
    const [sortOrder, setSortOrder] = useState("DESC");
    const [allTags, setAllTags] = useState<string[]>([]);
    const [tagCounts, setTagCounts] = useState<Record<string, number>>({});
    const [tagsOpen, setTagsOpen] = useState(false);

    useEffect(() => {
        if (!albumId || isNaN(albumId)) return;
        Promise.all([
            AlbumGetter(albumId),
            AlbumImagesGetter(albumId),
            TagGetter(),
        ]).then(([album, images, tags]) => {
            if (album.ok) setAlbumName(album.value.name || "Album");
            if (images.ok) setExistingIds(new Set(images.value));
            if (tags.ok) setAllTags(tags.value);
        });
    }, [albumId]);

    const fullQuery = useMemo(
        () => buildQuery(sortBy, sortOrder, searchTerm),
        [sortBy, sortOrder, searchTerm],
    );
    const activeTags = useMemo(
        () =>
            new Set(
                parseUserQuery(searchTerm)
                    .filter((t) => !t.error && t.kind === "tag")
                    .map((t) => t.value),
            ),
        [searchTerm],
    );
    const visibleTags = useMemo(
        () =>
            allTags.filter((t) => (tagCounts[t] ?? 0) > 0 || activeTags.has(t)),
        [allTags, tagCounts, activeTags],
    );
    const filteredIds = useMemo(
        () => availableIds.filter((id) => !existingIds.has(id)),
        [availableIds, existingIds],
    );

    useEffect(() => {
        const fetchTimer = setTimeout(() => {
            Promise.all([SearchHandler(fullQuery), TagCounter(fullQuery)]).then(
                ([searchRes, countRes]) => {
                    if (searchRes.ok) setAvailableIds(searchRes.value);
                    if (countRes.ok) setTagCounts(countRes.value);
                },
            );
        }, 300);
        return () => clearTimeout(fetchTimer);
    }, [fullQuery]);

    const toggleTag = (tag: string) =>
        setSearchTerm((prev) => {
            const removed = tryRemoveTagFromQuery(prev, tag);
            return removed === prev ? addTagToQuery(prev, tag) : removed;
        });

    const handleAdd = async () => {
        const ids = selectedIds.toArray();
        if (ids.length === 0) return;
        await Promise.all(ids.map((id) => AlbumImageAdder(id, albumId)));
        router.push(`/albums/${albumId}`);
        router.refresh();
    };

    return (
        <section className={styles.container}>
            <header className={styles.header}>
                <div className={styles.navSection}>
                    <button
                        onClick={() => router.back()}
                        className={styles.backBtn}
                    >
                        <ArrowLeft size={20} />
                    </button>
                    <div>
                        <h1 className={styles.title}>Add to {albumName}</h1>
                        <p className={styles.subtle}>
                            {filteredIds.length} available &bull;{" "}
                            {selectedIds.size()} selected
                        </p>
                    </div>
                </div>
                <div className={styles.actions}>
                    <button
                        className={styles.button}
                        onClick={() =>
                            setSelectedIds(new FreezeTagSet(filteredIds))
                        }
                    >
                        Select All
                    </button>
                    <button
                        className={styles.button}
                        onClick={() => setSelectedIds(new FreezeTagSet())}
                    >
                        Clear
                    </button>
                    <button
                        onClick={handleAdd}
                        disabled={selectedIds.size() === 0}
                        className={styles.saveBtn}
                    >
                        <Check size={18} /> Add
                    </button>
                </div>
            </header>

            <div className={styles.filterBar}>
                <div className={styles.searchArea}>
                    <SearchBar
                        value={searchTerm}
                        onChange={setSearchTerm}
                        allTags={allTags}
                        placeholder="Search..."
                        enabled={true}
                    />
                </div>

                <select
                    className={styles.select}
                    value={sortBy}
                    onChange={(e) => setSortBy(e.target.value)}
                >
                    <option value="DateCreated">Date Created</option>
                    <option value="DateAdded">Date Added</option>
                </select>

                <select
                    className={styles.select}
                    value={sortOrder}
                    onChange={(e) => setSortOrder(e.target.value)}
                >
                    <option value="DESC">Newest</option>
                    <option value="ASC">Oldest</option>
                </select>

                <button
                    className={styles.button}
                    onClick={() => setTagsOpen(!tagsOpen)}
                >
                    <Tags size={16} /> <span>Tags</span>{" "}
                    {tagsOpen ? (
                        <ChevronUp size={14} />
                    ) : (
                        <ChevronDown size={14} />
                    )}
                </button>
            </div>

            {tagsOpen && (
                <div className={styles.tagPanel}>
                    {visibleTags.map((tag) => (
                        <button
                            key={tag}
                            className={`${styles.tagPill} ${activeTags.has(tag) ? styles.tagPillActive : ""}`}
                            onClick={() => toggleTag(tag)}
                        >
                            <span>{tag}</span>{" "}
                            <span className={styles.tagCount}>
                                {tagCounts[tag] ?? 0}
                            </span>
                        </button>
                    ))}
                </div>
            )}

            <main className={styles.gridArea}>
                {filteredIds.length === 0 ? (
                    <p className={styles.empty}>No images match this search.</p>
                ) : (
                    <MassTaggingGallery
                        image_ids={filteredIds}
                        selectedIds={selectedIds}
                        onChange={(id) =>
                            setSelectedIds((prev) => prev.toggle(id))
                        }
                    />
                )}
            </main>
        </section>
    );
}
