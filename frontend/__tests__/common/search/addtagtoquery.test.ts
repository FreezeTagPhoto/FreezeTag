/**
 * @jest-environment node
 */

import { addTagToQuery } from "@/common/search/addtagtoquery";
import { parseUserQuery } from "@/common/search/parse";

describe("common/search/addtagtoquery.addTagToQuery", () => {
    it("returns input unchanged when tag is empty/whitespace", () => {
        expect(addTagToQuery("make=Toyota;", "")).toBe("make=Toyota;");
        expect(addTagToQuery("make=Toyota;", "   ")).toBe("make=Toyota;");
    });

    it("adds quoted tag to empty input", () => {
        expect(addTagToQuery("", "beach")).toBe(`"beach"; `);
        expect(addTagToQuery("   ", "beach")).toBe(`"beach"; `);
    });

    it("strips quotes inside tag before quoting (formatTagToken coverage)", () => {
        expect(addTagToQuery("", `a"b"c`)).toBe(`"abc"; `);
        expect(addTagToQuery("", `  a"b"c  `)).toBe(`"abc"; `);
    });

    it("appends tag to a non-empty query (removes trailing semicolon/whitespace first)", () => {
        expect(addTagToQuery(`make=Toyota`, "beach")).toBe(
            `make=Toyota; "beach"; `,
        );

        expect(addTagToQuery(`make=Toyota;`, "beach")).toBe(
            `make=Toyota; "beach"; `,
        );

        expect(addTagToQuery(`make=Toyota ;   `, "beach")).toBe(
            `make=Toyota; "beach"; `,
        );

        expect(addTagToQuery(`make=Toyota   ;   `, "beach")).toBe(
            `make=Toyota; "beach"; `,
        );
    });

    it("detects already-present tag via parseUserQuery and normalizes trailing separator", () => {
        const input = `beach`;
        const tokens = parseUserQuery(input);
        expect(
            tokens.some((t) => t.kind === "tag" && t.value === "beach"),
        ).toBe(true);

        expect(addTagToQuery(`beach`, "beach")).toBe(`beach; `);
        expect(addTagToQuery(`beach;`, "beach")).toBe(`beach; `);
        expect(addTagToQuery(`beach; `, "beach")).toBe(`beach; `);
        expect(addTagToQuery(`beach   `, "beach")).toBe(`beach; `);
    });

    it("treats quoted tags as already present when parseUserQuery yields same value", () => {
        expect(addTagToQuery(`"beach"`, "beach")).toBe(`"beach"; `);
        expect(addTagToQuery(`  "beach"  ;  `, "beach")).toBe(`  "beach"  ; `);
    });

    it("preserves input exactly when alreadyHas is true and it already ends with '; '", () => {
        expect(addTagToQuery(`beach; `, "beach")).toBe(`beach; `);
    });

    it("does not consider different casing as already present (exact match)", () => {
        expect(addTagToQuery(`Beach`, "beach")).toBe(`Beach; "beach"; `);
        expect(addTagToQuery(`"Beach"`, "beach")).toBe(`"Beach"; "beach"; `);
    });

    it("does not consider tag with extra spaces as already present (exact match on trimmed tag)", () => {
        expect(addTagToQuery(`"beach"`, "  beach  ")).toBe(`"beach"; `);
    });

    it("handles input that ends with a semicolon + spaces when alreadyHas is true", () => {
        expect(addTagToQuery(`beach;     `, "beach")).toBe(`beach; `);
    });

    it("handles blank tag but non-blank input", () => {
        expect(addTagToQuery("tag; ", "")).toBe("tag; ");
    });
});
