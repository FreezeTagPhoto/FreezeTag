/**
 * @jest-environment node
 */
import AuthChecker from "@/api/auth/authchecker";

describe("Auth Checker", () => {
    it("should pass full integration test", async () => {
        global.fetch = jest.fn(() => {
            return Promise.resolve({
                status: 200,
                ok: true,
                json: () => {
                    return { user_id: 0 };
                },
            });
        }) as jest.Mock;

        const result = await AuthChecker();
        expect(result).toBeTruthy();
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

        const result = await AuthChecker();

        expect(result).toBeFalsy();
    });

    it("should handle granted perms well", async () => {
        global.fetch = jest.fn(() => {
            return Promise.resolve({
                status: 200,
                ok: true,
                json: () => {
                    return { user_id: 0, permissions: ["delete", "push"] };
                },
            });
        }) as jest.Mock;

        const result = await AuthChecker("delete");
        expect(result).toBeTruthy();
    });

    it("should handle ungranted perms well", async () => {
        global.fetch = jest.fn(() => {
            return Promise.resolve({
                status: 200,
                ok: true,
                json: () => {
                    return { user_id: 0, permissions: ["delete", "push"] };
                },
            });
        }) as jest.Mock;

        const result = await AuthChecker("read");
        expect(result).toBeFalsy();
    });
});
