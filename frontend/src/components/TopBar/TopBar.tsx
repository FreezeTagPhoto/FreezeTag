"use client";

import { useState } from "react";
import styles from "./TopBar.module.css";
import SearchBar from "@/components/SearchBar/SearchBar";
import Pill from "@/components/UI/Pill/Pill";

type TopBarProps = {
    searchTerm: string;
    onSearchTermChange: (value: string) => void;

    sortBy: string;
    onSortByChange: (value: string) => void;

    sortOrder: string;
    onSortOrderChange: (value: string) => void;
};

export default function TopBar({
    searchTerm,
    onSearchTermChange,
    sortBy,
    onSortByChange,
    sortOrder,
    onSortOrderChange,
}: TopBarProps) {
    const [visibleSortMenu, setVisibleSortMenu] = useState(false);

    return (
        <div className={styles.bar}>
            <SearchBar value={searchTerm} onChange={onSearchTermChange} />

            <div className={styles.pills}>
                <Pill label="Tags" caret variant="menu" />

                <div className={styles.search_container}>
                    <Pill
                        label="Sort"
                        caret
                        invertCaret={visibleSortMenu}
                        variant="menu"
                        onClick={() => setVisibleSortMenu(!visibleSortMenu)}
                    />

                    {visibleSortMenu && (
                        <div className={styles.search_dropdown}>
                            <select
                                value={sortBy}
                                onChange={(event) => {
                                    onSortByChange(event.target.value);
                                }}
                                size={2}
                            >
                                <option value="DateCreated">
                                    Date Created
                                </option>
                                <option value="DateAdded">Date Added</option>
                            </select>

                            <select
                                value={sortOrder}
                                onChange={(event) => {
                                    onSortOrderChange(event.target.value);
                                }}
                                size={2}
                            >
                                <option value="ASC">Ascending</option>
                                <option value="DESC">Descending</option>
                            </select>
                        </div>
                    )}
                </div>

                <Pill label="Export" caret variant="menu" />
            </div>
        </div>
    );
}
