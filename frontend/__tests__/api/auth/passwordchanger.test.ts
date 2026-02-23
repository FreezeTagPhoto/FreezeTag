/**
 * @jest-environment node
 */

import PasswordChanger, {
    testing_PasswordChanger,
    type testing_PasswordChangeResponse,
} from "@/api/auth/passwordchanger";
import { Ok, Err, type Result } from "@/common/result";
import type { RequestError } from "@/api/common/apihandler";

type Handler = (
    data: BodyInit,
) => Promise<Result<testing_PasswordChangeResponse, RequestError>>;

type BlobLike = { text: () => Promise<string> };

function isBlobLike(value: unknown): value is BlobLike {
    return (
        typeof value === "object" &&
        value !== null &&
        "text" in value &&
        typeof (value as { text?: unknown }).text === "function"
    );
}

type MinimalFormData = { get: (name: string) => unknown };

function makeEvent(
    current_password?: unknown,
    new_password?: unknown,
): MinimalFormData {
    return {
        get: (name: string) => {
            if (name === "current_password") return current_password;
            if (name === "new_password") return new_password;
            return null;
        },
    };
}

// If the test runtime doesn’t provide Blob (older Node/Jest setups), polyfill enough for this file.
class TestBlob {
    private readonly _text: string;
    public readonly type: string;

    constructor(parts: unknown[], opts?: { type?: string }) {
        this._text = parts
            .map((p) => (typeof p === "string" ? p : String(p)))
            .join("");
        this.type = opts?.type ?? "";
    }

    async text(): Promise<string> {
        return this._text;
    }
}

const globalWithBlob = globalThis as unknown as { Blob?: typeof Blob };
if (typeof globalWithBlob.Blob === "undefined") {
    globalWithBlob.Blob = TestBlob as unknown as typeof Blob;
}

function makeMockResponse(spec: {
    status: number;
    ok: boolean;
    json?: unknown;
    text?: string;
    jsonThrows?: boolean;
}): Response {
    const jsonImpl = spec.jsonThrows
        ? async () => {
              throw new Error("bad json");
          }
        : async () => spec.json;

    const textImpl = async () =>
        spec.text ??
        (typeof spec.json === "string"
            ? spec.json
            : JSON.stringify(spec.json ?? ""));

    return {
        status: spec.status,
        ok: spec.ok,
        json: jsonImpl,
        text: textImpl,
    } as unknown as Response;
}

function setFetchMock(response: Response): void {
    globalThis.fetch = jest.fn(async () => response) as unknown as typeof fetch;
}

