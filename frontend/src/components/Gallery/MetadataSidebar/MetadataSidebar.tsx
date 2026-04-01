"use client";

import {
    useEffect,
    useLayoutEffect,
    useRef,
    useState,
    useCallback,
    memo,
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
    formatLocation,
    formatResultion,
} from "@/common/gallery/format";
import { formatLongDate } from "@/common/dateformat";
import { useTagEditor } from "@/common/gallery/tageditor";
import FileDownloader from "@/api/files/filedownloader";
import FileDeleter from "@/api/files/filedeleter";
import {
    MoreHorizontal,
    PlusCircle,
    Plus,
    X,
    XCircle,
    Calendar,
    CloudUpload,
    MapPin,
    Camera,
    Tag,
    FullscreenIcon,
    Download,
    Trash2,
    Loader2,
    PuzzleIcon,
} from "lucide-react";
import ManualRunMenu from "@/components/Plugins/ManualRunMenu/ManualRunMenu";
import FormPanel from "@/components/Plugins/FormPanel/FormPanel";
import PluginRunner from "@/api/plugins/pluginrunner";
import JobsDetailer, { JobsDetailResult } from "@/api/jobs/jobsdetailer";
import { useRouter } from "next/navigation";

// memoized so individual pills don't re-render when unrelated sidebar state changes
type TagTokenProps = {
    tag: string;
    disabled: boolean;
    onSearch?: (tag: string) => void;
    onRemove: (tag: string) => Promise<void>;
};
const TagToken = memo(function TagToken({
    tag,
    disabled,
    onSearch,
    onRemove,
}: TagTokenProps) {
    const handleSearch = useCallback(
        (e: ReactMouseEvent<HTMLButtonElement>) => {
            e.stopPropagation();
            onSearch?.(tag);
        },
        [tag, onSearch],
    );
    const handleRemove = useCallback(
        (e: ReactMouseEvent<HTMLButtonElement>) => {
            e.stopPropagation();
            void onRemove(tag);
        },
        [tag, onRemove],
    );
    return (
        <span className={styles.tagTokenWrap}>
            <Pill
                label={tag}
                variant="token"
                className={styles.tagPill}
                onClick={handleSearch}
            />
            <button
                className={styles.tagTokenClose}
                type="button"
                aria-label={`Remove tag ${tag}`}
                title="Remove"
                disabled={disabled}
                onMouseDown={(e) => e.preventDefault()}
                onClick={handleRemove}
            >
                <X className={styles.iconSm} />
            </button>
        </span>
    );
});

export type MetadataSidebarProps = {
    selectedId: number;
    onSearchTag?: (tag: string) => void;
    viewerRef: RefObject<HTMLDivElement | null>;

    // optional: lets the parent close the preview after deletion
    onDeleted?: () => void;
};

async function requestErrorToMessage(err: {
    status_code: number;
    response: Response;
}): Promise<string> {
    try {
        const text = await err.response.text();
        if (!text) return err.response.statusText || `HTTP ${err.status_code}`;

        try {
            const json = JSON.parse(text) as unknown;
            if (
                json &&
                typeof json === "object" &&
                "error" in json &&
                typeof (json as { error: unknown }).error === "string"
            ) {
                return (json as { error: string }).error;
            }
        } catch {
            // ignore
        }

        return text;
    } catch {
        return err.status_code === 0
            ? "Network error"
            : `HTTP ${err.status_code}`;
    }
}

function triggerBrowserDownload(blob: Blob, filename: string) {
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    a.remove();
    URL.revokeObjectURL(url);
}

