/**
 * @jest-environment node
 */

import {
    testing_SearchHandler,
    testing_SearchResponse,
} from "@/api/query/searchhandler";

import SearchHandler from "@/api/query/searchhandler";

import { RequestError } from "@/api/common/apihandler";

import { Result, Ok, Err } from "@/common/result";

type HandlerReturnType = Promise<Result<testing_SearchResponse, RequestError>>;

function kmToAngularDegrees(km: number): number {
    const EARTH_RADIUS_KM = 6371;
    return (km / EARTH_RADIUS_KM) * (180 / Math.PI);
}

function formatDeg(deg: number): string {
    return deg.toFixed(6).replace(/\.?0+$/, "");
}

describe("Search Handler", () => {
    it("Should percolate images in order", async () => {
        const handler = async (_: BodyInit): HandlerReturnType => {
            return Ok([6, 7, 1, 2, 3]);
        };

        const result = await testing_SearchHandler(handler, "");
        expect(result.ok).toBeTruthy();

        if (result.ok) {
            expect(result.value).toStrictEqual([6, 7, 1, 2, 3]);
        }
    });

    it("Should compile tag search properly", async () => {
        const handler = async (query: BodyInit): HandlerReturnType => {
            expect(typeof query === "string").toBeTruthy();
            if (typeof query === "string") {
                const components = query.split("&");
                expect(components.includes("tagLike=among%20us")).toBeTruthy();
                expect(
                    components.includes("takenBefore=1234567890"),
                ).toBeTruthy();
                expect(components.includes("make=Apple")).toBeTruthy();
            }
            return Ok([]);
        };

        const user_query = `among us; takenBefore=1234567890; make="Apple"`;
        await testing_SearchHandler(handler, user_query);
    });

    it("Should compile tag search properly 2", async () => {
        const handler = async (query: BodyInit): HandlerReturnType => {
            expect(typeof query === "string").toBeTruthy();
            if (typeof query === "string") {
                const components = query.split("&");
                expect(components.includes("tag=67")).toBeTruthy();
                expect(components.includes("makeLike=Samsung")).toBeTruthy();
                expect(components.includes("modelLike=Galaxy")).toBeTruthy();
                expect(components.includes("model=Note%207")).toBeTruthy();
            }
            return Ok([]);
        };

        const user_query = `"67"; make=Samsung; model=Galaxy; model="Note 7"`;
        await testing_SearchHandler(handler, user_query);
    });

    it("Should percolate status code on failure", async () => {
        const handler = async (_: BodyInit): HandlerReturnType => {
            return Err({ status_code: 404, response: new Response() });
        };

        const result = await testing_SearchHandler(handler, "");
        expect(result).toStrictEqual(
            Err({ status: 404, message: await new Response().text() }),
        );
    });

    it("Should get message on 400", async () => {
        const handler = async (_: BodyInit): HandlerReturnType => {
            const response = new Response('{"error": "true"}');
            return Err({
                status_code: 400,
                response,
            });
        };

        const result = await testing_SearchHandler(handler, "");
        expect(result).toStrictEqual(Err({ status: 400, message: "true" }));
    });

    it("Should have working near queries (unitless defaults to km)", async () => {
        const handler = async (query: BodyInit): HandlerReturnType => {
            // near=1,2,3  -> distance=3km -> converted to degrees and formatted
            const deg = formatDeg(kmToAngularDegrees(3));
            expect(query).toBe(`near=1%2C2%2C${encodeURIComponent(deg)}`);
            return Ok([]);
        };

        const user_query = `near=1,2,3`;
        await testing_SearchHandler(handler, user_query);
    });

    it("Shouldn't parse unix time", async () => {
        const handler = async (query: BodyInit): HandlerReturnType => {
            expect(query).toBe("takenBefore=123456789");
            return Ok([]);
        };

        const user_query = `takenBefore=123456789`;
        await testing_SearchHandler(handler, user_query);

        const user_query_quotes = `takenBefore="123456789"`;
        await testing_SearchHandler(handler, user_query_quotes);
    });

    it("Should parse date", async () => {
        const handler = async (query: BodyInit): HandlerReturnType => {
            expect(query).toBe(`takenAfter=${819170640000 / 1000}`); // Converted UNIX time for that date
            return Ok([]);
        };

        const user_query = `takenAfter="1995-12-17T03:24:00Z"`;
        await testing_SearchHandler(handler, user_query);
    });

    it("Should drop invalid date tokens (date parse fails => token error => omitted)", async () => {
        const handler = async (query: BodyInit): HandlerReturnType => {
            // compileTokensToApiQuery skips tokens with t.error, so nothing gets sent
            expect(query).toBe("");
            return Ok([]);
        };

        const user_query = `takenAfter=FakeDate`;
        await testing_SearchHandler(handler, user_query);
    });

    it("Should handle sorting queries well", async () => {
        const handler = async (query: BodyInit): HandlerReturnType => {
            // compileTokensToApiQuery skips tokens with t.error, so nothing gets sent
            expect(query).toBe(`sortBy=DateAdded&sortOrder=ASC`);
            return Ok([]);
        };

        const user_query = `sortBy=DateAdded;sortOrder=ASC;`;
        await testing_SearchHandler(handler, user_query);
    });

    it("should pass integration tests", async () => {
        global.fetch = jest.fn(() => {
            return Promise.resolve({
                status: 200,
                ok: true,
                json: () => {
                    return {};
                },
            });
        }) as jest.Mock;

        const result = await SearchHandler("");
        expect(result).toStrictEqual(Ok({}));
    });
});
