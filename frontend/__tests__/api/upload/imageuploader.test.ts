/**
 * @jest-environment node
 */

import {
    testing_ImageUploader,
    testing_UploadResponse,
} from "@/api/upload/imageuploader";

import { RequestError } from "@/api/common/apihandler";

import { Result, Ok, Err } from "@/common/result";

type HandlerReturnType = Promise<Result<testing_UploadResponse, RequestError>>;

describe("Image Uploader", () => {
    it("Should get all successful images", async () => {
        const handler = async (_: BodyInit): HandlerReturnType => {
            return Ok({
                uploaded: [
                    { id: 100, filename: "gopher.png" },
                    { id: 67, filename: "coffee.jpeg" },
                ],
                errors: [],
            });
        };

        const result = await testing_ImageUploader(handler, new FormData());
        expect(result.ok).toBeTruthy();

        if (result.ok) {
            const map = result.value;
            expect(map.get("gopher.png")).toStrictEqual(Ok(100));
            expect(map.get("coffee.jpeg")).toStrictEqual(Ok(67));
        }
    });

    it("Should get all failed images", async () => {
        const handler = async (_: BodyInit): HandlerReturnType => {
            return Ok({
                uploaded: [],
                errors: [
                    {
                        reason: "Gopher died on the way",
                        filename: "gopher.png",
                    },
                    { reason: "Coffee spilled", filename: "coffee.jpeg" },
                ],
            });
        };

        const result = await testing_ImageUploader(handler, new FormData());
        expect(result.ok).toBeTruthy();

        if (result.ok) {
            const map = result.value;
            expect(map.get("gopher.png")).toStrictEqual(
                Err("Gopher died on the way"),
            );
            expect(map.get("coffee.jpeg")).toStrictEqual(Err("Coffee spilled"));
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
});
