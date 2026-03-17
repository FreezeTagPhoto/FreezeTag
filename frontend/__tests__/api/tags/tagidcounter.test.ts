/**
 * @jest-environment node
 */

import { RequestError } from "@/api/common/apihandler";
import {
    testing_TagCountResponse,
    testing_TagCounter,
} from "@/api/tags/tagidcounter";

import TagIdCounter from "@/api/tags/tagidcounter";

import { Result, Ok, Err } from "@/common/result";

type HandlerReturnType = Promise<
    Result<testing_TagCountResponse, RequestError>
>;

describe("Tag Counter", () => {
    it("should correctly form query string", () => {
        const handler = async (query: BodyInit): HandlerReturnType => {
            expect(query).toBe("id=1&id=10");
            return Ok({});
        };

        testing_TagCounter(handler, [1, 10]);
    });

    it("should get message on 400", async () => {
        const handler = async (_: BodyInit): HandlerReturnType => {
            const response = new Response('{"error": "true"}');
            return Err({
                status_code: 400,
                response,
            });
        };

        const result = await testing_TagCounter(handler, []);
        expect(result).toStrictEqual(Err({ status: 400, message: "true" }));
    });

    it("should percolate status code on failure", async () => {
        const handler = async (_: BodyInit): HandlerReturnType => {
            return Err({ status_code: 404, response: new Response() });
        };

        const result = await testing_TagCounter(handler, [1, 23]);
        expect(result).toStrictEqual(
            Err({ status: 404, message: await new Response().text() }),
        );
    });

    it("should receive tags", async () => {
        const handler = async (_: BodyInit): HandlerReturnType => {
            return Ok({ arches: 1, wedding: 5, banquet: 3 });
        };

        const result = await testing_TagCounter(handler, [1, 2, 3, 4, 5, 6]);
        expect(result).toStrictEqual(Ok({ arches: 1, wedding: 5, banquet: 3 }));
    });

    it("should pass integration tests", async () => {
        global.fetch = jest.fn(() => {
            return Promise.resolve({
                status: 200,
                ok: true,
                json: () => {
                    return { sus: 1 };
                },
            });
        }) as jest.Mock;

        const result = await TagIdCounter([1, 2, 3, 4, 5]);
        expect(result).toStrictEqual(Ok({ sus: 1 }));
    });
});
