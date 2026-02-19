/**
 * @jest-environment node
 */
import AuthChecker from "@/api/auth/authchecker";
import { Some } from "@/common/option";

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

        expect(result.some).toBeFalsy();
    });

    it("should handle granted perms well", async () => {
        global.fetch = jest.fn(() => {
            return Promise.resolve({
                status: 200,
                ok: true,
                json: () => {
                    return {
                        user_id: 0,
                        permissions: [
                            {
                                permission: "delete",
                                name: "Delete",
                                description: "Deletes",
                            },
                            {
                                permission: "push",
                                name: "Push",
                                description: "Pushes",
                            },
                        ],
                    };
                },
            });
        }) as jest.Mock;

        const result = await AuthChecker("delete");
        expect(result).toStrictEqual(
            Some({
                user_id: 0,
                permissions: [
                    {
                        permission: "delete",
                        name: "Delete",
                        description: "Deletes",
                    },
                    {
                        permission: "push",
                        name: "Push",
                        description: "Pushes",
                    },
                ],
            }),
        );
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
        expect(result.some).toBeFalsy();
    });
});
