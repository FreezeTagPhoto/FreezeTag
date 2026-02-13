"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import TagGetter from "@/api/tags/taggetter";
import TagAdder from "@/api/tags/tagadder";
import TagRemover from "@/api/tags/tagremover";
import { normalizeTag, rankTag } from "@/common/gallery/tags";

export type UseTagEditorArgs = {
    selectedId: number | null;
    tagsById: Record<number, string[]>;
    setTagsById: React.Dispatch<React.SetStateAction<Record<number, string[]>>>;
    currentTags: string[] | null;
};

export type UseTagEditorReturn = {
    // mutation status
    tagMutating: boolean;
    tagMutateError: string | null;
    tagMutateInfo: string | null;

    // editor state
    addOpen: boolean;
    addValue: string;
    addInputRef: React.RefObject<HTMLInputElement | null>;
    addEditorRef: React.RefObject<HTMLDivElement | null>;

    // suggestions / all-tags
    allTagsLoading: boolean;
    tagSuggestOpen: boolean;
    tagSuggestPinned: boolean;
    tagSuggestIndex: number;
    tagSuggestDisabled: boolean;

    tagSuggestions: string[];
    showTagDropdown: boolean;

    // actions
    openAddEditor: () => Promise<void>;
    closeAddEditor: () => void;

    removeTagFromSelected: (tag: string) => Promise<void>;
    addTagToSelected: (tagOverride?: string) => Promise<void>;

    toggleSuggestions: () => Promise<void>;

    // input event handlers
    onAddValueChange: (value: string) => void;
    onAddInputFocusOrClick: () => void;
    onAddInputKeyDown: (key: string) => Promise<void>;

    // for suggestion hover
    setTagSuggestIndex: React.Dispatch<React.SetStateAction<number>>;
};

