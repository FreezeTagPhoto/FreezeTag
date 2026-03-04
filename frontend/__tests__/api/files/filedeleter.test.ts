/**
 * @jest-environment node
 */

import { RequestError } from "@/api/common/apihandler";
import FileDeleter, {
    testing_FileDeleter,
    type testing_FileDeleteResponse,
} from "@/api/files/filedeleter";
import { Result, Ok, Err } from "@/common/result";

type HandlerReturnType = Promise<
    Result<testing_FileDeleteResponse, RequestError>
>;

describe("FileDeleter", () => {
    it("should pass id correctly", async () => {
        const handler = async (data: BodyInit): HandlerReturnType => {
            expect(typeof data === "string").toBeTruthy();
            expect(data).toBe("7");
            return Ok({ id: 7, file: "some/path.jpg" });
        };

        const result = await testing_FileDeleter(handler, 7);
        expect(result).toStrictEqual(Ok({ id: 7, file: "some/path.jpg" }));
    });

    it("should percolate error on failure", async () => {
        const response = new Response("not found", { status: 404 });

        const handler = async (_: BodyInit): HandlerReturnType => {
            return Err({ status_code: 404, response });
        };

        const result = await testing_FileDeleter(handler, 123);
        expect(result).toStrictEqual(Err({ status_code: 404, response }));
    });

    it("should pass integration tests", async () => {
        global.fetch = jest.fn(() => {
            return Promise.resolve({
                status: 200,
                ok: true,
                json: async () => {
                    return { id: 9, file: "deleted-file.jpg" };
                },
            });
        }) as jest.Mock;

        const res = await FileDeleter(9)();
        expect(res).toStrictEqual(Ok({ id: 9, file: "deleted-file.jpg" }));

        expect((global.fetch as jest.Mock).mock.calls[0][0]).toContain(
            "/file/delete/9",
        );
        expect((global.fetch as jest.Mock).mock.calls[0][1]).toMatchObject({
            method: "DELETE",
        });
    });
});
