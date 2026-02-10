"use client";

import {
    useEffect,
    useState,
    useRef,
    MouseEvent as ReactMouseEvent,
    KeyboardEvent as ReactKeyboardEvent,
    useCallback,
    useMemo,
} from "react";
import styles from "./MainGallery.module.css";
import GalleryImage from "../GalleryImage/GalleryImage";
import MetadataGetter, { ImageMetadata } from "@/api/metadata/metadatagetter";
import TagGetter from "@/api/tags/taggetter";
import TagAdder from "@/api/tags/tagadder";
import TagRemover from "@/api/tags/tagremover";
import Pill from "@/components/UI/Pill/Pill";

export type GalleryProps = {
    image_ids: number[];
    onSearchTag?: (tag: string) => void;
};

// point (fx, fy) on image expressed as fraction of width/height (after zoom)
type PendingPan = null | { fx: number; fy: number };

// if int64 timestamp is huge, treat as ms, otherwise seconds
function toDate(ts: number): Date {
    return new Date(ts > 1e12 ? ts : ts * 1000);
}

function formatDate(ts: number | null): string {
    if (ts === null) return "—";
    const d = toDate(ts);
    return new Intl.DateTimeFormat(undefined, {
        year: "numeric",
        month: "short",
        day: "numeric",
        hour: "numeric",
        minute: "2-digit",
    }).format(d);
}

function formatLocation(lat: number | null, lon: number | null): string {
    if (lat === null || lon === null) return "—";
    return `${lat.toFixed(5)}, ${lon.toFixed(5)}`;
}

function formatCamera(make: string | null, model: string | null): string {
    const parts = [make, model].filter(
        (x) => x && x.trim().length > 0,
    ) as string[];
    return parts.length ? parts.join(" ") : "—";
}

function normalizeTag(s: string) {
    return s.trim().replace(/\s+/g, " ");
}

function isSubsequence(needle: string, hay: string) {
    let i = 0;
    for (let j = 0; j < hay.length && i < needle.length; j++) {
        if (hay[j] === needle[i]) i++;
    }
    return i === needle.length;
}

function rankTag(tag: string, needleRaw: string) {
    const needle = needleRaw.toLowerCase();
    const t = tag.toLowerCase();

    if (t === needle) return 0;
    if (t.startsWith(needle)) return 1;

    const idx = t.indexOf(needle);
    if (idx !== -1) return 2 + idx / 100;

    if (isSubsequence(needle, t)) return 3;

    return 999;
}

