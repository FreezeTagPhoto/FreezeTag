import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";

export type AlbumDeleteResult = Result<
    { message: string },
    { status: number; message: string }
>;

type AlbumDeleteResponse = { message: string };

export default async function AlbumDeleter(
    albumName: string,
): Promise<AlbumDeleteResult> {
    return delete_album_with_handler(
        ApiHandler<AlbumDeleteResponse>(SERVER_ADDRESS + "album/delete")(
            Method.POST,
        ),
        albumName,
    );
}

async function delete_album_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<AlbumDeleteResponse, RequestError>>,
    albumName: string,
): Promise<AlbumDeleteResult> {
    const result = await handler(JSON.stringify({ album_name: albumName }));

    if (!result.ok) {
        const status = result.error.status_code;
        const bodyText = await result.error.response.text();
        if (status === 400) {
            let message = bodyText;
            try {
                message = (JSON.parse(bodyText) as { error?: string }).error ||
                    bodyText;
            } catch {
                // noop? other files do it ig
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

export const testing_AlbumDeleter = delete_album_with_handler;
