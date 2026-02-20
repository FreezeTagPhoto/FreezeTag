/**
 * @jest-environment node
 */

import PermsGetter from "@/api/permissions/permsgetter";
import { Ok, Err } from "@/common/result";

describe("Perms Getter", () => {
    it("should pass full integration test", async () => {
        global.fetch = jest.fn((url) => {
            expect(url).toStrictEqual("/backend/users/permissions/1");
            return Promise.resolve({
                status: 200,
                ok: true,
                json: () => {
                    return [];
                },
            });
        }) as jest.Mock;

        const result = await PermsGetter(1);
        expect(result).toStrictEqual(Ok([]));
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

        const result = await PermsGetter(2);
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

        const result = await PermsGetter(5);
        expect(result).toStrictEqual(Err({ status: 404, message: "explode" }));
    });

    it("should handle null well", async () => {
        global.fetch = jest.fn((url) => {
            expect(url).toStrictEqual("/backend/users/permissions/5");
            return Promise.resolve({
                status: 200,
                ok: true,
                json: () => {
                    return undefined;
                },
            });
        }) as jest.Mock;

        const result = await PermsGetter(5);
        expect(result).toStrictEqual(Ok([]));
    });
});
