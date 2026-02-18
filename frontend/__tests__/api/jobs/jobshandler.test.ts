/**
 * @jest-environment node
 */

import JobsHandler from "@/api/jobs/jobshandler";

import { Ok, Err } from "@/common/result";

describe("Jobs Handler", () => {
    it("should handle complete well", async () => {
        global.fetch = jest.fn((body) => {
            if (typeof body === "string" && body.includes("summary")) {
                return Promise.resolve({
                    status: 200,
                    ok: true,
                    json: () => {
                        return {
                            complete: 1,
                            errors: 1,
                            in_progress: 0,
                            uuid: "5",
                        };
                    },
                });
            } else if (typeof body === "string" && body.includes("details")) {
                return Promise.resolve({
                    status: 200,
                    ok: true,
                    json: () => {
                        return {
                            completed: [{ filename: "sus", id: 1 }],
                            failed: [
                                { filename: "amogus", reason: "bad meme" },
                            ],
                            uuid: "5",
                            cancelled: false,
                        };
                    },
                });
            } else {
                expect(false).toBeTruthy();
            }
        }) as jest.Mock;

        const expected_map = new Map();
        expected_map.set("sus", Ok(1));
        expected_map.set("amogus", Err("bad meme"));

        const result = await JobsHandler("5");
        expect(result).toStrictEqual(Ok(Ok(expected_map)));
    });

    it("should handle summary error well", async () => {
        global.fetch = jest.fn((body) => {
            if (typeof body === "string" && body.includes("summary")) {
                return Promise.resolve({
                    status: 400,
                    ok: false,
                    json: () => {
                        return {
                            error: "failed",
                        };
                    },
                });
            } else if (typeof body === "string" && body.includes("details")) {
                return Promise.resolve({
                    status: 200,
                    ok: true,
                    json: () => {
                        return {
                            completed: [{ filename: "sus", id: 1 }],
                            failed: [
                                { filename: "amogus", reason: "bad meme" },
                            ],
                            uuid: "5",
                            cancelled: false,
                        };
                    },
                });
            } else {
                expect(false).toBeTruthy();
            }
        }) as jest.Mock;

        const result = await JobsHandler("5");
        expect(result).toStrictEqual(Err({ status: 400, message: "failed" }));
    });

    it("should handle non-400 summary error well", async () => {
        global.fetch = jest.fn((body) => {
            if (typeof body === "string" && body.includes("summary")) {
                return Promise.resolve({
                    status: 404,
                    ok: false,
                    text: () => {
                        return "failed";
                    },
                });
            } else if (typeof body === "string" && body.includes("details")) {
                return Promise.resolve({
                    status: 200,
                    ok: true,
                    json: () => {
                        return {
                            completed: [{ filename: "sus", id: 1 }],
                            failed: [
                                { filename: "amogus", reason: "bad meme" },
                            ],
                            uuid: "5",
                            cancelled: false,
                        };
                    },
                });
            } else {
                expect(false).toBeTruthy();
            }
        }) as jest.Mock;

        const result = await JobsHandler("5");
        expect(result).toStrictEqual(Err({ status: 404, message: "failed" }));
    });

    it("should handle incomplete well", async () => {
        global.fetch = jest.fn((body) => {
            if (typeof body === "string" && body.includes("summary")) {
                return Promise.resolve({
                    status: 200,
                    ok: true,
                    json: () => {
                        return {
                            complete: 1,
                            errors: 0,
                            in_progress: 1,
                            uuid: "5",
                        };
                    },
                });
            } else if (typeof body === "string" && body.includes("details")) {
                return Promise.resolve({
                    status: 200,
                    ok: true,
                    json: () => {
                        return {
                            completed: [{ filename: "sus", id: 1 }],
                            failed: [
                                { filename: "amogus", reason: "bad meme" },
                            ],
                            uuid: "5",
                            cancelled: false,
                        };
                    },
                });
            } else {
                expect(false).toBeTruthy();
            }
        }) as jest.Mock;

        const result = await JobsHandler("5");
        expect(result).toStrictEqual(Ok(Err(0.5)));
    });

    it("should handle summary error well", async () => {
        global.fetch = jest.fn((body) => {
            if (typeof body === "string" && body.includes("summary")) {
                return Promise.resolve({
                    status: 400,
                    ok: false,
                    json: () => {
                        return {
                            error: "failed",
                        };
                    },
                });
            } else if (typeof body === "string" && body.includes("details")) {
                return Promise.resolve({
                    status: 200,
                    ok: true,
                    json: () => {
                        return {
                            completed: [{ filename: "sus", id: 1 }],
                            failed: [
                                { filename: "amogus", reason: "bad meme" },
                            ],
                            uuid: "5",
                            cancelled: false,
                        };
                    },
                });
            } else {
                expect(false).toBeTruthy();
            }
        }) as jest.Mock;

        const result = await JobsHandler("5");
        expect(result).toStrictEqual(Err({ status: 400, message: "failed" }));
    });

    it("should handle non-400 summary error well", async () => {
        global.fetch = jest.fn((body) => {
            if (typeof body === "string" && body.includes("summary")) {
                return Promise.resolve({
                    status: 404,
                    ok: false,
                    text: () => {
                        return "failed";
                    },
                });
            } else if (typeof body === "string" && body.includes("details")) {
                return Promise.resolve({
                    status: 200,
                    ok: true,
                    json: () => {
                        return {
                            completed: [{ filename: "sus", id: 1 }],
                            failed: [
                                { filename: "amogus", reason: "bad meme" },
                            ],
                            uuid: "5",
                            cancelled: false,
                        };
                    },
                });
            } else {
                expect(false).toBeTruthy();
            }
        }) as jest.Mock;

        const result = await JobsHandler("5");
        expect(result).toStrictEqual(Err({ status: 404, message: "failed" }));
    });

    it("should handle details error well", async () => {
        global.fetch = jest.fn((body) => {
            if (typeof body === "string" && body.includes("summary")) {
                return Promise.resolve({
                    status: 200,
                    ok: true,
                    json: () => {
                        return {
                            complete: 1,
                            errors: 1,
                            in_progress: 0,
                            uuid: "5",
                        };
                    },
                });
            } else if (typeof body === "string" && body.includes("details")) {
                return Promise.resolve({
                    status: 400,
                    ok: false,
                    json: () => {
                        return {
                            error: "failed",
                        };
                    },
                });
            } else {
                expect(false).toBeTruthy();
            }
        }) as jest.Mock;

        const result = await JobsHandler("5");
        expect(result).toStrictEqual(Err({ status: 400, message: "failed" }));
    });

    it("should handle non-400 details error well", async () => {
        global.fetch = jest.fn((body) => {
            if (typeof body === "string" && body.includes("summary")) {
                return Promise.resolve({
                    status: 200,
                    ok: true,
                    json: () => {
                        return {
                            complete: 1,
                            errors: 1,
                            in_progress: 0,
                            uuid: "5",
                        };
                    },
                });
            } else if (typeof body === "string" && body.includes("details")) {
                return Promise.resolve({
                    status: 404,
                    ok: false,
                    text: () => {
                        return "failed";
                    },
                });
            } else {
                expect(false).toBeTruthy();
            }
        }) as jest.Mock;

        const result = await JobsHandler("5");
        expect(result).toStrictEqual(Err({ status: 404, message: "failed" }));
    });

    it("should handle empty job well", async () => {
        global.fetch = jest.fn((body) => {
            if (typeof body === "string" && body.includes("summary")) {
                return Promise.resolve({
                    status: 200,
                    ok: true,
                    json: () => {
                        return {
                            complete: 0,
                            errors: 0,
                            in_progress: 0,
                            uuid: "5",
                        };
                    },
                });
            } else if (typeof body === "string" && body.includes("details")) {
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
            } else {
                expect(false).toBeTruthy();
            }
        }) as jest.Mock;

        const expected_map = new Map();

        const result = await JobsHandler("5");
        expect(result).toStrictEqual(Ok(Ok(expected_map)));
    });
});
