/**
 * @jest-environment node
 */

import JobsLister from "@/api/jobs/jobslister";

import { Ok, Err } from "@/common/result";

describe("Jobs Handler", () => {
    it("should handle summary list well", async () => {
        global.fetch = jest.fn(() => {
            return Promise.resolve({
                status: 200,
                ok: true,
                json: () => {
                    return [
                        {
                            complete: 1,
                            errors: 1,
                            in_progress: 0,
                            uuid: "5",
                            status: "done",
                            title: "upload",
                        },
                        {
                            complete: 0,
                            errors: 0,
                            in_progress: 5,
                            uuid: "6",
                            status: "pending",
                            title: "plugin",
                        },
                    ];
                },
            });
        }) as jest.Mock;

        const result = await JobsLister();
        expect(result).toStrictEqual(
            Ok([
                {
                    complete: 1,
                    errors: 1,
                    in_progress: 0,
                    uuid: "5",
                    status: "done",
                    title: "upload",
                },
                {
                    complete: 0,
                    errors: 0,
                    in_progress: 5,
                    uuid: "6",
                    status: "pending",
                    title: "plugin",
                },
            ]),
        );
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

        const result = await JobsLister();
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

        const result = await JobsLister();
        expect(result).toStrictEqual(Err({ status: 404, message: "broke" }));
    });
});
