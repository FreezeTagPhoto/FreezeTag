import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";

export type AlbumImageNamesResult = Result<
    string[],
    { status: number; message: string }
>;

type AlbumImageNamesResponse = string[];

export default async function AlbumImageNamesGetter(
    imageId: number,
): Promise<AlbumImageNamesResult> {
    return get_image_album_names_with_handler(
        ApiHandler<AlbumImageNamesResponse>(
            SERVER_ADDRESS + "album/image_names",
        )(Method.GET),
        imageId,
    );
}

async function get_image_album_names_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<AlbumImageNamesResponse, RequestError>>,
    imageId: number,
): Promise<AlbumImageNamesResult> {
    const result = await handler(`/${imageId}`);

    if (!result.ok) {
        const status = result.error.status_code;
        if (status === 400) {
            return Err({
                status,
                message: ((await result.error.response.json()) as {
                    error: string;
                }).error,
            });
        }
        return Err({
            status,
            message: await result.error.response.text(),
        });
    }

    return result;
}

export const testing_AlbumImageNamesGetter = get_image_album_names_with_handler;
