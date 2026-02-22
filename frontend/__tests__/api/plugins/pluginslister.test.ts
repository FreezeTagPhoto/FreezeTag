/**
 * @jest-environment node
 */

import { Ok, Err } from "@/common/result";
import PluginsLister from "@/api/plugins/pluginslister";

describe("Plugins Lister", () => {
    it("should pass full integration test", async () => {
        global.fetch = jest.fn((url) => {
            expect(url).toStrictEqual("/backend/plugins");
            return Promise.resolve({
                status: 200,
                ok: true,
                json: () => {
                    return [
                        {
                            name: "Test Plugin",
                            enabled: false,
                            version: "67",
                            hooks: {
                                tag_image: {
                                    type: "post_upload",
                                    signature: "single_image",
                                },
                            },
                        },
                    ];
                },
            });
        }) as jest.Mock;

        const result = await PluginsLister();
        expect(result).toStrictEqual(
            Ok([
                {
                    name: "Test Plugin",
                    enabled: false,
                    version: "67",
                    hooks: {
                        tag_image: {
                            type: "post_upload",
                            signature: "single_image",
                        },
                    },
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

        const result = await PluginsLister();
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

        const result = await PluginsLister();
        expect(result).toStrictEqual(Err({ status: 404, message: "explode" }));
    });
});
