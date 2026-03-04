/**
 * @jest-environment node
 */

import FileDownloader from "@/api/files/filedownloader";
import { Ok } from "@/common/result";

function makeResponseWithHeaders(
    blob: Blob,
    headers: Record<string, string>,
    status: number = 200,
): Response {
    return new Response(blob, {
        status,
        headers,
    });
}

describe("FileDownloader", () => {
    beforeEach(() => {
        jest.restoreAllMocks();
    });

    it("should default filename when content-disposition missing", async () => {
        const blob = new Blob(["hello"], { type: "image/jpeg" });
        const response = makeResponseWithHeaders(blob, {
            "content-type": "image/jpeg",
        });

        global.fetch = jest.fn(() => Promise.resolve(response)) as jest.Mock;

        const res = await FileDownloader(42)();
        expect(res.ok).toBeTruthy();
        if (res.ok) {
            expect(res.value.filename).toBe("image-42");
            expect(res.value.contentType).toBe("image/jpeg");
            expect(res.value.blob).toBeInstanceOf(Blob);
        }
    });

    it('should parse basic filename="..." from content-disposition', async () => {
        const blob = new Blob(["hello"], { type: "image/png" });
        const response = makeResponseWithHeaders(blob, {
            "content-type": "image/png",
            "content-disposition": 'attachment; filename="pic.png"',
        });

        global.fetch = jest.fn(() => Promise.resolve(response)) as jest.Mock;

        const res = await FileDownloader(1)();
        expect(res).toStrictEqual(
            Ok({
                blob,
                filename: "pic.png",
                contentType: "image/png",
            }),
        );
    });

    it("should parse UTF-8 filename* from content-disposition", async () => {
        const blob = new Blob(["hello"], { type: "application/octet-stream" });
        // "weird name 🧊.jpg" URL-encoded
        const encoded = encodeURIComponent("weird name 🧊.jpg");
        const response = makeResponseWithHeaders(blob, {
            "content-type": "application/octet-stream",
            "content-disposition": `attachment; filename*=UTF-8''${encoded}`,
        });

        global.fetch = jest.fn(() => Promise.resolve(response)) as jest.Mock;

        const res = await FileDownloader(5)();
        expect(res.ok).toBeTruthy();
        if (res.ok) {
            expect(res.value.filename).toBe("weird name 🧊.jpg");
            expect(res.value.contentType).toBe("application/octet-stream");
        }
    });

    it("should return Err on non-ok response", async () => {
        const response = new Response("bad", { status: 400 });

        global.fetch = jest.fn(() => Promise.resolve(response)) as jest.Mock;

        const res = await FileDownloader(3)();
        expect(res.ok).toBeFalsy();
        if (!res.ok) {
            expect(res.error.status_code).toBe(400);
            expect(res.error.response.status).toBe(400);
        }
    });

    it("should return Err on fetch throw", async () => {
        global.fetch = jest.fn(() => {
            throw new Error("network down");
        }) as jest.Mock;

        const res = await FileDownloader(7)();
        expect(res.ok).toBeFalsy();
        if (!res.ok) {
            expect(res.error.status_code).toBe(0);
            const txt = await res.error.response.text();
            expect(txt).toContain("Catastrophic Error");
        }
    });

    it("should call correct endpoint and headers", async () => {
        const blob = new Blob(["x"], { type: "image/jpeg" });
        const response = makeResponseWithHeaders(blob, {
            "content-type": "image/jpeg",
        });

        global.fetch = jest.fn(() => Promise.resolve(response)) as jest.Mock;

        await FileDownloader(99)();

        const [url, init] = (global.fetch as jest.Mock).mock.calls[0];
        expect(String(url)).toContain("/file/download/99");
        expect(init).toMatchObject({
            method: "GET",
            headers: { accept: "application/octet-stream" },
        });
    });
});
