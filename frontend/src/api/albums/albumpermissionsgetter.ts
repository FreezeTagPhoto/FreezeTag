import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";

export type AlbumUserPermission = {
    user_id: number;
    permission: number;
};

export type AlbumPermissionsGetResult = Result<
    AlbumUserPermission[],
    { status: number; message: string }
>;

type AlbumPermissionsGetResponse = AlbumUserPermission[];

export default async function AlbumPermissionsGetter(
    albumId: number,
): Promise<AlbumPermissionsGetResult> {
    return get_album_permissions_with_handler(
        ApiHandler<AlbumPermissionsGetResponse>(
            `${SERVER_ADDRESS}album/${albumId}/permissions`,
        )(Method.GET),
    );
}

async function get_album_permissions_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<AlbumPermissionsGetResponse, RequestError>>,
): Promise<AlbumPermissionsGetResult> {
    const result = await handler("");

    if (!result.ok) {
        const status = result.error.status_code;
        const bodyText = await result.error.response.text();
        return Err({
            status,
            message: bodyText || "Failed to load album permissions.",
        });
    }

    return result;
}
