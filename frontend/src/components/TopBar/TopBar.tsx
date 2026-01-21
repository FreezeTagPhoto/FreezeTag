"use client";
import { useState } from "react";
import styles from "./TopBar.module.css";
import SearchBar from "@/components/SearchBar/SearchBar";
import Pill from "@/components/UI/Pill/Pill";

type TopBarProps = { onChangeHandler: (value: string) => void };

export default function TopBar({ onChangeHandler }: TopBarProps) {
    const [searchTerm, setSearchTerm] = useState("");

    return (
        <div className={styles.bar}>
            <SearchBar
                value={searchTerm}
                onChange={(v) => {
                    setSearchTerm(v);
                    onChangeHandler(v);
                }}
            />

            <div className={styles.pills}>
                <Pill label="Tags" caret variant="menu" />
                <Pill label="People" caret variant="menu" />
                <Pill label="Export" caret variant="menu" />
            </div>
        </div>
    );
}