const MetadataSidebar = memo(function MetadataSidebar({
    selectedId,
    onSearchTag,
    viewerRef,
    onDeleted,
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

    const [fileBusy, setFileBusy] = useState<null | "download" | "delete">(
        null,
    );
    const [fileError, setFileError] = useState<string | null>(null);

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

    const tagPillsScrollRef = useRef<HTMLDivElement>(null);
    const [tagPillsFade, setTagPillsFade] = useState({
        top: false,
        bottom: false,
    });
    const syncRafRef = useRef<number | null>(null);

    // schedules at most one RAF per frame so rapid scroll events don't queue up multiple state updates
    const syncTagPillsFade = useCallback(() => {
        if (syncRafRef.current !== null)
            cancelAnimationFrame(syncRafRef.current);
        syncRafRef.current = requestAnimationFrame(() => {
            syncRafRef.current = null;
            const el = tagPillsScrollRef.current;
            if (!el) {
                setTagPillsFade({ top: false, bottom: false });
                return;
            }
            const overflow = el.scrollHeight > el.clientHeight + 1;
            if (!overflow) {
                setTagPillsFade({ top: false, bottom: false });
                return;
            }
            const atTop = el.scrollTop <= 1;
            const atBottom =
                el.scrollTop + el.clientHeight >= el.scrollHeight - 1;
            setTagPillsFade((prev) => {
                const next = { top: !atTop, bottom: !atBottom };
                return prev.top === next.top && prev.bottom === next.bottom
                    ? prev
                    : next;
            });
        });
    }, []);

    useEffect(() => {
        const el = tagPillsScrollRef.current;
        if (!el) return;
        syncTagPillsFade();
        const ro = new ResizeObserver(() => syncTagPillsFade());
        ro.observe(el);
        return () => {
            if (syncRafRef.current !== null) {
                cancelAnimationFrame(syncRafRef.current);
                syncRafRef.current = null;
            }
            ro.disconnect();
        };
    }, [currentTags, syncTagPillsFade]);

    const stopClick = (e: ReactMouseEvent<HTMLElement>) => e.stopPropagation();

    const downloadBusy = fileBusy === "download";
    const deleteBusy = fileBusy === "delete";
    const anyBusy = fileBusy !== null;

    const handleDownload = async (e: ReactMouseEvent<HTMLButtonElement>) => {
        e.stopPropagation();
        if (anyBusy) return;

        setFileError(null);
        setFileBusy("download");

        const res = await FileDownloader(selectedId)();
        if (res.ok) {
            triggerBrowserDownload(res.value.blob, res.value.filename);
        } else {
            setFileError(await requestErrorToMessage(res.error));
        }

        setFileBusy(null);
    };

    const handleDelete = async (e: ReactMouseEvent<HTMLButtonElement>) => {
        e.stopPropagation();
        if (anyBusy) return;

        const confirmed = window.confirm(
            "Delete this image? This cannot be undone.",
        );
        if (!confirmed) return;

        setFileError(null);
        setFileBusy("delete");

        const res = await FileDeleter(selectedId)();
        if (res.ok) {
            onDeleted?.();
        } else {
            setFileError(await requestErrorToMessage(res.error));
        }

        setFileBusy(null);
    };

    const router = useRouter();
    const [selectingPlugin, setSelectingPlugin] = useState<boolean>(false);
    const [answeringForm, setAnsweringForm] = useState<string | undefined>(
        undefined,
    );
    const [answeringPlugin, setAnsweringPlugin] = useState<string>("");
    const [answeringHook, setAnsweringHook] = useState<string>("");

    const pluginRunner = async (
        plugin_name: string,
        hook_name: string,
        _hook_signature: string,
        hook_type: string,
        form_receive_hook_name?: string,
    ) => {
        setSelectingPlugin(false);
        const res_promise = PluginRunner(plugin_name, hook_name, selectedId);
        if (hook_type !== "generate_form") {
            router.replace("/jobs");
            return;
        }
        setAnsweringForm("<form><p>Waiting for form to load...</p></form>");

        if (!form_receive_hook_name) {
            console.error(
                "Should have returned a hook that we can give the form data to!",
            );
            return;
        }

        const res = await res_promise;
        if (!res.ok) {
            console.error(res.error);
            return;
        }

        const uuid = res.value;
        let job_output: JobsDetailResult | undefined = undefined;
        while (true) {
            job_output = await JobsDetailer(uuid);
            if (!job_output.ok) {
                console.error(job_output.error);
                return;
            }
            if (job_output.value.completed?.length === 1) {
                break;
            }
            // Wait 1 second before asking again
            await new Promise((r) => setTimeout(r, 1000));
        }

        const form: string | undefined = job_output.value.completed[0]?.form;
        if (!form) {
            console.error(`Did not get form out of plugin! Job UUID: ${uuid}`);
            return;
        }
        if (!form.startsWith("<form>") || !form.endsWith("</form>")) {
            console.error(`Did not get valid form! Form received: ${form}`);
            return;
        }
        setAnsweringForm(form);
        setAnsweringPlugin(plugin_name);
        setAnsweringHook(form_receive_hook_name);
    };

    return (
        <aside className={styles.viewerSidebar}>
            <div className={styles.detailsHeaderRow}>
                <h2 className={styles.sidebarTitle}>Image details</h2>

                <div className={styles.headerRight} onClick={stopClick}>
                    <button
                        type="button"
                        className={styles.headerIconButton}
                        onClick={handleDownload}
                        disabled={anyBusy}
                        aria-label="Download image"
                        title="Download"
                    >
                        {downloadBusy ? (
                            <Loader2
                                className={`${styles.icon} ${styles.spinning}`}
                            />
                        ) : (
                            <Download className={styles.icon} />
                        )}
                    </button>

                    <button
                        type="button"
                        className={`${styles.headerIconButton} ${styles.dangerButton}`}
                        onClick={handleDelete}
                        disabled={anyBusy}
                        aria-label="Delete image"
                        title="Delete"
                    >
                        {deleteBusy ? (
                            <Loader2
                                className={`${styles.icon} ${styles.spinning}`}
                            />
                        ) : (
                            <Trash2 className={styles.icon} />
                        )}
                    </button>

                    {metadataLoading && (
                        <span className={styles.pill}>Loading</span>
                    )}
                </div>
            </div>

            {fileError && (
                <div className={styles.errorBanner}>
                    File action failed: {fileError}
                </div>
            )}

            {metadataError && (
                <div className={styles.errorBanner}>
                    Failed to load metadata: {metadataError}
                </div>
            )}

            <div className={styles.detailGrid}>
                <div className={styles.detailRow}>
                    <div className={styles.detailLabelRow}>
                        <FullscreenIcon className={styles.detailLabelIcon} />
                        <span className={styles.detailLabel}>Resolution</span>
                    </div>
                    <div className={styles.detailValue}>
                        {currentMetadata
                            ? formatResultion(
                                  currentMetadata.width,
                                  currentMetadata.height,
                              )
                            : "—"}
                    </div>
                </div>

                <div className={styles.detailRow}>
                    <div className={styles.detailLabelRow}>
                        <Calendar className={styles.detailLabelIcon} />
                        <span className={styles.detailLabel}>Date taken</span>
                    </div>
                    <div className={styles.detailValue}>
                        {currentMetadata
                            ? formatLongDate(currentMetadata.dateTaken, {
                                  timeZone: "UTC",
                              })
                            : "—"}
                    </div>
                </div>

                <div className={styles.detailRow}>
                    <div className={styles.detailLabelRow}>
                        <CloudUpload className={styles.detailLabelIcon} />
                        <span className={styles.detailLabel}>
                            Date uploaded
                        </span>
                    </div>
                    <div className={styles.detailValue}>
                        {currentMetadata
                            ? formatLongDate(currentMetadata.dateUploaded)
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
                        <Tag className={styles.detailLabelIcon} />
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
                                    <div
                                        className={styles.tagPillsFadeWrap}
                                        data-fade-top={
                                            tagPillsFade.top ? "1" : "0"
                                        }
                                        data-fade-bottom={
                                            tagPillsFade.bottom ? "1" : "0"
                                        }
                                    >
                                        <div
                                            ref={tagPillsScrollRef}
                                            className={styles.tagPillsWrap}
                                            onScroll={syncTagPillsFade}
                                        >
                                            {(currentTags ?? []).map((t) => (
                                                <TagToken
                                                    key={t}
                                                    tag={t}
                                                    disabled={tagMutating}
                                                    onSearch={onSearchTag}
                                                    onRemove={
                                                        removeTagFromSelected
                                                    }
                                                />
                                            ))}
                                        </div>
                                    </div>

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
                                                className={`${styles.tagActionBtn}`}
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
                                                className={`${styles.tagActionBtn}`}
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

                <div className={styles.detailRow}>
                    <div className={styles.detailLabelRow}>
                        <PuzzleIcon className={styles.detailLabelIcon} />
                        <span className={styles.detailLabel}>Plugins</span>
                    </div>
                    <div className={styles.detailValue}>
                        <button
                            type="button"
                            className={`${styles.plugin_run_button}`}
                            onMouseDown={(e) => e.preventDefault()}
                            onClick={(e) => {
                                e.stopPropagation();
                                setSelectingPlugin(true);
                            }}
                            aria-label="Run Plugins"
                            title="Run Plugins"
                        >
                            Run Plugins
                        </button>
                    </div>
                </div>
            </div>
            {selectingPlugin && (
                <ManualRunMenu
                    onClose={() => setSelectingPlugin(false)}
                    onPluginChosen={pluginRunner}
                    multipleImages={false}
                />
            )}

            {answeringForm && (
                <FormPanel
                    onClose={() => setAnsweringForm(undefined)}
                    onFormSubmit={() => {
                        setAnsweringForm(undefined);
                        router.replace("/jobs");
                    }}
                    formString={answeringForm}
                    plugin={answeringPlugin}
                    hook={answeringHook}
                />
            )}
        </aside>
    );
});

export default MetadataSidebar;
