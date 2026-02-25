/**
 * @jest-environment node
 */

import JobsDetailer from "@/api/jobs/jobsdetailer";

import { Ok, Err } from "@/common/result";

describe("Jobs Detailer", () => {
    it("should handle complete well", async () => {
        global.fetch = jest.fn(() => {
            return Promise.resolve({
                status: 200,
                ok: true,
                json: () => {
                    return {
                        completed: [{ filename: "sus", id: 1 }],
                        failed: [{ filename: "amogus", reason: "bad meme" }],
                        uuid: "5",
                        cancelled: false,
                    };
                },
            });
        }) as jest.Mock;

        const result = await JobsDetailer("5");
        expect(result).toStrictEqual(
            Ok({
                completed: [{ filename: "sus", id: 1 }],
                failed: [{ filename: "amogus", reason: "bad meme" }],
                uuid: "5",
                cancelled: false,
            }),
        );
    });

    it("should handle details error well", async () => {
        global.fetch = jest.fn(() => {
            return Promise.resolve({
                status: 400,
                ok: false,
                json: () => {
                    return {
                        error: "failed",
                    };
                },
            });
        }) as jest.Mock;

        const result = await JobsDetailer("5");
        expect(result).toStrictEqual(Err({ status: 400, message: "failed" }));
    });

    it("should handle non-400 details error well", async () => {
        global.fetch = jest.fn(() => {
            return Promise.resolve({
                status: 404,
                ok: false,
                text: () => {
                    return "failed";
                },
            });
        }) as jest.Mock;

        const result = await JobsDetailer("5");
        expect(result).toStrictEqual(Err({ status: 404, message: "failed" }));
    });

    it("should handle empty job well", async () => {
        global.fetch = jest.fn(() => {
            return Promise.resolve({
                status: 200,
                ok: true,
                json: () => {
                    return {
                        uuid: "5",
                        cancelled: false,
                    };
                },
            });
        }) as jest.Mock;

        const result = await JobsDetailer("5");
        expect(result).toStrictEqual(
            Ok({
                uuid: "5",
                cancelled: false,
            }),
        );
    });
});
