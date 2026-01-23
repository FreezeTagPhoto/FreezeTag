"use client";
import { useEffect, useMemo, useRef, useState } from "react";
import SearchHandler from "@/api/query/searchhandler";
import styles from "./page.module.css";
import MainGallery from "@/components/Gallery/MainGallery/MainGallery";
import TopBar from "@/components/TopBar/TopBar";
import { addTagToQuery } from "@/common/search/addtagtoquery";

function normalizeSearchBarString(s: string): string {
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
    const parts = [`sortBy=${sortBy}`, `sortOrder=${sortOrder}`];
    const tail = normalizeSearchBarString(searchTerm);
    if (tail.length > 0) parts.push(tail);
    return parts.join(";");
}

export default function Home() {
    const [images_ids, set_image_ids] = useState<number[]>([]);

    const [searchTerm, setSearchTerm] = useState("");
    const [sortBy, setSortBy] = useState<string>("DateAdded");
    const [sortOrder, setSortOrder] = useState<string>("DESC");

    const query = useMemo(
        () => buildQuery(sortBy, sortOrder, searchTerm),
        [sortBy, sortOrder, searchTerm],
    );

    const reqIdRef = useRef(0);

    useEffect(() => {
        const myReqId = ++reqIdRef.current;

        (async () => {
            const result = await SearchHandler(query);
            if (myReqId !== reqIdRef.current) return;

            if (result.ok) {
                set_image_ids(result.value);
            } else {
                console.error(
                    "Got " +
                        result.error.status +
                        " from backend with message " +
                        result.error.message,
                );
                // TODO: Show error to user
            }
        })();

        return () => {
            reqIdRef.current++;
        };
    }, [query]);

    return (
        <>
            <TopBar
                searchTerm={searchTerm}
                onSearchTermChange={setSearchTerm}
                sortBy={sortBy}
                onSortByChange={setSortBy}
                sortOrder={sortOrder}
                onSortOrderChange={setSortOrder}
            />

            <main className={styles.main}>
                <header className={styles.headerRow}>
                    <div>
                        <h1 className={styles.h1}>Gallery</h1>
                        <p className={styles.subtle}>
                            {images_ids.length}{" "}
                            {images_ids.length !== 1 ? "images" : "image"}
                        </p>
                    </div>
                    <div className={styles.pillsRow} />
                </header>

                <MainGallery
                    image_ids={images_ids}
                    onSearchTag={(tag) => {
                        setSearchTerm((prev) => addTagToQuery(prev, tag));
                    }}
                />
            </main>
        </>
    );
}
