import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Ok } from "@/common/result";

export type AlbumCreateResponse = {
    album_id: number;
};

export default async function AlbumCreator(
    name: string,
    visibility_mode: number,
): Promise<Result<AlbumCreateResponse, RequestError>> {
    return create_album_with_handler(
        ApiHandler<AlbumCreateResponse>(SERVER_ADDRESS + "album/")(Method.POST),
        name,
        visibility_mode,
    );
}

async function create_album_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<AlbumCreateResponse, RequestError>>,
    name: string,
    visibility_mode: number,
): Promise<Result<AlbumCreateResponse, RequestError>> {
    const result = await handler(
        JSON.stringify({
            name,
            visibility_mode,
        }),
    );
    if (!result.ok) return result;
    return Ok(result.value);
}

export const testing_AlbumCreator = create_album_with_handler;
