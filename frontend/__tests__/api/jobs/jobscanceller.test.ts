/**
 * @jest-environment node
 */

import JobsCanceller from "@/api/jobs/jobscanceller";

import { Ok, Err } from "@/common/result";

describe("Jobs Canceller", () => {
    it("should handle success well", async () => {
        global.fetch = jest.fn((url) => {
            expect(typeof url === "string").toBeTruthy();
            if (typeof url === "string") {
                expect(url.endsWith("/jobs/cancel/5")).toBeTruthy();
            }
            return Promise.resolve({
                status: 200,
                ok: true,
                json: () => {
                    return { uuid: "5" };
                },
            });
        }) as jest.Mock;

        const result = await JobsCanceller("5");
        expect(result).toStrictEqual(Ok({ uuid: "5" }));
    });

    it("should handle 400 error well", async () => {
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

        const result = await JobsCanceller("5");
        expect(result).toStrictEqual(Err({ status: 400, message: "failed" }));
    });

    it("should handle 404 error well", async () => {
        global.fetch = jest.fn(() => {
            return Promise.resolve({
                status: 404,
                ok: false,
                text: () => {
                    return "broke";
                },
            });
        }) as jest.Mock;

        const result = await JobsCanceller("5");
        expect(result).toStrictEqual(Err({ status: 404, message: "broke" }));
    });
});
