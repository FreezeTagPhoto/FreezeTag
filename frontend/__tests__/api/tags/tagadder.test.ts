/**
 * @jest-environment node
 */

import { RequestError } from "@/api/common/apihandler";
import { testing_TagAddResponse, testing_TagAdder } from "@/api/tags/tagadder";

import TagAdder from "@/api/tags/tagadder";

import { Result, Ok, Err } from "@/common/result";

type HandlerReturnType = Promise<Result<testing_TagAddResponse, RequestError>>;

describe("Tag Adder", () => {
    it("should print out query string correctly", () => {
        const handler = async (query: BodyInit): HandlerReturnType => {
            expect(typeof query === "string").toBeTruthy();
            if (typeof query === "string") {
                const components = query.split("&");
                expect(components.includes("id=1")).toBeTruthy();
                expect(components.includes("id=2")).toBeTruthy();
                expect(components.includes("id=3")).toBeTruthy();
                expect(components.includes("tag=sus")).toBeTruthy();
                expect(components.includes("tag=wedding")).toBeTruthy();
                expect(components.includes("tag=67")).toBeTruthy();
            }
            return Ok({ added: [], errors: [] });
        };

        testing_TagAdder(handler, [1, 2, 3], ["sus", "wedding", "67"]);
    });

    it("should get message on 400", async () => {
        const handler = async (_: BodyInit): HandlerReturnType => {
            const response = new Response('{"error": "true"}');
            return Err({
                status_code: 400,
                response,
            });
        };

        const result = await testing_TagAdder(handler, [], []);
        expect(result).toStrictEqual(Err({ status: 400, message: "true" }));
    });

    it("should percolate status code on failure", async () => {
        const handler = async (_: BodyInit): HandlerReturnType => {
            return Err({ status_code: 404, response: new Response() });
        };

        const result = await testing_TagAdder(handler, [], []);
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
                    return { errors: {} };
                },
            });
        }) as jest.Mock;

        const result = await TagAdder([], []);
        expect(result).toStrictEqual(Ok({}));
    });
});
