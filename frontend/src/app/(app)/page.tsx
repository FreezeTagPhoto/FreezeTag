"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { CheckCheck, X, Puzzle, Trash2 } from "lucide-react";
import SearchHandler from "@/api/query/searchhandler";
import TagGetter from "@/api/tags/taggetter";
import styles from "../page.module.css";
import MainGallery from "@/components/Gallery/MainGallery/MainGallery";
import TopBar from "@/components/TopBar/TopBar";
import { addTagToQuery } from "@/common/search/addtagtoquery";
import TagCounter from "@/api/tags/tagcounter";
import MassTaggingGallery from "@/components/Gallery/MassTaggingGallery/MassTaggingGallery";
import TagChangeButton from "@/components/UI/TagChangeButton/TagChangeButton";
import FileDeleter from "@/api/files/filedeleter";
import FreezeTagSet from "@/common/freezetag/freezetagset";
import PluginRunner from "@/api/plugins/pluginrunner";
import ManualRunMenu from "@/components/Plugins/ManualRunMenu/ManualRunMenu";

type TagInfo = { name: string; count?: number };

function normalizeUserTail(s: string): string {
    return s
        .trim()
        .replace(/^\s*;\s*/, "")
        .replace(/\s*;\s*$/, "")
        .trim();
}

function buildQuery(
    sortBy: string,
    sortOrder: string,
    searchTerm: string,
): string {
    const tail = normalizeUserTail(searchTerm);
    return tail
        ? `sortBy=${sortBy};sortOrder=${sortOrder};${tail}`
        : `sortBy=${sortBy};sortOrder=${sortOrder}`;
}

