/**
 * @jest-environment node
 */

import { Ok, Err } from "@/common/result";
import PluginRunner from "@/api/plugins/pluginrunner";

describe("Plugins Runner", () => {
    it("should pass full integration test", async () => {
        global.fetch = jest.fn((url, extra) => {
            expect(url).toBe(
                "/backend/plugins/run?plugin=test_plugin&hook=test_hook",
            );
            expect(extra.body).toBe("67");
            return Promise.resolve({
                status: 200,
                ok: true,
                json: () => {
                    return "67uuid-420";
                },
            });
        }) as jest.Mock;

        const result = await PluginRunner("test_plugin", "test_hook", 67);
        expect(result).toStrictEqual(Ok("67uuid-420"));
    });

    it("should pass full integration test with multiple images", async () => {
        global.fetch = jest.fn((url, extra) => {
            expect(url).toBe(
                "/backend/plugins/run?plugin=test_plugin&hook=test_hook",
            );
            expect(extra.body).toBe("[67,420,69]");
            return Promise.resolve({
                status: 200,
                ok: true,
                json: () => {
                    return "67uuid-420";
                },
            });
        }) as jest.Mock;

        const result = await PluginRunner(
            "test_plugin",
            "test_hook",
            [67, 420, 69],
        );
        expect(result).toStrictEqual(Ok("67uuid-420"));
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

        const result = await PluginRunner("test", "test", 0);
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

        const result = await PluginRunner("test", "test", 0);
        expect(result).toStrictEqual(Err({ status: 404, message: "explode" }));
    });
});
