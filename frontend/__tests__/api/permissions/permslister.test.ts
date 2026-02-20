/**
 * @jest-environment node
 */

import { Ok, Err } from "@/common/result";
import PermsLister from "@/api/permissions/permslister";

describe("Perms Lister", () => {
    it("should pass full integration test", async () => {
        global.fetch = jest.fn((url) => {
            expect(url).toStrictEqual("/backend/permissions/list");
            return Promise.resolve({
                status: 200,
                ok: true,
                json: () => {
                    return [
                        {
                            permission: "read:user",
                            name: "Read Users",
                            description: "lets you read users",
                        },
                        {
                            permission: "write:user",
                            name: "Write Users",
                            description: "lets you write users",
                        },
                    ];
                },
            });
        }) as jest.Mock;

        const result = await PermsLister();
        expect(result).toStrictEqual(
            Ok([
                {
                    permission: "read:user",
                    name: "Read Users",
                    description: "lets you read users",
                },
                {
                    permission: "write:user",
                    name: "Write Users",
                    description: "lets you write users",
                },
            ]),
        );
    });

    it("should handle 400 well", async () => {
        global.fetch = jest.fn((_) => {
            return Promise.resolve({
                status: 400,
                ok: false,
                json: () => {
                    return { error: "explode" };
                },
            });
        }) as jest.Mock;

        const result = await PermsLister();
        expect(result).toStrictEqual(Err({ status: 400, message: "explode" }));
    });

    it("should handle 404 well", async () => {
        global.fetch = jest.fn((_) => {
            return Promise.resolve({
                status: 404,
                ok: false,
                text: () => {
                    return "explode";
                },
            });
        }) as jest.Mock;

        const result = await PermsLister();
        expect(result).toStrictEqual(Err({ status: 404, message: "explode" }));
    });
});