describe("Password Changer", () => {
    const originalFetch = globalThis.fetch;

    afterEach(() => {
        jest.restoreAllMocks();
        globalThis.fetch = originalFetch;
    });

    it("should return Err(400) when password fields are missing (and not call handler)", async () => {
        const handler: jest.MockedFunction<Handler> = jest.fn(
            async (_data: BodyInit) => Ok({ message: "ok" }),
        );

        const event = makeEvent("old_password", undefined);
        const result = await testing_PasswordChanger(
            handler,
            event as unknown as FormData,
        );

        expect(handler).not.toHaveBeenCalled();

        expect(result.ok).toBe(false);
        if (!result.ok) {
            expect(result.error.status).toBe(400);
            expect(result.error.message).toBe("Missing password fields");
        }
    });

    it("should return Err(400) when a password field is not a string (and not call handler)", async () => {
        const handler: jest.MockedFunction<Handler> = jest.fn(
            async (_data: BodyInit) => Ok({ message: "ok" }),
        );

        const event = makeEvent("old_password", { not: "a string" });
        const result = await testing_PasswordChanger(
            handler,
            event as unknown as FormData,
        );

        expect(handler).not.toHaveBeenCalled();

        expect(result.ok).toBe(false);
        if (!result.ok) {
            expect(result.error.status).toBe(400);
            expect(result.error.message).toBe("Missing password fields");
        }
    });

    it("should call handler with the JSON blob body and pass through Ok", async () => {
        const handler: jest.MockedFunction<Handler> = jest.fn(
            async (_data: BodyInit) => Ok({ message: "Password changed" }),
        );

        const event = makeEvent("old_password", "new_password");
        const result = await testing_PasswordChanger(
            handler,
            event as unknown as FormData,
        );

        expect(handler).toHaveBeenCalledTimes(1);

        const body = handler.mock.calls[0]?.[0];
        expect(body).toBeTruthy();

        if (!isBlobLike(body)) {
            throw new Error("Expected Blob-like body with a .text() method");
        }

        const bodyText = await body.text();
        expect(JSON.parse(bodyText)).toEqual({
            current_password: "old_password",
            new_password: "new_password",
        });

        expect(result.ok).toBe(true);
        if (result.ok) {
            expect(result.value.message).toBe("Password changed");
        }
    });

    it("should map 400/401/500 JSON error bodies into Err with message", async () => {
        const response = makeMockResponse({
            status: 400,
            ok: false,
            json: { error: "bad request" },
        });

        const handler: jest.MockedFunction<Handler> = jest.fn(
            async (_data: BodyInit) =>
                Err({
                    status_code: 400,
                    response,
                }),
        );

        const event = makeEvent("old_password", "new_password");
        const result = await testing_PasswordChanger(
            handler,
            event as unknown as FormData,
        );

        expect(result.ok).toBe(false);
        if (!result.ok) {
            expect(result.error.status).toBe(400);
            expect(result.error.message).toBe("bad request");
        }
    });

    it("should fall back to response.text() if response.json() throws for 400/401/500", async () => {
        const response = makeMockResponse({
            status: 401,
            ok: false,
            jsonThrows: true,
            text: "plain text error",
        });

        const handler: jest.MockedFunction<Handler> = jest.fn(
            async (_data: BodyInit) =>
                Err({
                    status_code: 401,
                    response,
                }),
        );

        const event = makeEvent("old_password", "new_password");
        const result = await testing_PasswordChanger(
            handler,
            event as unknown as FormData,
        );

        expect(result.ok).toBe(false);
        if (!result.ok) {
            expect(result.error.status).toBe(401);
            expect(result.error.message).toBe("plain text error");
        }
    });

    it("should use response.text() for non-400/401/500 statuses", async () => {
        const response = makeMockResponse({
            status: 404,
            ok: false,
            text: "not found",
        });

        const handler: jest.MockedFunction<Handler> = jest.fn(
            async (_data: BodyInit) =>
                Err({
                    status_code: 404,
                    response,
                }),
        );

        const event = makeEvent("old_password", "new_password");
        const result = await testing_PasswordChanger(
            handler,
            event as unknown as FormData,
        );

        expect(result.ok).toBe(false);
        if (!result.ok) {
            expect(result.error.status).toBe(404);
            expect(result.error.message).toBe("not found");
        }
    });

    it("should pass full integration test (fetch ok)", async () => {
        setFetchMock(
            makeMockResponse({
                status: 200,
                ok: true,
                json: { message: "ok" },
            }),
        );

        const event = makeEvent("old_password", "new_password");
        const result = await PasswordChanger(event as unknown as FormData);

        expect(result.ok).toBe(true);
        if (result.ok) {
            expect(result.value.message).toBe("ok");
        }
    });

    it("should handle 400 integration errors (fetch not ok, json error body)", async () => {
        setFetchMock(
            makeMockResponse({
                status: 400,
                ok: false,
                json: { error: "bad" },
            }),
        );

        const event = makeEvent("old_password", "new_password");
        const result = await PasswordChanger(event as unknown as FormData);

        expect(result.ok).toBe(false);
        if (!result.ok) {
            expect(result.error.status).toBe(400);
            expect(result.error.message).toBe("bad");
        }
    });

    it("should handle 404 integration errors (fetch not ok, text body)", async () => {
        setFetchMock(
            makeMockResponse({
                status: 404,
                ok: false,
                text: "sucks",
            }),
        );

        const event = makeEvent("old_password", "new_password");
        const result = await PasswordChanger(event as unknown as FormData);

        expect(result.ok).toBe(false);
        if (!result.ok) {
            expect(result.error.status).toBe(404);
            expect(result.error.message).toBe("sucks");
        }
    });
});
