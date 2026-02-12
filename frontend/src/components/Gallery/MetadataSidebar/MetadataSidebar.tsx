"use client";

import {
    useLayoutEffect,
    useRef,
    useState,
    RefObject,
    MouseEvent as ReactMouseEvent,
} from "react";
import styles from "./MetadataSidebar.module.css";
import MetadataGetter, { ImageMetadata } from "@/api/metadata/metadatagetter";
import TagGetter from "@/api/tags/taggetter";
import Pill from "@/components/UI/Pill/Pill";
import { useCachedById } from "@/common/gallery/cache";
import {
    formatCamera,
    formatDate,
    formatLocation,
} from "@/common/gallery/format";
import { useTagEditor } from "@/common/gallery/tageditor";
import {
    MoreHorizontal,
    PlusCircle,
    Plus,
    X,
    XCircle,
    FileText,
    Calendar,
    Upload,
    MapPin,
    Camera,
    Tags,
} from "lucide-react";

export type MetadataSidebarProps = {
    selectedId: number;
    onSearchTag?: (tag: string) => void;
    viewerRef: RefObject<HTMLDivElement | null>;
};

export default function MetadataSidebar({
    selectedId,
    onSearchTag,
    viewerRef,
}: MetadataSidebarProps) {
    const comboRef = useRef<HTMLDivElement | null>(null);
    const [tagDropdownMaxHeight, setTagDropdownMaxHeight] =
        useState<number>(240);

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

    useLayoutEffect(() => {
        if (!showTagDropdown) return;

        const compute = () => {
            const viewerEl = viewerRef.current;
            const comboEl = comboRef.current;
            if (!viewerEl || !comboEl) return;

            const viewerRect = viewerEl.getBoundingClientRect();
            const comboRect = comboEl.getBoundingClientRect();
            const dropdownTop = comboRect.bottom + 8;
            const bottomPad = 12;
            const available = viewerRect.bottom - dropdownTop - bottomPad;

            const clamped = Math.max(120, Math.min(240, Math.floor(available)));
            setTagDropdownMaxHeight(clamped);
        };

        compute();
        window.addEventListener("resize", compute);
        return () => window.removeEventListener("resize", compute);
    }, [showTagDropdown, viewerRef]);

    const stopClick = (e: ReactMouseEvent<HTMLElement>) => e.stopPropagation();

    return (
        <aside className={styles.viewerSidebar}>
            <div className={styles.detailsHeaderRow}>
                <h2 className={styles.sidebarTitle}>Image details</h2>
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
                    <div className={styles.detailLabelRow}>
                        <FileText className={styles.detailLabelIcon} />
                        <span className={styles.detailLabel}>Filename</span>
                    </div>
                    <div className={styles.detailValue}>
                        {currentMetadata?.fileName ?? "—"}
                    </div>
                </div>

                <div className={styles.detailRow}>
                    <div className={styles.detailLabelRow}>
                        <Calendar className={styles.detailLabelIcon} />
                        <span className={styles.detailLabel}>Date taken</span>
                    </div>
                    <div className={styles.detailValue}>
                        {currentMetadata
                            ? formatDate(currentMetadata.dateTaken)
                            : "—"}
                    </div>
                </div>

                <div className={styles.detailRow}>
                    <div className={styles.detailLabelRow}>
                        <Upload className={styles.detailLabelIcon} />
                        <span className={styles.detailLabel}>
                            Date uploaded
                        </span>
                    </div>
                    <div className={styles.detailValue}>
                        {currentMetadata
                            ? formatDate(currentMetadata.dateUploaded)
                            : "—"}
                    </div>
                </div>

                <div className={styles.detailRow}>
                    <div className={styles.detailLabelRow}>
                        <MapPin className={styles.detailLabelIcon} />
                        <span className={styles.detailLabel}>Location</span>
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
                    <div className={styles.detailLabelRow}>
                        <Camera className={styles.detailLabelIcon} />
                        <span className={styles.detailLabel}>Camera</span>
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
                    <div className={styles.detailLabelRow}>
                        <Tags className={styles.detailLabelIcon} />
                        <span className={styles.detailLabel}>Tags</span>
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
                                        className={styles.errorBanner}
                                        style={{ marginBottom: 10 }}
                                    >
                                        Tag update failed: {tagMutateError}
                                    </div>
                                )}

                                {tagMutateInfo && !tagMutateError && (
                                    <div className={styles.tagInfoBanner}>
                                        {tagMutateInfo}
                                    </div>
                                )}

                                <div className={styles.tagWrap}>
                                    {(currentTags ?? []).map((t) => (
                                        <span
                                            key={t}
                                            className={styles.tagTokenWrap}
                                        >
                                            <Pill
                                                label={t}
                                                variant="token"
                                                className={styles.tagPill}
                                                onClick={(e) => {
                                                    e.stopPropagation();
                                                    onSearchTag?.(t);
                                                }}
                                            />
                                            <button
                                                className={styles.tagTokenClose}
                                                type="button"
                                                aria-label={`Remove tag ${t}`}
                                                title="Remove"
                                                disabled={tagMutating}
                                                onMouseDown={(e) =>
                                                    e.preventDefault()
                                                }
                                                onClick={(e) => {
                                                    e.stopPropagation();
                                                    void removeTagFromSelected(
                                                        t,
                                                    );
                                                }}
                                            >
                                                <X className={styles.iconSm} />
                                            </button>
                                        </span>
                                    ))}

                                    {!addOpen ? (
                                        <button
                                            type="button"
                                            className={`${styles.tagAddIconPill} ${styles.tagAddPill}`}
                                            onMouseDown={(e) =>
                                                e.preventDefault()
                                            }
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                void openAddEditor();
                                            }}
                                            aria-label="Add tag"
                                            title="Add tag"
                                        >
                                            <Plus className={styles.iconSm} />
                                        </button>
                                    ) : (
                                        <div
                                            ref={addEditorRef}
                                            className={styles.tagAddEditor}
                                            onClick={stopClick}
                                        >
                                            <div
                                                ref={comboRef}
                                                className={
                                                    styles.tagAddInputWrap
                                                }
                                                role="combobox"
                                                aria-label="New tag"
                                                aria-haspopup="listbox"
                                                aria-expanded={showTagDropdown}
                                                aria-controls="tag-suggest-dropdown"
                                            >
                                                <input
                                                    ref={addInputRef}
                                                    className={
                                                        styles.tagAddInput
                                                    }
                                                    placeholder={
                                                        allTagsLoading
                                                            ? "Loading tags..."
                                                            : "Add tag..."
                                                    }
                                                    value={addValue}
                                                    onChange={(e) =>
                                                        onAddValueChange(
                                                            e.target.value,
                                                        )
                                                    }
                                                    onFocus={
                                                        onAddInputFocusOrClick
                                                    }
                                                    onClick={
                                                        onAddInputFocusOrClick
                                                    }
                                                    onKeyDown={(e) => {
                                                        if (
                                                            e.key ===
                                                                "ArrowDown" ||
                                                            e.key ===
                                                                "ArrowUp" ||
                                                            e.key === "Enter"
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
                                                    onMouseDown={(e) =>
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
                                                    <MoreHorizontal
                                                        className={styles.icon}
                                                    />
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
                                                        <div
                                                            className={
                                                                styles.tagSuggestScroll
                                                            }
                                                            style={{
                                                                maxHeight:
                                                                    tagDropdownMaxHeight,
                                                            }}
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
                                                    </div>
                                                )}
                                            </div>

                                            {/* icon-only add */}
                                            <button
                                                type="button"
                                                className={`${styles.tagActionBtn} ${styles.tagAddBtn}`}
                                                onMouseDown={(e) =>
                                                    e.preventDefault()
                                                }
                                                onClick={() => {
                                                    void addTagToSelected();
                                                }}
                                                disabled={
                                                    tagMutating ||
                                                    addValue.trim().length === 0
                                                }
                                                aria-label="Add tag"
                                                title="Add"
                                            >
                                                <PlusCircle
                                                    className={styles.icon}
                                                />
                                            </button>

                                            {/* icon-only cancel */}
                                            <button
                                                type="button"
                                                className={`${styles.tagActionBtn} ${styles.tagCancelBtn}`}
                                                onMouseDown={(e) =>
                                                    e.preventDefault()
                                                }
                                                onClick={closeAddEditor}
                                                disabled={tagMutating}
                                                aria-label="Cancel"
                                                title="Cancel"
                                            >
                                                <XCircle
                                                    className={styles.icon}
                                                />
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
    );
}
