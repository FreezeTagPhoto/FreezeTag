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
                    return { user_id: "sus" };
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
});
