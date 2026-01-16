/**
 * @jest-environment node
 */

import {
    testing_JobsHandler,
    testing_JobResponse,
} from "@/api/jobs/jobshandler";

import { RequestError } from "@/api/common/apihandler";

import { Result, Ok, Err } from "@/common/result";

type HandlerReturnType = Promise<Result<testing_JobResponse, RequestError>>;

describe("Jobs Handler", () => {
    it("Should get progress amount", async () => {
        const handler = async (_: BodyInit): HandlerReturnType => {
            return Ok({
                in_progress: [{ name: "gopher.png", status: "almost!" }],
                results: [
                    {
                        success: {
                            filename: "mocha.png",
                            id: 1,
                        },
                        error: undefined,
                    },
                ],
                uuid: "uuid",
            });
        };

        const result = await testing_JobsHandler(handler, "uuid");
        expect(result.ok).toBeTruthy();

        if (result.ok) {
            expect(result.value.ok).toBeFalsy();
            if (!result.value.ok) {
                expect(result.value.error).toBe(0.5);
            }
        }
    });

    it("Should get completed images", async () => {
        const handler = async (_: BodyInit): HandlerReturnType => {
            return Ok({
                in_progress: [],
                results: [
                    {
                        success: {
                            filename: "mocha.png",
                            id: 1,
                        },
                        error: undefined,
                    },
                    {
                        success: undefined,
                        error: {
                            filename: "gopher.png",
                            reason: "gopher died",
                        },
                    },
                ],
                uuid: "uuid",
            });
        };

        const result = await testing_JobsHandler(handler, "uuid");
        expect(result.ok).toBeTruthy();

        if (result.ok) {
            expect(result.value.ok).toBeTruthy();
            if (result.value.ok) {
                const map = result.value.value;
                expect(map.size).toBe(2);
                expect(map.get("gopher.png")).toStrictEqual(Err("gopher died"));
                expect(map.get("mocha.png")).toStrictEqual(Ok(1));
            }
        }
    });

    it("Should skip bad images", async () => {
        // Mostly a symptom of the API datatype being temporarily incorrect
        const handler = async (_: BodyInit): HandlerReturnType => {
            return Ok({
                in_progress: [],
                results: [
                    {
                        success: undefined,
                        error: undefined,
                    },
                ],
                uuid: "uuid",
            });
        };

        const spy = jest.spyOn(console, "error").mockImplementation(() => {});
        const result = await testing_JobsHandler(handler, "uuid");
        expect(result.ok).toBeTruthy();
        expect(spy).toHaveBeenCalled();

        if (result.ok) {
            expect(result.value.ok).toBeTruthy();
            if (result.value.ok) {
                const map = result.value.value;
                expect(map.size).toBe(0);
            }
        }
    });

    it("Should percolate status code on failure", async () => {
        const handler = async (_: BodyInit): HandlerReturnType => {
            return Err({ status_code: 404, response: new Response() });
        };

        const result = await testing_JobsHandler(handler, "uuid");
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

        const result = await testing_JobsHandler(handler, "uuid");
        expect(result).toStrictEqual(Err({ status: 400, message: "true" }));
    });
});
