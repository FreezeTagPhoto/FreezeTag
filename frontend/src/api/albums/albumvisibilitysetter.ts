import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";

export type AlbumVisibilitySetResult = Result<
    { message: string },
    { status: number; message: string }
>;

type AlbumVisibilitySetResponse = { message: string };

export default async function AlbumVisibilitySetter(
    albumId: number,
    visibilityMode: number,
): Promise<AlbumVisibilitySetResult> {
    return set_album_visibility_with_handler(
        ApiHandler<AlbumVisibilitySetResponse>(
            `${SERVER_ADDRESS}album/${albumId}/visibility`,
        )(Method.PATCH),
        albumId,
        visibilityMode,
    );
}

async function set_album_visibility_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<AlbumVisibilitySetResponse, RequestError>>,
    albumId: number,
    visibilityMode: number,
): Promise<AlbumVisibilitySetResult> {
    const result = await handler(
        JSON.stringify({
            album_id: albumId,
            visibility_mode: visibilityMode,
        }),
    );

    if (!result.ok) {
        const status = result.error.status_code;
        const bodyText = await result.error.response.text();
        return Err({
            status,
            message: bodyText || "Failed to update album visibility.",
        });
    }

    return result;
}
