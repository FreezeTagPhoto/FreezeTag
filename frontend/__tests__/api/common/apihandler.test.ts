/**
 * @jest-environment node
 */

import "@testing-library/jest-dom";
import { ApiHandler, Method } from "@/api/common/apihandler";

describe("API Handler", () => {
    it("can receive a response", async () => {
        const handler = ApiHandler(
            "https://jsonplaceholder.typicode.com/todos/1",
        )(Method.GET);
        const response = await handler("");

        expect(response.ok).toBeTruthy();
        if (response.ok) {
            expect(response.value).not.toBeNull();
        }
    });

    it("properly handles a 404", async () => {
        const handler = ApiHandler("http://google.com/free-ice-cream")(
            Method.GET,
        );
        const response = await handler("");

        expect(response.ok).toBeFalsy();
        if (!response.ok) {
            expect(response.error.status_code).toBe(404);
        }
    });

    it("properly handles a 405 response", async () => {
        const handler = ApiHandler("http://www.google.com")(Method.POST);
        const response = await handler("{status: 'good'}");

        expect(response.ok).toBeFalsy();
        if (!response.ok) {
            expect(response.error.status_code).toBe(405);
        }
    });

    it("properly handles a failed request (no server)", async () => {
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
