import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";

export type AlbumPermissionSetResult = Result<
    { message: string },
    { status: number; message: string }
>;

type AlbumPermissionSetResponse = { message: string };

export default async function AlbumPermissionSetter(
    album_id: number,
    target_user_id: number,
    permission: number,
): Promise<AlbumPermissionSetResult> {
    return set_album_permission_with_handler(
        ApiHandler<AlbumPermissionSetResponse>(
            SERVER_ADDRESS + `album/${album_id}/permissions`,
        )(Method.PUT),
        album_id,
        target_user_id,
        permission,
    );
}

async function set_album_permission_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<AlbumPermissionSetResponse, RequestError>>,
    album_id: number,
    target_user_id: number,
    permission: number,
): Promise<AlbumPermissionSetResult> {
    const result = await handler(
        JSON.stringify({
            album_id,
            target_user_id,
            permission,
        }),
    );

    if (!result.ok) {
        const status = result.error.status_code;
        const bodyText = await result.error.response.text();
        return Err({
            status,
            message: bodyText || "Failed to update sharing.",
        });
    }

    return result;
}