export function useTagEditor({
    selectedId,
    tagsById,
    setTagsById,
    currentTags,
}: UseTagEditorArgs): UseTagEditorReturn {
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
    const [tagSuggestDisabled, setTagSuggestDisabled] = useState(false);

    const ensureAllTagsLoaded = useCallback(async () => {
        if (allTags !== null || allTagsLoading) return;
        setAllTagsLoading(true);
        const res = await TagGetter();
        if (res.ok) setAllTags(res.value);
        else setAllTags([]);
        setAllTagsLoading(false);
    }, [allTags, allTagsLoading]);

    const setCurrentTagsForSelected = useCallback(
        (next: string[]) => {
            if (selectedId === null) return;
            setTagsById((prev) => ({ ...prev, [selectedId]: next }));
        },
        [selectedId, setTagsById],
    );

    const removeTagFromSelected = useCallback(
        async (tag: string) => {
            if (selectedId === null) return;

            const current = tagsById[selectedId] ?? [];
            const next = current.filter((t) => t !== tag);

            setTagMutateError(null);
            setTagMutateInfo(null);
            setTagMutating(true);

            setCurrentTagsForSelected(next);

            const res = await TagRemover([selectedId], [tag]);
            setTagMutating(false);

            if (!res.ok) {
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

    const closeAddEditor = useCallback(() => {
        setAddOpen(false);
        setAddValue("");
        setTagSuggestOpen(false);
        setTagSuggestPinned(false);
        setTagSuggestIndex(0);
        setTagSuggestDisabled(false);
    }, []);

    const openAddEditor = useCallback(async () => {
        setAddOpen(true);
        setAddValue("");
        setTagSuggestIndex(0);
        setTagSuggestOpen(false);
        setTagSuggestPinned(false);
        setTagSuggestDisabled(false);
        await ensureAllTagsLoaded();
    }, [ensureAllTagsLoaded]);

    const addTagToSelected = useCallback(
        async (tagOverride?: string) => {
            if (selectedId === null) return;

            const tag = normalizeTag(tagOverride ?? addValue);
            if (!tag) return;

            const current = tagsById[selectedId] ?? [];

            if (current.includes(tag)) {
                setTagMutateInfo(`Tag "${tag}" already exists.`);
                closeAddEditor();
                return;
            }

            const next = [...current, tag];

            setTagMutateError(null);
            setTagMutateInfo(null);
            setTagMutating(true);

            setCurrentTagsForSelected(next);

            const res = await TagAdder([selectedId], [tag]);
            setTagMutating(false);

            if (!res.ok) {
                setCurrentTagsForSelected(current);
                setTagMutateError(res.error.message);
                return;
            }

            // keep suggestion list fresh
            setAllTags((prev) => {
                if (!prev) return prev;
                return prev.includes(tag) ? prev : [...prev, tag];
            });

            if (res.value.length > 0) setTagMutateInfo(res.value.join(" • "));

            closeAddEditor();
        },
        [
            selectedId,
            addValue,
            tagsById,
            setCurrentTagsForSelected,
            closeAddEditor,
        ],
    );

    // reset tag editor UI when changing images
    useEffect(() => {
        setTagMutateError(null);
        setTagMutateInfo(null);
        closeAddEditor();
    }, [selectedId, closeAddEditor]);

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
        const allowEmpty = tagSuggestPinned;

        if (!needle) {
            if (!allowEmpty) return [];
            return [...candidates].sort((a, b) => a.localeCompare(b));
            // .slice(0, 10);
        }

        return (
            candidates
                .map((t) => ({ tag: t, score: rankTag(t, needle) }))
                .filter((x) => x.score < 999)
                .sort((a, b) => a.score - b.score || a.tag.localeCompare(b.tag))
                // .slice(0, 10)
                .map((x) => x.tag)
        );
    }, [addOpen, allTags, currentTags, addValue, tagSuggestPinned]);

    const showTagDropdown =
        tagSuggestOpen && (allTagsLoading || tagSuggestions.length > 0);

    useEffect(() => {
        if (!tagSuggestOpen) return;
        setTagSuggestIndex((i) =>
            Math.max(0, Math.min(i, Math.max(0, tagSuggestions.length - 1))),
        );
    }, [tagSuggestOpen, tagSuggestions.length]);

    // click outside closes suggestions
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

    const toggleSuggestions = useCallback(async () => {
        const hasNeedle = normalizeTag(addValue).length > 0;

        if (tagSuggestDisabled) {
            setTagSuggestDisabled(false);

            if (tagSuggestPinned || hasNeedle) {
                await ensureAllTagsLoaded();
                setTagSuggestIndex(0);
                setTagSuggestOpen(true);
            } else {
                setTagSuggestOpen(false);
            }
            return;
        }

        if (hasNeedle && !tagSuggestPinned) {
            setTagSuggestDisabled(true);
            setTagSuggestOpen(false);
            return;
        }

        const nextPinned = !tagSuggestPinned;
        setTagSuggestPinned(nextPinned);

        if (nextPinned) {
            await ensureAllTagsLoaded();
            setTagSuggestIndex(0);
            setTagSuggestOpen(true);
        } else {
            setTagSuggestOpen(hasNeedle && !tagSuggestDisabled);
        }
    }, [addValue, tagSuggestDisabled, tagSuggestPinned, ensureAllTagsLoaded]);

    const onAddValueChange = useCallback(
        (value: string) => {
            setAddValue(value);
            setTagSuggestIndex(0);

            const hasNeedle = normalizeTag(value).length > 0;

            if (tagSuggestDisabled) {
                setTagSuggestOpen(false);
                return;
            }

            if (hasNeedle) {
                setTagSuggestOpen(true);
            } else {
                setTagSuggestOpen(tagSuggestPinned);
            }
        },
        [tagSuggestDisabled, tagSuggestPinned],
    );

    const onAddInputFocusOrClick = useCallback(() => {
        if (tagSuggestDisabled) return;

        const hasNeedle = normalizeTag(addValue).length > 0;
        if (hasNeedle || tagSuggestPinned) setTagSuggestOpen(true);
    }, [tagSuggestDisabled, addValue, tagSuggestPinned]);

    const onAddInputKeyDown = useCallback(
        async (key: string) => {
            if (key === "ArrowDown") {
                if (tagSuggestions.length === 0) return;
                setTagSuggestOpen(true);
                setTagSuggestIndex((i) =>
                    Math.min(i + 1, tagSuggestions.length - 1),
                );
                return;
            }

            if (key === "ArrowUp") {
                if (tagSuggestions.length === 0) return;
                setTagSuggestOpen(true);
                setTagSuggestIndex((i) => Math.max(i - 1, 0));
                return;
            }

            if (key === "Enter") {
                const chosen =
                    tagSuggestOpen && tagSuggestions[tagSuggestIndex]
                        ? tagSuggestions[tagSuggestIndex]
                        : undefined;

                await addTagToSelected(chosen);

                setTagSuggestOpen(false);
                setTagSuggestPinned(false);
            }
        },
        [addTagToSelected, tagSuggestions, tagSuggestIndex, tagSuggestOpen],
    );

    return {
        tagMutating,
        tagMutateError,
        tagMutateInfo,

        addOpen,
        addValue,
        addInputRef,
        addEditorRef,

        allTagsLoading,
        tagSuggestOpen,
        tagSuggestPinned,
        tagSuggestIndex,
        tagSuggestDisabled,

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
    };
}