export default function Home() {
    const router = useRouter();
    const searchParams = useSearchParams();

    const [imageIds, setImageIds] = useState<number[]>([]);
    const [searchTerm, setSearchTerm] = useState("");
    const [sortBy, setSortBy] = useState("DateAdded");
    const [sortOrder, setSortOrder] = useState("DESC");

    const [allTags, setAllTags] = useState<string[]>([]);
    const [tagCounts, setTagCounts] = useState<Record<string, number>>({});
    const lastAppliedQRef = useRef<string | null>(null);

    const [multiSelect, setMultiSelect] = useState<boolean>(false);
    const [selectedIds, setSelectedIds] = useState<FreezeTagSet<number>>(
        new FreezeTagSet(),
    );
    const [selectingPlugin, setSelectingPlugin] = useState<boolean>(false);

    useEffect(() => {
        const q = searchParams.get("q");
        if (!q) return;

        if (lastAppliedQRef.current === q) return;
        lastAppliedQRef.current = q;

        setSearchTerm(q);

        router.replace("/", { scroll: false });
    }, [searchParams, router]);

    const query = useMemo(
        () => buildQuery(sortBy, sortOrder, searchTerm),
        [sortBy, sortOrder, searchTerm],
    );

    const reqIdRef = useRef(0);

    // Load images when the query changes
    useEffect(() => {
        let cancelled = false;
        const myReqId = ++reqIdRef.current;

        (async () => {
            const result = await SearchHandler(query);

            if (cancelled) return;
            if (myReqId !== reqIdRef.current) return;

            if (result.ok) {
                setImageIds(result.value);
            } else {
                console.error(
                    `Got ${result.error.status} from backend with message ${result.error.message}`,
                );
            }
        })();

        return () => {
            cancelled = true;
        };
    }, [query]);

    // Load all tags for the top bar
    useEffect(() => {
        let cancelled = false;

        (async () => {
            const res = await TagGetter();
            if (cancelled) return;
            if (res.ok) setAllTags(res.value);
            else {
                console.error("Failed to load tags:", res.error.message);
                setAllTags([]);
            }
        })();

        return () => {
            cancelled = true;
        };
    }, []);

    useEffect(() => {
        const onTagCreated = (e: Event) => {
            const tag = (e as CustomEvent<{ tag: string }>).detail?.tag;
            if (!tag) return;
            setAllTags((prev) =>
                prev.includes(tag) ? prev : [...prev, tag],
            );
            setTagCounts((prev) => ({
                ...prev,
                [tag]: (prev[tag] ?? 0) + 1,
            }));
        };
        window.addEventListener("freezetag:tag-created", onTagCreated);
        return () =>
            window.removeEventListener("freezetag:tag-created", onTagCreated);
    }, []);

    // Load tag counts for the currently visible images
    useEffect(() => {
        let cancelled = false;

        (async () => {
            const counts_result = await TagCounter(query);
            if (cancelled) return;
            if (counts_result.ok) {
                setTagCounts(counts_result.value);
            } else {
                setTagCounts({});
                console.error("Tag Counter did not work!");
            }
        })();

        return () => {
            cancelled = true;
        };
    }, [query]);

    // Prepare tags for the top bar
    const tagsForTopBar: TagInfo[] = useMemo(
        () => allTags.map((name) => ({ name, count: tagCounts[name] ?? 0 })),
        [allTags, tagCounts],
    );

    return (
        <>
            <TopBar
                searchTerm={searchTerm}
                onSearchTermChange={setSearchTerm}
                sortBy={sortBy}
                onSortByChange={setSortBy}
                sortOrder={sortOrder}
                onSortOrderChange={setSortOrder}
                multiSelect={multiSelect}
                onMultiSelectChange={(value) => {
                    setMultiSelect(value);
                    setSelectedIds(new FreezeTagSet());
                }}
                tags={tagsForTopBar}
            />

            <main
                className={`${styles.main} ${multiSelect ? styles.main_select_mode : ""}`}
            >
                <header className={styles.headerRow}>
                    <div>
                        <h1 className={styles.h1}>
                            {multiSelect ? "Selecting Images" : "Gallery"}
                        </h1>
                        <p className={styles.subtle}>
                            {imageIds.length}{" "}
                            {imageIds.length !== 1 ? "images" : "image"}
                        </p>
                    </div>
                    <div className={styles.pillsRow} />
                </header>

                {multiSelect ? (
                    <div className={styles.page}>
                        <div className={styles.gallery_tags_container}>
                            <div className={styles.gallery_select_container}>
                                <div className={styles.select_container}>
                                    <button
                                        type="button"
                                        className={styles.select_button}
                                        onClick={() =>
                                            setSelectedIds(
                                                new FreezeTagSet(imageIds),
                                            )
                                        }
                                    >
                                        <CheckCheck className={styles.select_button_icon} aria-hidden="true" />
                                        Select All
                                    </button>
                                    <button
                                        type="button"
                                        className={styles.select_button}
                                        onClick={() =>
                                            setSelectedIds(selectedIds.clear())
                                        }
                                    >
                                        <X className={styles.select_button_icon} aria-hidden="true" />
                                        Deselect All
                                    </button>
                                    <button
                                        type="button"
                                        className={styles.select_button}
                                        onClick={async () => {
                                            setSelectingPlugin(true);
                                        }}
                                    >
                                        <Puzzle className={styles.select_button_icon} aria-hidden="true" />
                                        Run Plugins
                                    </button>
                                    <button
                                        type="button"
                                        className={`${styles.select_button} ${styles.select_button_danger}`}
                                        onClick={async (e) => {
                                            e.stopPropagation();

                                            const confirmed = window.confirm(
                                                "Delete these images? This cannot be undone.",
                                            );
                                            if (!confirmed) return;

                                            (
                                                await Promise.all(
                                                    selectedIds
                                                        .toArray()
                                                        .filter(
                                                            (val) => val !== 0,
                                                        )
                                                        .map((val) =>
                                                            FileDeleter(val)(),
                                                        ),
                                                )
                                            ).forEach(async (prom) => {
                                                const result = prom;
                                                if (!result.ok)
                                                    console.error(
                                                        `Could not delete file! ${await result.error.response.text()}`,
                                                    );
                                            });

                                            setImageIds(
                                                imageIds.filter(
                                                    (val) =>
                                                        !selectedIds.has(val),
                                                ),
                                            );
                                            setSelectedIds(new FreezeTagSet());
                                        }}
                                    >
                                        <Trash2 className={styles.select_button_icon} aria-hidden="true" />
                                        Delete Images
                                    </button>
                                </div>
                                <div className={styles.gallery}>
                                    <MassTaggingGallery
                                        image_ids={imageIds}
                                        selectedIds={selectedIds}
                                        onChange={(id) => {
                                            setSelectedIds(
                                                selectedIds.toggle(id),
                                            );
                                        }}
                                    />
                                </div>
                            </div>
                            <div className={styles.tag_change_container}>
                                <TagChangeButton
                                    image_ids={selectedIds}
                                    do_seeding={true}
                                />
                            </div>
                        </div>
                    </div>
                ) : (
                    <div className={styles.main_parent}>
                        <MainGallery
                            image_ids={imageIds}
                            onSearchTag={(tag) =>
                                setSearchTerm((prev) =>
                                    addTagToQuery(prev, tag),
                                )
                            }
                            onDelete={(_deletedId) => setSearchTerm("" + query)} // Forces the query to recompute and fetch new IDs
                        />
                    </div>
                )}

                {selectingPlugin && (
                    <ManualRunMenu
                        onClose={() => setSelectingPlugin(false)}
                        onPluginChosen={(plugin_name, hook_name) => {
                            PluginRunner(
                                plugin_name,
                                hook_name,
                                selectedIds.toArray(),
                            );
                            router.replace("/jobs");
                        }}
                        hookSignature="image_batch"
                    />
                )}
            </main>
        </>
    );
}
