/**
 * @jest-environment node
 */

import { RequestError } from "@/api/common/apihandler";
import {
    testing_TagGetResponse,
    testing_TagGetter,
} from "@/api/tags/taggetter";

import TagGetter from "@/api/tags/taggetter";

import { Result, Ok, Err } from "@/common/result";

type HandlerReturnType = Promise<Result<testing_TagGetResponse, RequestError>>;

describe("Tag Getter", () => {
    it("should not include a leading slash without an image id", () => {
        const handler = async (query: BodyInit): HandlerReturnType => {
            expect(query).toBe("");
            return Ok([]);
        };

        testing_TagGetter(handler);
    });

    it("should include a leading slash without an image id", () => {
        const handler = async (query: BodyInit): HandlerReturnType => {
            expect(query).toBe("/67");
            return Ok([]);
        };

        testing_TagGetter(handler, 67);
    });

    it("should get message on 400", async () => {
        const handler = async (_: BodyInit): HandlerReturnType => {
            const response = new Response('{"error": "true"}');
            return Err({
                status_code: 400,
                response,
            });
        };

        const result = await testing_TagGetter(handler);
        expect(result).toStrictEqual(Err({ status: 400, message: "true" }));
    });

    it("should percolate status code on failure", async () => {
        const handler = async (_: BodyInit): HandlerReturnType => {
            return Err({ status_code: 404, response: new Response() });
        };

        const result = await testing_TagGetter(handler);
        expect(result).toStrictEqual(
            Err({ status: 404, message: await new Response().text() }),
        );
    });

    it("should receive tags in order", async () => {
        const handler = async (_: BodyInit): HandlerReturnType => {
            return Ok(["arches", "wedding", "banquet"]);
        };

        const result = await testing_TagGetter(handler);
        expect(result).toStrictEqual(Ok(["arches", "wedding", "banquet"]));
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

        const result = await TagGetter(0);
        expect(result).toStrictEqual(Ok({}));
    });
});
