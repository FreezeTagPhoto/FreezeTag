/**
 * @jest-environment node
 */

import { RequestError } from "@/api/common/apihandler";
import {
    testing_MetadataGetResponse,
    testing_MetadataGetter,
} from "@/api/metadata/metadatagetter";
import { Result, Ok, Err } from "@/common/result";

type HandlerReturnType = Promise<
    Result<testing_MetadataGetResponse, RequestError>
>;

describe("Metadata Getter", () => {
    it("should include a leading slash and the image id", async () => {
        const handler = async (query: BodyInit): HandlerReturnType => {
            expect(query).toBe("/67");
            return Ok({
                fileName: null,
                dateTaken: null,
                dateUploaded: null,
                cameraMake: null,
                cameraModel: null,
                latitude: null,
                longitude: null,
            });
        };

        await testing_MetadataGetter(handler, 67);
    });

    it("should get message on 400", async () => {
        const handler = async (_: BodyInit): HandlerReturnType => {
            const response = new Response(
                '{"error": "Invalid image ID parameter"}',
            );
            return Err({
                status_code: 400,
                response,
            });
        };

        const result = await testing_MetadataGetter(handler, 1);
        expect(result).toStrictEqual(
            Err({ status: 400, message: "Invalid image ID parameter" }),
        );
    });

    it("should percolate status code on failure", async () => {
        const handler = async (_: BodyInit): HandlerReturnType => {
            return Err({
                status_code: 500,
                response: new Response("server died"),
            });
        };

        const result = await testing_MetadataGetter(handler, 1);
        expect(result).toStrictEqual(
            Err({ status: 500, message: "server died" }),
        );
    });

    it("should receive metadata", async () => {
        const handler = async (_: BodyInit): HandlerReturnType => {
            return Ok({
                fileName: "IMG_1234.JPG",
                dateTaken: 1700000000,
                dateUploaded: 1700000100,
                cameraMake: "Canon",
                cameraModel: "R6",
                latitude: 40,
                longitude: -111,
            });
        };

        const result = await testing_MetadataGetter(handler, 1);
        expect(result).toStrictEqual(
            Ok({
                fileName: "IMG_1234.JPG",
                dateTaken: 1700000000,
                dateUploaded: 1700000100,
                cameraMake: "Canon",
                cameraModel: "R6",
                latitude: 40,
                longitude: -111,
            }),
        );
    });

    it("should fall back to text if error JSON can't be parsed", async () => {
        const handler = async (_: BodyInit): HandlerReturnType => {
            return Err({
                status_code: 400,
                response: new Response("{not json"),
            });
        };

        const result = await testing_MetadataGetter(handler, 1);
        expect(result).toStrictEqual(
            Err({
                status: 400,
                message: await new Response("{not json").text(),
            }),
        );
    });
});
