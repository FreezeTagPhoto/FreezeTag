/**
 * @jest-environment node
 */

import "@testing-library/jest-dom";
import {
    computeTagSuggestions,
    normalizeTag,
    rankTag,
} from "@/common/gallery/tags";

describe("common/gallery/tags", () => {
    it("normalizeTag trims and collapses whitespace", () => {
        expect(normalizeTag("  hello   world  ")).toBe("hello world");
        expect(normalizeTag("a\t\tb\nc")).toBe("a b c");
        expect(normalizeTag("x")).toBe("x");
    });

    it("rankTag scoring: exact, prefix, contains, subsequence, none", () => {
        expect(rankTag("Beach", "beach")).toBe(0); // exact (case-insensitive)
        expect(rankTag("Beach Day", "bea")).toBe(1); // prefix
        const contains = rankTag("my beach day", "beach");
        expect(contains).toBeGreaterThanOrEqual(2);
        expect(contains).toBeLessThan(3);

        expect(rankTag("alphabet", "abt")).toBe(3); // subsequence a-b-t
        expect(rankTag("hello", "zzz")).toBe(999);
    });

    it("computeTagSuggestions returns [] when addOpen false or allTags empty", () => {
        expect(
            computeTagSuggestions({
                addOpen: false,
                allTags: ["a"],
                currentTags: [],
                addValue: "",
                tagSuggestPinned: false,
            }),
        ).toStrictEqual([]);

        expect(
            computeTagSuggestions({
                addOpen: true,
                allTags: null,
                currentTags: [],
                addValue: "",
                tagSuggestPinned: true,
            }),
        ).toStrictEqual([]);

        expect(
            computeTagSuggestions({
                addOpen: true,
                allTags: [],
                currentTags: [],
                addValue: "",
                tagSuggestPinned: true,
            }),
        ).toStrictEqual([]);
    });

    it("computeTagSuggestions excludes currentTags", () => {
        const out = computeTagSuggestions({
            addOpen: true,
            allTags: ["beach", "rock", "snow"],
            currentTags: ["rock"],
            addValue: "",
            tagSuggestPinned: true,
        });

        expect(out).toStrictEqual(["beach", "snow"]); // sorted alpha, excludes rock
    });

    it("computeTagSuggestions: empty needle returns [] unless pinned; pinned returns sorted + limited", () => {
        const notPinned = computeTagSuggestions({
            addOpen: true,
            allTags: ["b", "a", "c"],
            currentTags: [],
            addValue: "   ",
            tagSuggestPinned: false,
        });
        expect(notPinned).toStrictEqual([]);

        const pinned = computeTagSuggestions({
            addOpen: true,
            allTags: ["b", "a", "c"],
            currentTags: [],
            addValue: "   ",
            tagSuggestPinned: true,
            limit: 2,
        });
        expect(pinned).toStrictEqual(["a", "b"]);
    });

    it("computeTagSuggestions ranks by rankTag then alpha and applies limit", () => {
        const out = computeTagSuggestions({
            addOpen: true,
            allTags: ["mountain", "beach", "my beach day", "alphabet", "zzz"],
            currentTags: [],
            addValue: "beach",
            tagSuggestPinned: false,
            limit: 10,
        });

        // best should be exact-ish/prefix/contains)
        expect(out.includes("zzz")).toBe(false);
        expect(out[0]).toBe("beach"); // exact
        // contains should still appear
        expect(out).toContain("my beach day");
    });
});
