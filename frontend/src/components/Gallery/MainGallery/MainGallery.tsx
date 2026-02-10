"use client";

import {
    useEffect,
    useState,
    useRef,
    MouseEvent as ReactMouseEvent,
    KeyboardEvent as ReactKeyboardEvent,
    useCallback,
} from "react";
import styles from "./MainGallery.module.css";
import GalleryImage from "../GalleryImage/GalleryImage";
import MetadataGetter, { ImageMetadata } from "@/api/metadata/metadatagetter";
import TagGetter from "@/api/tags/taggetter";
import Pill from "@/components/UI/Pill/Pill";
import { useCachedById } from "@/common/gallery/cache";
import {
    formatDate,
    formatLocation,
    formatCamera,
} from "@/common/gallery/format";
import { useTagEditor } from "@/common/gallery/tageditor";

export type GalleryProps = {
    image_ids: number[];
    onSearchTag?: (tag: string) => void;
};

// point (fx, fy) on image expressed as fraction of width/height (after zoom)
type PendingPan = null | { fx: number; fy: number };

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

    // metadata + tags: cached per image id
    const metadata = useCachedById<ImageMetadata>(selectedId, MetadataGetter);
    const tags = useCachedById<string[]>(selectedId, TagGetter);

    const metadataLoading = metadata.loading;
    const metadataError = metadata.error.some ? metadata.error.value : null;
    const currentMetadata: ImageMetadata | null = metadata.current.some
        ? metadata.current.value
        : null;

    const tagsById = tags.byId;
    const setTagsById = tags.setById;
    const tagsLoading = tags.loading;
    const tagsError = tags.error.some ? tags.error.value : null;

    const currentTags: string[] | null = tags.current.some
        ? tags.current.value
        : null;

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

    const handleZoomButtonClick = (
        event: ReactMouseEvent<HTMLButtonElement>,
    ) => {
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

    const {
        tagMutating,
        tagMutateError,
        tagMutateInfo,

        addOpen,
        addValue,
        addInputRef,
        addEditorRef,

        allTagsLoading,
        tagSuggestPinned,
        tagSuggestIndex,

        tagSuggestions,
        showTagDropdown,

        openAddEditor,
        closeAddEditor,

        removeTagFromSelected,
        addTagToSelected,

        toggleSuggestions,

        onAddValueChange,
        onAddInputFocusOrClick,
        onAddInputKeyDown,

        setTagSuggestIndex,
    } = useTagEditor({
        selectedId,
        tagsById,
        setTagsById,
        currentTags,
    });

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
                                            <span
                                                className={styles.inlineError}
                                            >
                                                {tagsError}
                                            </span>
                                        ) : tagsLoading &&
                                          currentTags === null ? (
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
                                                                        void removeTagFromSelected(
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
                                                                void openAddEditor();
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
                                                                role="combobox"
                                                                aria-label="New tag"
                                                                aria-haspopup="listbox"
                                                                aria-expanded={
                                                                    showTagDropdown
                                                                }
                                                                aria-controls="tag-suggest-dropdown"
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
                                                                            ? "Loading tags..."
                                                                            : "Add tag..."
                                                                    }
                                                                    value={
                                                                        addValue
                                                                    }
                                                                    onChange={(
                                                                        e,
                                                                    ) =>
                                                                        onAddValueChange(
                                                                            e
                                                                                .target
                                                                                .value,
                                                                        )
                                                                    }
                                                                    onFocus={
                                                                        onAddInputFocusOrClick
                                                                    }
                                                                    onClick={
                                                                        onAddInputFocusOrClick
                                                                    }
                                                                    onKeyDown={(
                                                                        e,
                                                                    ) => {
                                                                        if (
                                                                            e.key ===
                                                                                "ArrowDown" ||
                                                                            e.key ===
                                                                                "ArrowUp" ||
                                                                            e.key ===
                                                                                "Enter"
                                                                        ) {
                                                                            e.preventDefault();
                                                                            void onAddInputKeyDown(
                                                                                e.key,
                                                                            );
                                                                        }
                                                                    }}
                                                                    aria-autocomplete="list"
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
                                                                                        className={`${
                                                                                            styles.tagSuggestItem
                                                                                        } ${
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
                                                                                            void addTagToSelected(
                                                                                                t,
                                                                                            );
                                                                                        }}
                                                                                    >
                                                                                        <span
                                                                                            className={
                                                                                                styles.tagSuggestLabel
                                                                                            }
                                                                                        >
                                                                                            {
                                                                                                t
                                                                                            }
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
                                                                    void addTagToSelected();
                                                                }}
                                                                disabled={
                                                                    tagMutating ||
                                                                    addValue.trim()
                                                                        .length ===
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
                                                                onClick={
                                                                    closeAddEditor
                                                                }
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
