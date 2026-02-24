/**
 * @jest-environment node
 */

import { RequestError } from "@/api/common/apihandler";
import {
    testing_TagDeleteResponse,
    testing_TagDeleter,
} from "@/api/tags/tagdeleter";

import TagDeleter from "@/api/tags/tagdeleter";

import { Result, Ok, Err } from "@/common/result";

type HandlerReturnType = Promise<
    Result<testing_TagDeleteResponse, RequestError>
>;

describe("Tag Deleter", () => {
    it("should print out query string correctly", async () => {
        const handler = async (query: BodyInit): HandlerReturnType => {
            expect(typeof query === "string").toBeTruthy();

            if (typeof query === "string") {
                const components = query.split("&");

                // ensure we only send tag=... params, properly encoded
                expect(components.includes("tag=a%20b%20c")).toBeTruthy();
                expect(components.includes("tag=wedding")).toBeTruthy();
                expect(components.includes("tag=67")).toBeTruthy();
                expect(components.includes("tag=a%2Bb%3Fc")).toBeTruthy();

                for (const c of components) {
                    expect(c.startsWith("tag=")).toBeTruthy();
                }
            }

            return Ok({ deleted: 4 });
        };

        const result = await testing_TagDeleter(handler, [
            "a b c",
            "wedding",
            "67",
            "a+b?c",
        ]);
        expect(result).toStrictEqual(Ok(4));
    });

    it("should get message on 400 (json {error})", async () => {
        const handler = async (_: BodyInit): HandlerReturnType => {
            const response = new Response('{"error": "true"}');
            return Err({
                status_code: 400,
                response,
            });
        };

        const result = await testing_TagDeleter(handler, []);
        expect(result).toStrictEqual(Err({ status: 400, message: "true" }));
    });

    it("should fall back to response.text() if error body is not json", async () => {
        const handler = async (_: BodyInit): HandlerReturnType => {
            const response = new Response("plain text error");
            return Err({
                status_code: 500,
                response,
            });
        };

        const result = await testing_TagDeleter(handler, []);
        expect(result).toStrictEqual(
            Err({ status: 500, message: "plain text error" }),
        );
    });

    it("should percolate status code on failure", async () => {
        const handler = async (_: BodyInit): HandlerReturnType => {
            return Err({ status_code: 404, response: new Response() });
        };

        const result = await testing_TagDeleter(handler, []);
        expect(result).toStrictEqual(
            Err({ status: 404, message: await new Response().text() }),
        );
    });

    it("should pass integration tests", async () => {
        global.fetch = jest.fn(() => {
            return Promise.resolve({
                status: 200,
                ok: true,
                json: () => {
                    return { deleted: 2 };
                },
            });
        }) as jest.Mock;

        const result = await TagDeleter(["x"]);
        expect(result).toStrictEqual(Ok(2));
    });
});
