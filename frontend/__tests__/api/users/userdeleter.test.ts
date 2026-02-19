/**
 * @jest-environment node
 */

import { Ok, Err } from "@/common/result";
import UserDeleter from "@/api/users/userdeleter";

describe("User Getter", () => {
    it("should pass full integration test", async () => {
        global.fetch = jest.fn((url) => {
            expect(url).toStrictEqual("/backend/users/1");
            return Promise.resolve({
                status: 200,
                ok: true,
                json: () => {
                    return { message: "cool" };
                },
            });
        }) as jest.Mock;

        const result = await UserDeleter(1);
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

        const result = await UserDeleter(2);
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

        const result = await UserDeleter(5);
        expect(result).toStrictEqual(Err({ status: 404, message: "explode" }));
    });
});