export default function MainGallery({ image_ids, onSearchTag }: GalleryProps) {
    // Full Screen Preview Handling
    const [selectedId, setSelectedId] = useState<number | null>(null);
    const [focusedIndex, setFocusedIndex] = useState<number | null>(null);

    const itemRefs = useRef<(HTMLButtonElement | null)[]>([]);

    // zoom: 1 = fit, 2 = zoomed
    const [zoom, setZoom] = useState<number>(1);
    const scrollRef = useRef<HTMLDivElement | null>(null);
    const [hoveringImage, setHoveringImage] = useState(false);

    // remember where to pan after zoom has been applied
    const [pendingPan, setPendingPan] = useState<PendingPan>(null);

    // metadata state
    const [metadataById, setMetadataById] = useState<
        Record<number, ImageMetadata>
    >({});
    const [metadataLoading, setMetadataLoading] = useState(false);
    const [metadataError, setMetadataError] = useState<string | null>(null);

    const currentMetadata: ImageMetadata | null =
        selectedId !== null ? (metadataById[selectedId] ?? null) : null;

    // fetch metadata whenever selectedId changes
    useEffect(() => {
        if (selectedId === null) return;

        // already cached
        if (metadataById[selectedId]) {
            setMetadataError(null);
            setMetadataLoading(false);
            return;
        }

        let cancelled = false;

        (async () => {
            setMetadataLoading(true);
            setMetadataError(null);

            const res = await MetadataGetter(selectedId);

            if (cancelled) return;

            if (!res.ok) {
                setMetadataError(res.error.message);
                setMetadataLoading(false);
                return;
            }

            setMetadataById((prev) => ({ ...prev, [selectedId]: res.value }));
            setMetadataLoading(false);
        })();

        return () => {
            cancelled = true;
        };
    }, [selectedId, metadataById]);

    // tags state
    const [tagsById, setTagsById] = useState<Record<number, string[]>>({});
    const [tagsLoading, setTagsLoading] = useState(false);
    const [tagsError, setTagsError] = useState<string | null>(null);

    const currentTags: string[] | null =
        selectedId !== null ? (tagsById[selectedId] ?? null) : null;

    useEffect(() => {
        if (selectedId === null) return;

        // already cached
        if (tagsById[selectedId]) {
            setTagsError(null);
            setTagsLoading(false);
            return;
        }

        let cancelled = false;

        (async () => {
            setTagsLoading(true);
            setTagsError(null);

            const res = await TagGetter(selectedId);

            if (cancelled) return;

            if (!res.ok) {
                setTagsError(res.error.message);
                setTagsLoading(false);
                return;
            }

            setTagsById((prev) => ({ ...prev, [selectedId]: res.value }));
            setTagsLoading(false);
        })();

        return () => {
            cancelled = true;
        };
    }, [selectedId, tagsById]);

    const moveSelection = useCallback(
        (direction: "next" | "prev") => {
            if (selectedId === null) return;

            const currentIndex = image_ids.indexOf(selectedId);
            if (currentIndex === -1) return;

            const delta = direction === "next" ? 1 : -1;
            const nextIndex = currentIndex + delta;

            if (nextIndex < 0 || nextIndex >= image_ids.length) return;

            const nextId = image_ids[nextIndex];
            setSelectedId(nextId);
            setFocusedIndex(nextIndex);
        },
        [selectedId, image_ids],
    );

    // reset zoom + pan when opening / changing image
    useEffect(() => {
        setZoom(1);
        setPendingPan(null);
        const scroller = scrollRef.current;
        if (scroller) {
            scroller.scrollLeft = 0;
            scroller.scrollTop = 0;
        }
    }, [selectedId]);

    useEffect(() => {
        if (selectedId === null) return;

        const handleKeyDown = (event: KeyboardEvent) => {
            if (event.key === "Escape") {
                event.preventDefault();
                event.stopPropagation();
                setSelectedId(null);
                return;
            }

            if (event.key !== "ArrowRight" && event.key !== "ArrowLeft") {
                return;
            }

            event.preventDefault();
            event.stopPropagation();

            if (event.key === "ArrowRight") {
                moveSelection("next");
            } else if (event.key === "ArrowLeft") {
                moveSelection("prev");
            }
        };

        window.addEventListener("keydown", handleKeyDown, true);
        return () => window.removeEventListener("keydown", handleKeyDown, true);
    }, [selectedId, moveSelection]);

    const handleBackdropClick = () => {
        setSelectedId(null);
    };

    const stopPropagation = (e: ReactMouseEvent<HTMLDivElement>) => {
        e.stopPropagation();
    };

    const handleGridKeyDown = (event: ReactKeyboardEvent<HTMLDivElement>) => {
        if (selectedId !== null) return;

        const count = image_ids.length;
        if (count === 0) return;

        switch (event.key) {
            case "ArrowRight": {
                event.preventDefault();
                const next =
                    focusedIndex === null
                        ? 0
                        : Math.min(focusedIndex + 1, count - 1);
                setFocusedIndex(next);
                itemRefs.current[next]?.focus();
                break;
            }
            case "ArrowLeft": {
                event.preventDefault();
                const prev =
                    focusedIndex === null
                        ? count - 1
                        : Math.max(focusedIndex - 1, 0);
                setFocusedIndex(prev);
                itemRefs.current[prev]?.focus();
                break;
            }
            case "Escape": {
                event.preventDefault();

                if (focusedIndex !== null) {
                    itemRefs.current[focusedIndex]?.blur();
                    setFocusedIndex(null);
                } else {
                    (event.currentTarget as HTMLElement).blur();
                }
                break;
            }
        }
    };

    const zoomOut = () => {
        setZoom(1);
        setPendingPan(null);
        const scroller = scrollRef.current;
        if (!scroller) return;

        scroller.scrollLeft = 0;
        scroller.scrollTop = 0;
    };

    // zoom to center of image: (fx, fy) = (0.5, 0.5)
    const zoomInCentered = () => {
        setZoom(2);
        setPendingPan({ fx: 0.5, fy: 0.5 });
    };

    const handleZoomButtonClick = (event: ReactMouseEvent<HTMLButtonElement>) => {
        event.stopPropagation();
        if (zoom === 1) {
            zoomInCentered();
        } else {
            zoomOut();
        }
    };

    // if click the image at 1×, it zooms to 2× and pans so clicked spot is centered
    // if click again at 2×, it zooms back out.
    const handleImageClick = (event: ReactMouseEvent<HTMLImageElement>) => {
        const scroller = scrollRef.current;
        if (!scroller) return; // if no scroll container, escape

        event.stopPropagation();

        if (zoom === 1) {
            const imgRect = event.currentTarget.getBoundingClientRect();
            const fx = (event.clientX - imgRect.left) / imgRect.width;
            const fy = (event.clientY - imgRect.top) / imgRect.height;

            setZoom(2);
            setPendingPan({ fx, fy });
        } else {
            zoomOut();
        }
    };

    // after zoom changes, pan so that the chosen point (fx, fy) is at the center
    useEffect(() => {
        if (zoom === 1 || !pendingPan) return;

        const scroller = scrollRef.current;
        if (!scroller) return;

        const scrollWidth = scroller.scrollWidth;
        const scrollHeight = scroller.scrollHeight;
        const clientWidth = scroller.clientWidth;
        const clientHeight = scroller.clientHeight;

        const maxLeft = Math.max(0, scrollWidth - clientWidth);
        const maxTop = Math.max(0, scrollHeight - clientHeight);

        const rawLeft = pendingPan.fx * scrollWidth - clientWidth / 2;
        const rawTop = pendingPan.fy * scrollHeight - clientHeight / 2;

        const targetLeft = Math.max(0, Math.min(rawLeft, maxLeft));
        const targetTop = Math.max(0, Math.min(rawTop, maxTop));

        scroller.scrollLeft = targetLeft;
        scroller.scrollTop = targetTop;

        // clear so this runs only once per zoom action
        setPendingPan(null);
    }, [zoom, pendingPan]);

    // -----------------------------
    // Tag mutations + add dropdown
    // -----------------------------

    const [tagMutating, setTagMutating] = useState(false);
    const [tagMutateError, setTagMutateError] = useState<string | null>(null);
    const [tagMutateInfo, setTagMutateInfo] = useState<string | null>(null);

    const [addOpen, setAddOpen] = useState(false);
    const [addValue, setAddValue] = useState("");
    const addInputRef = useRef<HTMLInputElement | null>(null);
    const addEditorRef = useRef<HTMLDivElement | null>(null);

    const [allTags, setAllTags] = useState<string[] | null>(null);
    const [allTagsLoading, setAllTagsLoading] = useState(false);

    const [tagSuggestOpen, setTagSuggestOpen] = useState(false);
    const [tagSuggestPinned, setTagSuggestPinned] = useState(false);
    const [tagSuggestIndex, setTagSuggestIndex] = useState(0);

    const ensureAllTagsLoaded = useCallback(async () => {
        if (allTags !== null || allTagsLoading) return;
        setAllTagsLoading(true);
        const res = await TagGetter();
        if (res.ok) setAllTags(res.value);
        else setAllTags([]); // fail-safe
        setAllTagsLoading(false);
    }, [allTags, allTagsLoading]);

    const setCurrentTagsForSelected = useCallback(
        (next: string[]) => {
            if (selectedId === null) return;
            setTagsById((prev) => ({ ...prev, [selectedId]: next }));
        },
        [selectedId],
    );

    const removeTagFromSelected = useCallback(
        async (tag: string) => {
            if (selectedId === null) return;
            const current = tagsById[selectedId] ?? [];
            const next = current.filter((t) => t !== tag);

            setTagMutateError(null);
            setTagMutateInfo(null);
            setTagMutating(true);

            // optimistic UI
            setCurrentTagsForSelected(next);

            const res = await TagRemover([selectedId], [tag]);
            setTagMutating(false);

            if (!res.ok) {
                // revert
                setCurrentTagsForSelected(current);
                setTagMutateError(res.error.message);
                return;
            }

            if (res.value.length > 0) {
                setTagMutateInfo(res.value.join(" • "));
            }
        },
        [selectedId, tagsById, setCurrentTagsForSelected],
    );

    const addTagToSelected = useCallback(
        async (tagOverride?: string) => {
            if (selectedId === null) return;

            const tag = normalizeTag(tagOverride ?? addValue);
            if (!tag) return;

            const current = tagsById[selectedId] ?? [];

            if (current.includes(tag)) {
                setTagMutateInfo(`Tag "${tag}" already exists.`);
                setAddOpen(false);
                setAddValue("");
                setTagSuggestOpen(false);
                setTagSuggestPinned(false);
                return;
            }

            const next = [...current, tag];

            setTagMutateError(null);
            setTagMutateInfo(null);
            setTagMutating(true);

            // optimistic
            setCurrentTagsForSelected(next);

            const res = await TagAdder([selectedId], [tag]);
            setTagMutating(false);

            if (!res.ok) {
                setCurrentTagsForSelected(current);
                setTagMutateError(res.error.message);
                return;
            }

            setAllTags((prev) => {
                if (!prev) return prev;
                return prev.includes(tag) ? prev : [...prev, tag];
            });

            if (res.value.length > 0) setTagMutateInfo(res.value.join(" • "));

            setAddOpen(false);
            setAddValue("");
            setTagSuggestOpen(false);
            setTagSuggestPinned(false);
        },
        [selectedId, addValue, tagsById, setCurrentTagsForSelected],
    );

    // reset tag editor UI when changing images
    useEffect(() => {
        setTagMutateError(null);
        setTagMutateInfo(null);
        setAddOpen(false);
        setAddValue("");
        setTagSuggestOpen(false);
        setTagSuggestPinned(false);
        setTagSuggestIndex(0);
    }, [selectedId]);

    // focus input when opening editor
    useEffect(() => {
        if (!addOpen) return;
        requestAnimationFrame(() => addInputRef.current?.focus());
    }, [addOpen]);

    const tagSuggestions = useMemo(() => {
        if (!addOpen) return [];
        if (!allTags || allTags.length === 0) return [];

        const current = new Set((currentTags ?? []).map((t) => t));
        const candidates = allTags.filter((t) => !current.has(t));

        const needle = normalizeTag(addValue);

        // If user pinned open via ▾, allow empty input to show top suggestions.
        // Otherwise, empty input yields an empty list (prevents auto-open)
        const allowEmpty = tagSuggestPinned;

        if (!needle) {
            if (!allowEmpty) return [];
            return [...candidates].sort((a, b) => a.localeCompare(b)).slice(0, 10);
        }

        return candidates
            .map((t) => ({ tag: t, score: rankTag(t, needle) }))
            .filter((x) => x.score < 999)
            .sort((a, b) => a.score - b.score || a.tag.localeCompare(b.tag))
            .slice(0, 10)
            .map((x) => x.tag);
    }, [addOpen, allTags, currentTags, addValue, tagSuggestPinned]);

    const showTagDropdown =
        tagSuggestOpen && (allTagsLoading || tagSuggestions.length > 0);

    // keep index in bounds
    useEffect(() => {
        if (!tagSuggestOpen) return;
        setTagSuggestIndex((i) =>
            Math.max(0, Math.min(i, Math.max(0, tagSuggestions.length - 1))),
        );
    }, [tagSuggestOpen, tagSuggestions.length]);

    // click outside editor closes suggestions (and unpins)
    useEffect(() => {
        if (!addOpen) return;

        const onMouseDown = (e: MouseEvent) => {
            const root = addEditorRef.current;
            if (!root) return;
            if (!root.contains(e.target as Node)) {
                setTagSuggestOpen(false);
                setTagSuggestPinned(false);
            }
        };

        document.addEventListener("mousedown", onMouseDown);
        return () => document.removeEventListener("mousedown", onMouseDown);
    }, [addOpen]);

    const toggleSuggestions = async () => {
        const nextPinned = !tagSuggestPinned;
        setTagSuggestPinned(nextPinned);

        if (nextPinned) {
            await ensureAllTagsLoaded();
            setTagSuggestIndex(0);
            setTagSuggestOpen(true);
        } else {
            // only keep open if user has typed something
            const hasNeedle = normalizeTag(addValue).length > 0;
            setTagSuggestOpen(hasNeedle);
        }
    };

    return (
        <>
            {/* thumbnails */}
            <div
                className={styles.grid}
                tabIndex={0}
                onKeyDown={handleGridKeyDown}
                aria-label="Photo gallery"
            >
                {image_ids.map((id, index) => (
                    <GalleryImage
                        key={id}
                        id={id}
                        onClick={() => {
                            setSelectedId(id);
                        }}
                        onFocus={() => {
                            setFocusedIndex(index);
                        }}
                        buttonRef={(el) => {
                            itemRefs.current[index] = el;
                        }}
                        selected={false}
                    />
                ))}
            </div>

            {/* fullscreen previews */}
            {selectedId !== null && (
                <div
                    className={styles.viewerBackdrop}
                    onClick={handleBackdropClick}
                >
                    <div className={styles.viewer} onClick={stopPropagation}>
                        <div className={styles.viewerImageArea}>
                            {/* chevron buttons */}
                            <button
                                type="button"
                                className={`${styles.navButton} ${styles.navButtonLeft}`}
                                onClick={(e) => {
                                    e.stopPropagation();
                                    moveSelection("prev");
                                }}
                                aria-label="Previous image"
                            >
                                ‹
                            </button>
                            <button
                                type="button"
                                className={`${styles.navButton} ${styles.navButtonRight}`}
                                onClick={(e) => {
                                    e.stopPropagation();
                                    moveSelection("next");
                                }}
                                aria-label="Next image"
                            >
                                ›
                            </button>

                            {/* close button */}
                            <button
                                type="button"
                                className={styles.closeButton}
                                onClick={(e) => {
                                    e.stopPropagation();
                                    setSelectedId(null);
                                }}
                                aria-label="Close"
                            >
                                ×
                            </button>

                            {/* zoom toggle button (1x / 2x) */}
                            <button
                                type="button"
                                className={styles.zoomButton}
                                onClick={handleZoomButtonClick}
                                aria-label={
                                    zoom === 1 ? "Zoom to 2x" : "Zoom to 1x"
                                }
                            >
                                {zoom === 1 ? "1×" : "2×"}
                            </button>

                            {/* scrollable image area */}
                            <div
                                className={styles.viewerImageScroll}
                                ref={scrollRef}
                                style={{
                                    cursor: hoveringImage
                                        ? zoom === 1
                                            ? "zoom-in"
                                            : "zoom-out"
                                        : "default",
                                    justifyContent:
                                        zoom === 1 ? "center" : "flex-start",
                                    alignItems:
                                        zoom === 1 ? "center" : "flex-start",
                                }}
                            >
                                <img
                                    src={`http://localhost:3824/thumbnails/${selectedId}?size=2`}
                                    alt={`Preview of image ${selectedId}`}
                                    className={styles.viewerImage}
                                    draggable={false}
                                    onMouseEnter={() => setHoveringImage(true)}
                                    onMouseLeave={() => setHoveringImage(false)}
                                    onClick={handleImageClick}
                                    style={
                                        zoom === 1
                                            ? {}
                                            : {
                                                  width: `${zoom * 100}%`,
                                                  height: "auto",
                                                  maxWidth: "none",
                                                  maxHeight: "none",
                                              }
                                    }
                                />
                            </div>
                        </div>

                        <aside className={styles.viewerSidebar}>
                            <div className={styles.detailsHeaderRow}>
                                <h2 className={styles.sidebarTitle}>
                                    Image details
                                </h2>
                                {metadataLoading && (
                                    <span className={styles.pill}>Loading</span>
                                )}
                            </div>

                            {metadataError && (
                                <div className={styles.errorBanner}>
                                    Failed to load metadata: {metadataError}
                                </div>
                            )}

                            <div className={styles.detailGrid}>
                                <div className={styles.detailRow}>
                                    <div className={styles.detailLabel}>
                                        Filename
                                    </div>
                                    <div className={styles.detailValue}>
                                        {currentMetadata?.fileName ?? "—"}
                                    </div>
                                </div>

                                <div className={styles.detailRow}>
                                    <div className={styles.detailLabel}>
                                        Date taken
                                    </div>
                                    <div className={styles.detailValue}>
                                        {currentMetadata
                                            ? formatDate(
                                                  currentMetadata.dateTaken,
                                              )
                                            : "—"}
                                    </div>
                                </div>

                                <div className={styles.detailRow}>
                                    <div className={styles.detailLabel}>
                                        Date uploaded
                                    </div>
                                    <div className={styles.detailValue}>
                                        {currentMetadata
                                            ? formatDate(
                                                  currentMetadata.dateUploaded,
                                              )
                                            : "—"}
                                    </div>
                                </div>

                                <div className={styles.detailRow}>
                                    <div className={styles.detailLabel}>
                                        Location
                                    </div>
                                    <div className={styles.detailValue}>
                                        {currentMetadata
                                            ? formatLocation(
                                                  currentMetadata.latitude,
                                                  currentMetadata.longitude,
                                              )
                                            : "—"}
                                    </div>
                                </div>

                                <div className={styles.detailRow}>
                                    <div className={styles.detailLabel}>
                                        Camera
                                    </div>
                                    <div className={styles.detailValue}>
                                        {currentMetadata
                                            ? formatCamera(
                                                  currentMetadata.cameraMake,
                                                  currentMetadata.cameraModel,
                                              )
                                            : "—"}
                                    </div>
                                </div>

                                <div className={styles.detailRow}>
                                    <div className={styles.detailLabel}>
                                        Tags
                                    </div>
                                    <div className={styles.detailValue}>
                                        {tagsError ? (
                                            <span className={styles.inlineError}>
                                                {tagsError}
                                            </span>
                                        ) : tagsLoading && currentTags === null ? (
                                            "Loading…"
                                        ) : (
                                            <>
                                                {tagMutateError && (
                                                    <div
                                                        className={
                                                            styles.errorBanner
                                                        }
                                                        style={{
                                                            marginBottom: 10,
                                                        }}
                                                    >
                                                        Tag update failed:{" "}
                                                        {tagMutateError}
                                                    </div>
                                                )}

                                                {tagMutateInfo &&
                                                    !tagMutateError && (
                                                        <div
                                                            className={
                                                                styles.tagInfoBanner
                                                            }
                                                        >
                                                            {tagMutateInfo}
                                                        </div>
                                                    )}

                                                <div className={styles.tagWrap}>
                                                    {(currentTags ?? []).map(
                                                        (t) => (
                                                            <span
                                                                key={t}
                                                                className={
                                                                    styles.tagTokenWrap
                                                                }
                                                            >
                                                                <Pill
                                                                    label={t}
                                                                    variant="token"
                                                                    className={
                                                                        styles.tagPill
                                                                    }
                                                                    onClick={(
                                                                        e,
                                                                    ) => {
                                                                        e.stopPropagation();
                                                                        onSearchTag?.(
                                                                            t,
                                                                        );
                                                                        setSelectedId(
                                                                            null,
                                                                        );
                                                                    }}
                                                                />
                                                                <button
                                                                    className={
                                                                        styles.tagTokenClose
                                                                    }
                                                                    type="button"
                                                                    aria-label={`Remove tag ${t}`}
                                                                    title="Remove"
                                                                    disabled={
                                                                        tagMutating
                                                                    }
                                                                    onMouseDown={(
                                                                        e,
                                                                    ) =>
                                                                        e.preventDefault()
                                                                    }
                                                                    onClick={(
                                                                        e,
                                                                    ) => {
                                                                        e.stopPropagation();
                                                                        removeTagFromSelected(
                                                                            t,
                                                                        );
                                                                    }}
                                                                >
                                                                    ✕
                                                                </button>
                                                            </span>
                                                        ),
                                                    )}

                                                    {!addOpen ? (
                                                        <Pill
                                                            label="+"
                                                            variant="token"
                                                            className={`${styles.tagPill} ${styles.tagAddPill}`}
                                                            onClick={(e) => {
                                                                e.stopPropagation();
                                                                setAddOpen(true);
                                                                setAddValue("");
                                                                setTagSuggestIndex(
                                                                    0,
                                                                );
                                                                setTagSuggestOpen(
                                                                    false,
                                                                ); // <- closed by default
                                                                setTagSuggestPinned(
                                                                    false,
                                                                );
                                                                ensureAllTagsLoaded();
                                                            }}
                                                        />
                                                    ) : (
                                                        <div
                                                            ref={addEditorRef}
                                                            className={
                                                                styles.tagAddEditor
                                                            }
                                                            onClick={(e) =>
                                                                e.stopPropagation()
                                                            }
                                                        >
                                                            <div
                                                                className={
                                                                    styles.tagAddInputWrap
                                                                }
                                                            >
                                                                <input
                                                                    ref={
                                                                        addInputRef
                                                                    }
                                                                    className={
                                                                        styles.tagAddInput
                                                                    }
                                                                    placeholder={
                                                                        allTagsLoading
                                                                            ? "Loading tags…"
                                                                            : "Add tag…"
                                                                    }
                                                                    value={
                                                                        addValue
                                                                    }
                                                                    onChange={(
                                                                        e,
                                                                    ) => {
                                                                        const v =
                                                                            e
                                                                                .target
                                                                                .value;
                                                                        setAddValue(
                                                                            v,
                                                                        );
                                                                        setTagSuggestIndex(
                                                                            0,
                                                                        );

                                                                        const hasNeedle =
                                                                            normalizeTag(
                                                                                v,
                                                                            )
                                                                                .length >
                                                                            0;
                                                                        if (
                                                                            hasNeedle
                                                                        ) {
                                                                            setTagSuggestOpen(
                                                                                true,
                                                                            );
                                                                        } else {
                                                                            setTagSuggestOpen(
                                                                                tagSuggestPinned,
                                                                            );
                                                                        }
                                                                    }}
                                                                    onFocus={() => {
                                                                        const hasNeedle =
                                                                            normalizeTag(
                                                                                addValue,
                                                                            )
                                                                                .length >
                                                                            0;
                                                                        if (
                                                                            hasNeedle ||
                                                                            tagSuggestPinned
                                                                        ) {
                                                                            setTagSuggestOpen(
                                                                                true,
                                                                            );
                                                                        }
                                                                    }}
                                                                    onClick={() => {
                                                                        const hasNeedle =
                                                                            normalizeTag(
                                                                                addValue,
                                                                            )
                                                                                .length >
                                                                            0;
                                                                        if (
                                                                            hasNeedle ||
                                                                            tagSuggestPinned
                                                                        ) {
                                                                            setTagSuggestOpen(
                                                                                true,
                                                                            );
                                                                        }
                                                                    }}
                                                                    onKeyDown={(
                                                                        e,
                                                                    ) => {
                                                                        if (
                                                                            e.key ===
                                                                            "ArrowDown"
                                                                        ) {
                                                                            if (
                                                                                tagSuggestions.length ===
                                                                                0
                                                                            )
                                                                                return;
                                                                            e.preventDefault();
                                                                            setTagSuggestOpen(
                                                                                true,
                                                                            );
                                                                            setTagSuggestIndex(
                                                                                (
                                                                                    i,
                                                                                ) =>
                                                                                    Math.min(
                                                                                        i +
                                                                                            1,
                                                                                        tagSuggestions.length -
                                                                                            1,
                                                                                    ),
                                                                            );
                                                                            return;
                                                                        }

                                                                        if (
                                                                            e.key ===
                                                                            "ArrowUp"
                                                                        ) {
                                                                            if (
                                                                                tagSuggestions.length ===
                                                                                0
                                                                            )
                                                                                return;
                                                                            e.preventDefault();
                                                                            setTagSuggestOpen(
                                                                                true,
                                                                            );
                                                                            setTagSuggestIndex(
                                                                                (
                                                                                    i,
                                                                                ) =>
                                                                                    Math.max(
                                                                                        i -
                                                                                            1,
                                                                                        0,
                                                                                    ),
                                                                            );
                                                                            return;
                                                                        }

                                                                        if (
                                                                            e.key ===
                                                                            "Enter"
                                                                        ) {
                                                                            e.preventDefault();
                                                                            if (
                                                                                tagSuggestOpen &&
                                                                                tagSuggestions[
                                                                                    tagSuggestIndex
                                                                                ]
                                                                            ) {
                                                                                addTagToSelected(
                                                                                    tagSuggestions[
                                                                                        tagSuggestIndex
                                                                                    ],
                                                                                );
                                                                            } else {
                                                                                addTagToSelected();
                                                                            }
                                                                            setTagSuggestOpen(
                                                                                false,
                                                                            );
                                                                            setTagSuggestPinned(
                                                                                false,
                                                                            );
                                                                        }

                                                                        // No Escape handling here on purpose:
                                                                        // Escape should close the preview window.
                                                                    }}
                                                                    aria-label="New tag"
                                                                    aria-expanded={
                                                                        showTagDropdown
                                                                    }
                                                                    aria-controls="tag-suggest-dropdown"
                                                                />

                                                                <button
                                                                    type="button"
                                                                    className={
                                                                        styles.tagSuggestToggle
                                                                    }
                                                                    aria-label={
                                                                        tagSuggestPinned
                                                                            ? "Hide tag suggestions"
                                                                            : "Show tag suggestions"
                                                                    }
                                                                    aria-pressed={
                                                                        tagSuggestPinned
                                                                    }
                                                                    onMouseDown={(
                                                                        e,
                                                                    ) =>
                                                                        e.preventDefault()
                                                                    }
                                                                    onClick={async () => {
                                                                        await toggleSuggestions();
                                                                    }}
                                                                    title={
                                                                        tagSuggestPinned
                                                                            ? "Hide suggestions"
                                                                            : "Show suggestions"
                                                                    }
                                                                >
                                                                    ...
                                                                </button>

                                                                {showTagDropdown && (
                                                                    <div
                                                                        id="tag-suggest-dropdown"
                                                                        className={
                                                                            styles.tagSuggestDropdown
                                                                        }
                                                                        role="listbox"
                                                                        aria-label="Tag suggestions"
                                                                    >
                                                                        {allTagsLoading ? (
                                                                            <div
                                                                                className={
                                                                                    styles.tagSuggestLoading
                                                                                }
                                                                            >
                                                                                Loading…
                                                                            </div>
                                                                        ) : (
                                                                            tagSuggestions.map(
                                                                                (
                                                                                    t,
                                                                                    idx,
                                                                                ) => (
                                                                                    <button
                                                                                        key={
                                                                                            t
                                                                                        }
                                                                                        type="button"
                                                                                        className={`${styles.tagSuggestItem} ${
                                                                                            idx ===
                                                                                            tagSuggestIndex
                                                                                                ? styles.tagSuggestActive
                                                                                                : ""
                                                                                        }`}
                                                                                        onMouseDown={(
                                                                                            ev,
                                                                                        ) =>
                                                                                            ev.preventDefault()
                                                                                        }
                                                                                        onMouseEnter={() =>
                                                                                            setTagSuggestIndex(
                                                                                                idx,
                                                                                            )
                                                                                        }
                                                                                        onClick={() => {
                                                                                            addTagToSelected(
                                                                                                t,
                                                                                            );
                                                                                            setTagSuggestOpen(
                                                                                                false,
                                                                                            );
                                                                                            setTagSuggestPinned(
                                                                                                false,
                                                                                            );
                                                                                        }}
                                                                                    >
                                                                                        <span
                                                                                            className={
                                                                                                styles.tagSuggestLabel
                                                                                            }
                                                                                        >
                                                                                            {t}
                                                                                        </span>
                                                                                    </button>
                                                                                ),
                                                                            )
                                                                        )}
                                                                    </div>
                                                                )}
                                                            </div>

                                                            <button
                                                                type="button"
                                                                className={
                                                                    styles.tagAddBtn
                                                                }
                                                                onMouseDown={(
                                                                    e,
                                                                ) =>
                                                                    e.preventDefault()
                                                                }
                                                                onClick={() => {
                                                                    addTagToSelected();
                                                                    setTagSuggestOpen(
                                                                        false,
                                                                    );
                                                                    setTagSuggestPinned(
                                                                        false,
                                                                    );
                                                                }}
                                                                disabled={
                                                                    tagMutating ||
                                                                    normalizeTag(
                                                                        addValue,
                                                                    ).length ===
                                                                        0
                                                                }
                                                                title="Add"
                                                            >
                                                                Add
                                                            </button>

                                                            <button
                                                                type="button"
                                                                className={
                                                                    styles.tagCancelBtn
                                                                }
                                                                onMouseDown={(
                                                                    e,
                                                                ) =>
                                                                    e.preventDefault()
                                                                }
                                                                onClick={() => {
                                                                    setAddOpen(
                                                                        false,
                                                                    );
                                                                    setAddValue(
                                                                        "",
                                                                    );
                                                                    setTagSuggestOpen(
                                                                        false,
                                                                    );
                                                                    setTagSuggestPinned(
                                                                        false,
                                                                    );
                                                                }}
                                                                disabled={
                                                                    tagMutating
                                                                }
                                                                title="Cancel"
                                                            >
                                                                Cancel
                                                            </button>
                                                        </div>
                                                    )}
                                                </div>
                                            </>
                                        )}
                                    </div>
                                </div>
                            </div>
                        </aside>
                    </div>
                </div>
            )}
        </>
    );
}