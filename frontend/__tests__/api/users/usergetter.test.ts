/**
 * @jest-environment node
 */

import { Ok, Err } from "@/common/result";
import UserGetter from "@/api/users/usergetter";

describe("User Getter", () => {
    it("should pass full integration test", async () => {
        global.fetch = jest.fn((url) => {
            expect(url).toStrictEqual("/backend/users/1");
            return Promise.resolve({
                status: 200,
                ok: true,
                json: () => {
                    return { createdAt: 4, id: 1, username: "amogus" };
                },
            });
        }) as jest.Mock;

        const result = await UserGetter(1);
        expect(result).toStrictEqual(
            Ok({ createdAt: 4, id: 1, username: "amogus" }),
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

        const result = await UserGetter(2);
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

        const result = await UserGetter(5);
        expect(result).toStrictEqual(Err({ status: 404, message: "explode" }));
    });
});
