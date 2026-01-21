"use client";
import { useState } from "react";
import styles from "./TopBar.module.css";
import SearchBar from "@/components/SearchBar/SearchBar";
import Pill from "@/components/UI/Pill/Pill";

type TopBarProps = { onChangeHandler: (value: string) => void };

export default function TopBar({ onChangeHandler }: TopBarProps) {
    const [searchTerm, setSearchTerm] = useState("");

    // Sort Dropdown
    const [sortBy, setSortBy] = useState<string>("DateAdded");
    const [sortOrder, setSortOrder] = useState<string>("DESC");
    const [visibleSortMenu, setVisibleSortMenu] = useState<boolean>(false);

    const formQueryAndChangeHandler = (
        sortBy: string,
        sortOrder: string,
        searchBarString: string,
    ) => {
        const query = `sortBy=${sortBy};sortOrder=${sortOrder};${searchBarString}`;
        onChangeHandler(query);
    };

    return (
        <div className={styles.bar}>
            <SearchBar
                value={searchTerm}
                onChange={(v) => {
                    setSearchTerm(v);
                    formQueryAndChangeHandler(sortBy, sortOrder, v);
                }}
            />

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
                                defaultValue={sortBy}
                                onChange={(event) => {
                                    setSortBy(event.target.value);
                                    formQueryAndChangeHandler(
                                        event.target.value,
                                        sortOrder,
                                        searchTerm,
                                    );
                                }}
                                size={2}
                            >
                                <option value="DateCreated">
                                    Date Created
                                </option>
                                <option value="DateAdded">Date Added</option>
                            </select>
                            <select
                                defaultValue={sortOrder}
                                onChange={(event) => {
                                    setSortOrder(event.target.value);
                                    formQueryAndChangeHandler(
                                        sortBy,
                                        event.target.value,
                                        searchTerm,
                                    );
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
