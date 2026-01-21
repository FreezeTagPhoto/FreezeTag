/**
 * @jest-environment node
 */

import { parseUserQuery } from "@/common/search/parse";

function kmToAngularDegrees(km: number): number {
    const EARTH_RADIUS_KM = 6371;
    return (km / EARTH_RADIUS_KM) * (180 / Math.PI);
}

function formatDeg(deg: number): string {
    return deg.toFixed(6).replace(/\.?0+$/, "");
}

describe("common/search/parse.parseUserQuery", () => {
    it("ignores empty chunks and whitespace-only chunks", () => {
        const tokens = parseUserQuery(` ;   ;  `);
        expect(tokens).toHaveLength(0);
    });

    it("parses unquoted tag and quoted tag, and handles missing closing quote for tag", () => {
        const tokens = parseUserQuery(`beach; "among us"; "oops`);
        expect(tokens).toHaveLength(3);

        expect(tokens[0].kind).toBe("tag");
        if (tokens[0].kind === "tag") {
            expect(tokens[0].value).toBe("beach");
            expect(tokens[0].exact).toBe(false);
            expect(tokens[0].error).toBeUndefined();
        }

        expect(tokens[1].kind).toBe("tag");
        if (tokens[1].kind === "tag") {
            expect(tokens[1].value).toBe("among us");
            expect(tokens[1].exact).toBe(true);
            expect(tokens[1].error).toBeUndefined();
        }

        expect(tokens[2].kind).toBe("tag");
        if (tokens[2].kind === "tag") {
            expect(tokens[2].value).toBe("oops");
            expect(tokens[2].exact).toBe(true);
            expect(tokens[2].error).toBe("Missing closing quote");
        }
    });

    it("unknown filter keys become tag tokens with an error", () => {
        const tokens = parseUserQuery(`wat=123`);
        expect(tokens).toHaveLength(1);

        const t = tokens[0];
        expect(t.kind).toBe("tag");
        if (t.kind === "tag") {
            expect(t.error).toBe(`Unknown filter "wat"`);
            expect(t.value).toBe(`wat=123`);
        }
    });

    it("parses normal field keys and detects missing closing quote", () => {
        const tokens = parseUserQuery(`make="Apple"; model="Galaxy`);
        expect(tokens).toHaveLength(2);

        const make = tokens[0];
        expect(make.kind).toBe("field");
        if (make.kind === "field") {
            expect(make.key).toBe("make");
            expect(make.value).toBe("Apple");
            expect(make.exact).toBe(true);
            expect(make.error).toBeUndefined();
        }

        const model = tokens[1];
        expect(model.kind).toBe("field");
        if (model.kind === "field") {
            expect(model.key).toBe("model");
            expect(model.value).toBe("Galaxy");
            expect(model.exact).toBe(true);
            expect(model.error).toBe("Missing closing quote");
        }
    });

    it("date keys: unix passthrough, ISO date conversion, invalid date produces error", () => {
        const iso = "1995-12-17T03:24:00Z";
        const expectedUnix = Math.floor(
            new Date(iso).getTime() / 1000,
        ).toString();

        const tokens = parseUserQuery(
            `takenBefore=123456789; takenAfter="${iso}"; uploadedAfter=FakeDate`,
        );
        expect(tokens).toHaveLength(3);

        const a = tokens[0];
        expect(a.kind).toBe("field");
        if (a.kind === "field") {
            expect(a.key).toBe("takenBefore");
            expect(a.value).toBe("123456789");
            expect(a.error).toBeUndefined();
        }

        const b = tokens[1];
        expect(b.kind).toBe("field");
        if (b.kind === "field") {
            expect(b.key).toBe("takenAfter");
            expect(b.value).toBe(expectedUnix);
            expect(b.error).toBeUndefined();
        }

        const c = tokens[2];
        expect(c.kind).toBe("field");
        if (c.kind === "field") {
            expect(c.key).toBe("uploadedAfter");
            expect(c.value).toBe("FakeDate");
            expect(c.error).toBe(
                "Invalid date (use YYYY-MM-DD or unix seconds)",
            );
        }
    });

    it("near key: wrong arity -> error", () => {
        const tokens = parseUserQuery(`near=1,2`);
        expect(tokens).toHaveLength(1);

        const t = tokens[0];
        expect(t.kind).toBe("field");
        if (t.kind === "field") {
            expect(t.key).toBe("near");
            expect(t.error).toMatch(/near expects 3 comma-separated values/i);
        }
    });

    it("near key: non-numeric lat/lon -> error", () => {
        const tokens = parseUserQuery(`near=a,b,5km`);
        const t = tokens[0];
        expect(t.kind).toBe("field");
        if (t.kind === "field") {
            expect(t.key).toBe("near");
            expect(t.error).toMatch(/lat\/lon must be numbers/i);
        }
    });

    it("near key: out of bounds lat/lon -> error", () => {
        const tokens1 = parseUserQuery(`near=91,0,5km`);
        expect(tokens1[0].kind).toBe("field");
        if (tokens1[0].kind === "field") {
            expect(tokens1[0].error).toMatch(/latitude must be between/i);
        }

        const tokens2 = parseUserQuery(`near=0,181,5km`);
        expect(tokens2[0].kind).toBe("field");
        if (tokens2[0].kind === "field") {
            expect(tokens2[0].error).toMatch(/longitude must be between/i);
        }
    });

    it("near key: distance regex mismatch -> error", () => {
        const tokens = parseUserQuery(`near=40,-110,huh`);
        const t = tokens[0];
        expect(t.kind).toBe("field");
        if (t.kind === "field") {
            expect(t.key).toBe("near");
            expect(t.error).toMatch(/near distance must look like/i);
        }
    });

    it("near key: distance <= 0 -> error", () => {
        const tokens = parseUserQuery(`near=40,-110,0km`);
        const t = tokens[0];
        expect(t.kind).toBe("field");
        if (t.kind === "field") {
            expect(t.error).toMatch(/positive number/i);
        }
    });

    it("near key: supports km default (unitless), m, mi, ft, yd, deg, and degree symbol + cleaning", () => {
        // unitless defaults to KM
        const degUnitless = formatDeg(kmToAngularDegrees(3));
        const t1 = parseUserQuery(`near=1,2,3`)[0];
        expect(t1.kind).toBe("field");
        if (t1.kind === "field") {
            expect(t1.value).toBe(`1,2,${degUnitless}`);
            expect(t1.error).toBeUndefined();
        }

        // meters
        const degM = formatDeg(kmToAngularDegrees(0.5)); // 500m = 0.5km
        const t2 = parseUserQuery(`near=1,2,500m`)[0];
        if (t2.kind === "field") {
            expect(t2.value).toBe(`1,2,${degM}`);
        }

        // miles
        const degMi = formatDeg(kmToAngularDegrees(3 * 1.609344));
        const t3 = parseUserQuery(`near=1,2,3mi`)[0];
        if (t3.kind === "field") {
            expect(t3.value).toBe(`1,2,${degMi}`);
        }

        // feet (case + whitespace + trailing punctuation clean)
        const degFt = formatDeg(kmToAngularDegrees(20 * 0.0003048));
        const t4 = parseUserQuery(`near=1,2, 20FEET.)`)[0];
        if (t4.kind === "field") {
            expect(t4.value).toBe(`1,2,${degFt}`);
        }

        // yards
        const degYd = formatDeg(kmToAngularDegrees(10 * 0.0009144));
        const t5 = parseUserQuery(`near=1,2,10yards`)[0];
        if (t5.kind === "field") {
            expect(t5.value).toBe(`1,2,${degYd}`);
        }

        // explicit degrees
        const t6 = parseUserQuery(`near=1,2,0.1deg`)[0];
        if (t6.kind === "field") {
            expect(t6.value).toBe(`1,2,0.1`);
        }

        // degree symbol: "0.1°" -> "0.1deg"
        const t7 = parseUserQuery(`near=1,2,0.1°`)[0];
        if (t7.kind === "field") {
            expect(t7.value).toBe(`1,2,0.1`);
        }
    });

    it("range computation: selects the trimmed chunk region (sanity check)", () => {
        const input = `  make="Apple"  ;  beach `;
        const tokens = parseUserQuery(input);
        expect(tokens).toHaveLength(2);

        const make = tokens[0];
        expect(make.range.start).toBeGreaterThanOrEqual(0);
        expect(make.range.end).toBeGreaterThan(make.range.start);
        expect(input.slice(make.range.start, make.range.end)).toContain(
            `make="Apple"`,
        );

        const tag = tokens[1];
        expect(tag.range.start).toBeGreaterThanOrEqual(0);
        expect(tag.range.end).toBeGreaterThan(tag.range.start);
        expect(input.slice(tag.range.start, tag.range.end)).toBe(`beach`);
    });

    it("handles missing closing quote in date field too", () => {
        const tokens = parseUserQuery(`takenAfter="1995-12-17T03:24:00Z`);
        expect(tokens).toHaveLength(1);

        const t = tokens[0];
        expect(t.kind).toBe("field");
        if (t.kind === "field") {
            expect(t.key).toBe("takenAfter");
            expect(t.exact).toBe(true);
            expect(t.error).toBe("Missing closing quote");
        }
    });
});
