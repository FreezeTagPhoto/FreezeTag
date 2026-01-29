/**
 * @jest-environment node
 */

import {
    testing_ImageUploader,
    testing_UploadResponse,
} from "@/api/upload/imageuploader";

import ImageUploader from "@/api/upload/imageuploader";

import { RequestError } from "@/api/common/apihandler";

import { Result, Ok, Err } from "@/common/result";

type HandlerReturnType = Promise<Result<testing_UploadResponse, RequestError>>;

describe("Image Uploader", () => {
    it("Should get job code", async () => {
        const handler = async (_: BodyInit): HandlerReturnType => {
            return Ok("371fb38e-88c1-4fc7-b43d-be9ca67e4b51");
        };

        const result = await testing_ImageUploader(handler, new FormData());
        expect(result.ok).toBeTruthy();

        if (result.ok) {
            expect(result.value).toStrictEqual(
                "371fb38e-88c1-4fc7-b43d-be9ca67e4b51",
            );
        }
    });

    it("Should percolate status code on failure", async () => {
        const handler = async (_: BodyInit): HandlerReturnType => {
            return Err({ status_code: 404, response: new Response() });
        };

        const result = await testing_ImageUploader(handler, new FormData());
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

        const result = await testing_ImageUploader(handler, new FormData());
        expect(result).toStrictEqual(Err({ status: 400, message: "true" }));
    });

    it("should pass integration tests", async () => {
        global.fetch = jest.fn(() => {
            return Promise.resolve({
                status: 200,
                ok: true,
                json: () => {
                    return "UUID";
                },
            });
        }) as jest.Mock;

        const result = await ImageUploader(new FormData());
        expect(result).toStrictEqual(Ok("UUID"));
    });
});
