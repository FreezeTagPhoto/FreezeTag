import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Ok } from "@/common/result";

export type AlbumImageAddResponse = {
    message: string;
};

export default async function AlbumImageAdder(
    image_id: number,
    album_name: string,
): Promise<Result<AlbumImageAddResponse, RequestError>> {
    return add_image_to_album_with_handler(
        ApiHandler<AlbumImageAddResponse>(
            SERVER_ADDRESS + "album/add_image_by_name",
        )(Method.POST),
        image_id,
        album_name,
    );
}

async function add_image_to_album_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<AlbumImageAddResponse, RequestError>>,
    image_id: number,
    album_name: string,
): Promise<Result<AlbumImageAddResponse, RequestError>> {
    const result = await handler(
        JSON.stringify({
            image_id,
            album_name,
        }),
    );
    if (!result.ok) return result;
    return Ok(result.value);
}

export const testing_AlbumImageAdder = add_image_to_album_with_handler;
