/**
 * @jest-environment node
 */
import {
    testing_LoginResponse,
} from "@/api/auth/loginhandler";
import LoginHandler from "@/api/auth/loginhandler";

import { RequestError } from "@/api/common/apihandler";
import { None } from "@/common/option";

import { Result } from "@/common/result";

type _HandlerReturnType = Promise<Result<testing_LoginResponse, RequestError>>;

describe("Login Handler", () => {
    // TODO: Write Tests

    it("should pass full integration test", async () => {
        global.fetch = jest.fn(() => {
            return Promise.resolve({
                status: 200,
                ok: true,
                json: () => {
                    return { token: "sus" };
                },
            });
        }) as jest.Mock;

        const result = await LoginHandler(new FormData());
        expect(result).toStrictEqual(None());
    });
});
