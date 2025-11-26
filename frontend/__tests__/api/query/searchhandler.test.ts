/**
 * @jest-environment node
 */

import {
    testing_SearchHandler,
    testing_SearchResponse,
} from "@/api/query/searchhandler";

import { RequestError } from "@/api/common/apihandler";

import { Result, Ok, Err } from "@/common/result";

type HandlerReturnType = Promise<Result<testing_SearchResponse, RequestError>>;

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
                expect(components.includes("tagLike=among us")).toBeTruthy();
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
                expect(components.includes("model=Note 7")).toBeTruthy();
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

    it("Should have working near queries", async () => {
        const handler = async (query: BodyInit): HandlerReturnType => {
            expect(query).toBe("near=1,2,3");
            return Ok([]);
        };

        const user_query = `near=1,2,3`;
        await testing_SearchHandler(handler, user_query);
    });
});
