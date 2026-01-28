/**
 * @jest-environment jsdom
 */
import {
    testing_LoginHandler,
    testing_LoginResponse,
} from "@/api/auth/loginhandler";

import { RequestError } from "@/api/common/apihandler";

import { Result, Ok } from "@/common/result";

type HandlerReturnType = Promise<Result<testing_LoginResponse, RequestError>>;

describe("Login Handler", () => {
    it("Shouldn't get token, and should put it in localStorage", async () => {
        const setItemSpy = jest.spyOn(Storage.prototype, "setItem");

        const handler = async (_: BodyInit): HandlerReturnType => {
            return Ok({ token: "371fb38e-88c1-4fc7-b43d-be9ca67e4b51" });
        };

        const result = await testing_LoginHandler(handler, new FormData());
        expect(result.some).toBeFalsy();
        expect(setItemSpy).toHaveBeenCalledWith(
            "freezetag_token",
            "371fb38e-88c1-4fc7-b43d-be9ca67e4b51",
        );
    });

    // TODO: Figure out how to polyfill rejected tests (JSDom doesn't have Response/TextEncoder/ReadableStream for ???)
});
