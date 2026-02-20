/**
 * @jest-environment node
 */

import PermsAdder from "@/api/permissions/permsadder";
import { Ok, Err } from "@/common/result";

describe("Perms Adder", () => {
    it("should pass full integration test", async () => {
        global.fetch = jest.fn((url) => {
            expect(url).toStrictEqual(
                "/backend/users/permissions/1?permission=post&permission=read",
            );
            return Promise.resolve({
                status: 200,
                ok: true,
                json: () => {
                    return { message: "cool" };
                },
            });
        }) as jest.Mock;

        const result = await PermsAdder(1, ["post", "read"]);
        expect(result).toStrictEqual(Ok({ message: "cool" }));
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

        const result = await PermsAdder(2, []);
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

        const result = await PermsAdder(5, ["delete"]);
        expect(result).toStrictEqual(Err({ status: 404, message: "explode" }));
    });
});
