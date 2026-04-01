import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";

export type AlbumImagesResult = Result<
    number[],
    { status: number; message: string }
>;

type AlbumImagesResponse = number[];

export default async function AlbumImagesGetter(
    albumName: string,
): Promise<AlbumImagesResult> {
    return get_album_images_with_handler(
        ApiHandler<AlbumImagesResponse>(
            SERVER_ADDRESS + "album/images/",
            false,
        )(Method.GET),
        albumName,
    );
}

async function get_album_images_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<AlbumImagesResponse, RequestError>>,
    albumName: string,
): Promise<AlbumImagesResult> {
    const result = await handler(encodeURIComponent(albumName));

    if (!result.ok) {
        const status = result.error.status_code;
        const bodyText = await result.error.response.text();
        if (status === 400 || status === 404) {
            let message = bodyText;
            try {
                message = (JSON.parse(bodyText) as { error?: string }).error ||
                    bodyText;
            } catch {
                // keep raw response text if backend returned plain text
            }
            return Err({
                status,
                message,
            });
        }
        return Err({
            status,
            message: bodyText,
        });
    }

    return result;
}

export const testing_AlbumImagesGetter = get_album_images_with_handler;
