/**
 * @jest-environment node
 */

import {
    testing_UserCreator,
    testing_UserCreateResponse,
} from "@/api/auth/usercreator";

import { RequestError } from "@/api/common/apihandler";

import { Result, Ok, Err } from "@/common/result";

type HandlerReturnType = Promise<
    Result<testing_UserCreateResponse, RequestError>
>;

describe("User Creator", () => {
    it("Should get success data", async () => {
        const handler = async (_: BodyInit): HandlerReturnType => {
            return Ok({ username: "bob", id: 123, created_at: 154 });
        };

        const result = await testing_UserCreator(handler, new FormData());
        expect(result.ok).toBeTruthy();

        if (result.ok) {
            expect(result.value).toStrictEqual({
                username: "bob",
                id: 123,
                created_at: 154,
            });
        }
    });

    it("Should percolate status code on failure", async () => {
        const handler = async (_: BodyInit): HandlerReturnType => {
            return Err({ status_code: 404, response: new Response() });
        };

        const result = await testing_UserCreator(handler, new FormData());
        expect(result).toStrictEqual(
            Err({ status: 404, message: await new Response().text() }),
        );
    });

    it("Should get message on 400", async () => {
        const handler = async (_: BodyInit): HandlerReturnType => {
            const response = new Response('{"error": "true"}');
            return Err({
                status_code: 400,
                response,
            });
        };

        const result = await testing_UserCreator(handler, new FormData());
        expect(result).toStrictEqual(Err({ status: 400, message: "true" }));
    });
});
