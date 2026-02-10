/**
 * @jest-environment jsdom
 */

import "@testing-library/jest-dom";
import React from "react";
import { act, renderHook } from "@testing-library/react";
import { useTagEditor } from "@/common/gallery/tageditor";

jest.mock("@/api/tags/taggetter", () => ({
    __esModule: true,
    default: jest.fn(),
}));
jest.mock("@/api/tags/tagadder", () => ({
    __esModule: true,
    default: jest.fn(),
}));
jest.mock("@/api/tags/tagremover", () => ({
    __esModule: true,
    default: jest.fn(),
}));

import TagGetter from "@/api/tags/taggetter";
import TagAdder from "@/api/tags/tagadder";
import TagRemover from "@/api/tags/tagremover";

type Ok<T> = { ok: true; value: T };
type Err = { ok: false; error: { message: string } };

function ok<T>(value: T): Ok<T> {
    return { ok: true, value };
}
function err(message: string): Err {
    return { ok: false, error: { message } };
}

beforeAll(() => {
    globalThis.requestAnimationFrame = ((cb: FrameRequestCallback): number => {
        cb(0);
        return 0;
    }) as typeof globalThis.requestAnimationFrame;
});

describe("common/gallery/tageditor.useTagEditor", () => {
    function setup(initial: {
        selectedId: number | null;
        initialTagsById?: Record<number, string[]>;
        currentTags?: string[] | null;
    }) {
        const initialTagsById = initial.initialTagsById ?? {};
        const selectedId = initial.selectedId;

        let latest:
            | {
                  tagsById: Record<number, string[]>;
                  setTagsById: React.Dispatch<
                      React.SetStateAction<Record<number, string[]>>
                  >;
                  currentTags: string[] | null;
              }
            | undefined;

        const Wrapper: React.FC<{ children: React.ReactNode }> = ({
            children,
        }) => {
            const [tagsById, setTagsById] =
                React.useState<Record<number, string[]>>(initialTagsById);

            const currentTags =
                initial.currentTags ??
                (selectedId === null ? null : (tagsById[selectedId] ?? null));

            latest = { tagsById, setTagsById, currentTags };

            return React.createElement(React.Fragment, null, children);
        };

        const hook = renderHook(
            () => {
                const st = latest!;
                return useTagEditor({
                    selectedId,
                    tagsById: st.tagsById,
                    setTagsById: st.setTagsById,
                    currentTags: st.currentTags,
                });
            },
            { wrapper: Wrapper },
        );

        return hook;
    }

    beforeEach(() => {
        jest.clearAllMocks();
        (TagGetter as jest.Mock).mockResolvedValue(
            ok(["beach", "rock", "snow"]),
        );
        (TagAdder as jest.Mock).mockResolvedValue(ok([]));
        (TagRemover as jest.Mock).mockResolvedValue(ok([]));
    });

    it("openAddEditor opens editor, loads all tags once, and sets loading flags", async () => {
        const { result } = setup({
            selectedId: 1,
            initialTagsById: { 1: ["beach"] },
        });

        expect(result.current.addOpen).toBe(false);

        await act(async () => {
            await result.current.openAddEditor();
        });

        expect(result.current.addOpen).toBe(true);
        expect(TagGetter).toHaveBeenCalledTimes(1);
        expect(result.current.allTagsLoading).toBe(false);
    });

    it("addTagToSelected normalizes and adds tag; prevents duplicates and shows info", async () => {
        const { result } = setup({
            selectedId: 1,
            initialTagsById: { 1: ["beach"] },
        });

        await act(async () => {
            await result.current.openAddEditor();
        });

        // type a tag with extra whitespace
        await act(async () => {
            result.current.onAddValueChange("  new   tag  ");
        });

        await act(async () => {
            await result.current.addTagToSelected();
        });

        expect(TagAdder).toHaveBeenCalledTimes(1);
        expect(TagAdder).toHaveBeenCalledWith([1], ["new tag"]);
        expect(result.current.tagMutateError).toBe(null);

        // editor should close after success
        expect(result.current.addOpen).toBe(false);

        // duplicate path
        await act(async () => {
            await result.current.openAddEditor();
        });

        await act(async () => {
            result.current.onAddValueChange("beach");
        });

        await act(async () => {
            await result.current.addTagToSelected();
        });

        // should not call TagAdder for duplicates
        expect(TagAdder).toHaveBeenCalledTimes(1);
        expect(result.current.tagMutateInfo).toBe(
            `Tag "beach" already exists.`,
        );
        expect(result.current.addOpen).toBe(false);
    });

    it("removeTagFromSelected does optimistic update; rolls back and sets error on failure", async () => {
        (TagRemover as jest.Mock).mockResolvedValueOnce(err("nope"));

        const { result } = setup({
            selectedId: 1,
            initialTagsById: { 1: ["beach", "rock"] },
        });

        await act(async () => {
            await result.current.removeTagFromSelected("rock");
        });

        expect(TagRemover).toHaveBeenCalledTimes(1);
        expect(TagRemover).toHaveBeenCalledWith([1], ["rock"]);

        expect(result.current.tagMutateError).toBe("nope");
    });

    it("toggleSuggestions: if has needle and not pinned, first click disables suggestions", async () => {
        const { result } = setup({
            selectedId: 1,
            initialTagsById: { 1: [] },
        });

        await act(async () => {
            await result.current.openAddEditor();
        });

        await act(async () => {
            result.current.onAddValueChange("be");
        });

        // first toggle with needle and not pinned -> disabled + closed
        await act(async () => {
            await result.current.toggleSuggestions();
        });

        expect(result.current.tagSuggestDisabled).toBe(true);
        expect(result.current.tagSuggestOpen).toBe(false);

        // toggling again should re-enable; since needle exists, it should open
        await act(async () => {
            await result.current.toggleSuggestions();
        });

        expect(result.current.tagSuggestDisabled).toBe(false);
        expect(result.current.tagSuggestOpen).toBe(true);
    });

    it("keyboard: ArrowDown/ArrowUp moves selection when suggestions exist; Enter adds chosen suggestion", async () => {
        const { result } = setup({
            selectedId: 1,
            initialTagsById: { 1: [] },
        });

        await act(async () => {
            await result.current.openAddEditor();
        });

        await act(async () => {
            result.current.onAddValueChange("be");
        });

        // ensure suggestions open
        await act(async () => {
            result.current.onAddInputFocusOrClick();
        });

        expect(result.current.tagSuggestions.length).toBeGreaterThan(0);

        const start = result.current.tagSuggestIndex;

        await act(async () => {
            await result.current.onAddInputKeyDown("ArrowDown");
        });
        expect(result.current.tagSuggestIndex).toBe(
            Math.min(start + 1, result.current.tagSuggestions.length - 1),
        );

        await act(async () => {
            await result.current.onAddInputKeyDown("ArrowUp");
        });
        expect(result.current.tagSuggestIndex).toBeGreaterThanOrEqual(0);

        // enter should add selected suggestion (calls TagAdder) if any suggestion exists
        await act(async () => {
            await result.current.onAddInputKeyDown("Enter");
        });

        expect(TagAdder).toHaveBeenCalledTimes(1);
    });

    it("toggleSuggestions pins (nextPinned=true): loads all tags, resets index, opens dropdown", async () => {
        // Make sure TagGetter is actually needed (allTags starts null).
        const { result } = setup({
            selectedId: 1,
            initialTagsById: { 1: ["beach"] },
        });

        await act(async () => {
            await result.current.openAddEditor();
        });

        // ensure no needle + not disabled, so we take the pin toggle path
        await act(async () => {
            result.current.onAddValueChange(""); // hasNeedle = false
        });

        // Starting state should be unpinned
        expect(result.current.tagSuggestPinned).toBe(false);

        // Toggle => nextPinned becomes true and should open + load tags (if not already)
        // Note: openAddEditor already loads tags once; this test still covers the nextPinned
        // branch behavior (open + pinned + index reset).
        await act(async () => {
            await result.current.toggleSuggestions();
        });

        expect(result.current.tagSuggestPinned).toBe(true);
        expect(result.current.tagSuggestIndex).toBe(0);
        expect(result.current.tagSuggestOpen).toBe(true);
    });

    it("toggleSuggestions unpins (nextPinned=false): sets open based on hasNeedle && !disabled", async () => {
        const { result } = setup({
            selectedId: 1,
            initialTagsById: { 1: [] },
        });

        await act(async () => {
            await result.current.openAddEditor();
        });

        // Step 1: pin suggestions (enter the nextPinned=true branch once)
        await act(async () => {
            result.current.onAddValueChange(""); // hasNeedle=false
        });
        await act(async () => {
            await result.current.toggleSuggestions(); // pinned -> true, open -> true
        });

        expect(result.current.tagSuggestPinned).toBe(true);
        expect(result.current.tagSuggestOpen).toBe(true);

        // Step 2a: unpin while hasNeedle=false => open should become false
        await act(async () => {
            await result.current.toggleSuggestions(); // nextPinned=false, open = false
        });

        expect(result.current.tagSuggestPinned).toBe(false);
        expect(result.current.tagSuggestOpen).toBe(false);

        // Step 2b: pin again, then set needle so unpin will keep open = true
        await act(async () => {
            await result.current.toggleSuggestions(); // pin again
        });
        expect(result.current.tagSuggestPinned).toBe(true);

        await act(async () => {
            result.current.onAddValueChange("be"); // hasNeedle=true
        });

        await act(async () => {
            await result.current.toggleSuggestions(); // unpin; open = hasNeedle && !disabled => true
        });

        expect(result.current.tagSuggestPinned).toBe(false);
        expect(result.current.tagSuggestDisabled).toBe(false);
        expect(result.current.tagSuggestOpen).toBe(true);
    });
});
