/**
 * @jest-environment node
 */

import LogoutHandler from "@/api/auth/logouthandler";

describe("Logout Handler", () => {
    it("should pass full integration test", async () => {
        global.fetch = jest.fn(() => {
            return Promise.resolve({
                status: 200,
                ok: true,
                json: () => {
                    return { token: "sus" };
                },
            });
        }) as jest.Mock;

        await LogoutHandler();
    });

    it("should handle 400 errors well", async () => {
        global.fetch = jest.fn(() => {
            return Promise.resolve({
                status: 400,
                ok: false,
                json: () => {
                    return { error: "bad" };
                },
            });
        }) as jest.Mock;

        await LogoutHandler();
    });

    it("should handle 404 errors well", async () => {
        global.fetch = jest.fn(() => {
            return Promise.resolve({
                status: 404,
                ok: false,
                text: () => {
                    return "sucks";
                },
            });
        }) as jest.Mock;

        await LogoutHandler();
    });
});
