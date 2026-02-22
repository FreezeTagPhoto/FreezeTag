/**
 * @jest-environment node
 */

import { Ok, Err } from "@/common/result";
import PluginsAbler from "@/api/plugins/pluginsabler";

describe("Plugins Lister", () => {
    it("should pass full integration test", async () => {
        global.fetch = jest.fn((url) => {
            expect(url).toStrictEqual(
                "/backend/plugins/enable?plugin=test&enabled=false",
            );
            return Promise.resolve({
                status: 200,
                ok: true,
                json: () => {
                    return {
                        disabled: true,
                    };
                },
            });
        }) as jest.Mock;

        const result = await PluginsAbler("test", false);
        expect(result).toStrictEqual(Ok({ disabled: true }));
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

        const result = await PluginsAbler("test", true);
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

        const result = await PluginsAbler("test", false);
        expect(result).toStrictEqual(Err({ status: 404, message: "explode" }));
    });
});
