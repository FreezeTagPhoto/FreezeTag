/**
 * @jest-environment node
 */

import "@testing-library/jest-dom";
import { ApiHandler, Method } from "@/api/common/apihandler";

describe("API Handler", () => {
    it("can receive a response", async () => {
        global.fetch = jest.fn((request: string) => {
            expect(request).toStrictEqual(
                "http://good_url.test/valid_endpoint",
            );
            return Promise.resolve({
                status: 200,
                ok: true,
                json: () => {
                    return { sus: "sus" };
                },
            });
        }) as jest.Mock;

        const handler = ApiHandler("http://good_url.test/")(Method.GET);
        const response = await handler("valid_endpoint");

        expect(response.ok).toBeTruthy();
        if (response.ok) {
            expect(response.value).toStrictEqual({ sus: "sus" });
        }
    });

    it("properly handles a 404", async () => {
        global.fetch = jest.fn(() =>
            Promise.resolve({
                status: 404,
                ok: false,
            }),
        ) as jest.Mock;

        const handler = ApiHandler("http://good_url.test/")(Method.GET);
        const response = await handler("bad_endpoint");

        expect(response.ok).toBeFalsy();
        if (!response.ok) {
            expect(response.error.status_code).toBe(404);
        }
    });

    it("properly can post request", async () => {
        global.fetch = jest.fn((request: string, init: RequestInit) => {
            expect(request).toBe("http://good_url.test/");
            expect(init.body).toStrictEqual(new FormData());
            return Promise.resolve({
                status: 200,
                ok: true,
                json: () => {
                    return { sus: "sus" };
                },
            });
        }) as jest.Mock;

        const handler = ApiHandler("http://good_url.test/", true)(Method.POST);
        const response = await handler(new FormData());

        expect(response.ok).toBeTruthy();
        if (response.ok) {
            expect(response.value).toStrictEqual({ sus: "sus" });
        }
    });

    it("handles bad post request", async () => {
        global.fetch = jest.fn((request: string, init: RequestInit) => {
            expect(request).toBe("http://good_url.test/bad_endpoint");
            expect(init.body).toStrictEqual(new FormData());
            return Promise.resolve({
                status: 405,
                ok: false,
                json: () => {
                    return { sus: "sus" };
                },
            });
        }) as jest.Mock;

        const handler = ApiHandler(
            "http://good_url.test/bad_endpoint",
            true,
        )(Method.POST);
        const response = await handler(new FormData());

        expect(response.ok).toBeFalsy();
        if (!response.ok) {
            expect(response.error.response.json()).toStrictEqual({
                sus: "sus",
            });
            expect(response.error.status_code).toBe(405);
        }
    });

    it("properly handles a failed request (no server)", async () => {
        global.fetch = jest.fn(() => Promise.reject("error")) as jest.Mock;

        const handler = ApiHandler("http://fake_website.fakefakefakefake/")(
            Method.GET,
        );
        const response = await handler("sus");

        expect(response.ok).toBeFalsy();
        if (!response.ok) {
            expect(response.error.status_code).toBe(0);
        }
    });

    it("properly handles a failed request (no server) for post", async () => {
        global.fetch = jest.fn(() => Promise.reject("error")) as jest.Mock;

        const handler = ApiHandler("http://fake_website.fakefakefakefake/")(
            Method.POST,
        );
        const response = await handler("sus");

        expect(response.ok).toBeFalsy();
        if (!response.ok) {
            expect(response.error.status_code).toBe(0);
        }
    });
});
