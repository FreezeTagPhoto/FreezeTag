"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import SearchHandler from "@/api/query/searchhandler";
import TagGetter from "@/api/tags/taggetter";
import styles from "../page.module.css";
import MainGallery from "@/components/Gallery/MainGallery/MainGallery";
import TopBar from "@/components/TopBar/TopBar";
import { addTagToQuery } from "@/common/search/addtagtoquery";
import TagCounter from "@/api/tags/tagcounter";

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
                tags={tagsForTopBar}
            />

            <main className={styles.main}>
                <header className={styles.headerRow}>
                    <div>
                        <h1 className={styles.h1}>Gallery</h1>
                        <p className={styles.subtle}>
                            {imageIds.length}{" "}
                            {imageIds.length !== 1 ? "images" : "image"}
                        </p>
                    </div>
                    <div className={styles.pillsRow} />
                </header>

                <MainGallery
                    image_ids={imageIds}
                    onSearchTag={(tag) =>
                        setSearchTerm((prev) => addTagToQuery(prev, tag))
                    }
                />
            </main>
        </>
    );
}
