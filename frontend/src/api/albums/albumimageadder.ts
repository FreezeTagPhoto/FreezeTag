import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Ok } from "@/common/result";

export type AlbumImageAddResponse = {
    message: string;
};

export default async function AlbumImageAdder(
    image_id: number,
    album_id: number,
): Promise<Result<AlbumImageAddResponse, RequestError>> {
    return add_image_to_album_with_handler(
        ApiHandler<AlbumImageAddResponse>(
            SERVER_ADDRESS + "album/" + album_id + "/images",
        )(Method.POST),
        image_id,
    );
}

async function add_image_to_album_with_handler(
    handler: (
        data: BodyInit,
    ) => Promise<Result<AlbumImageAddResponse, RequestError>>,
    image_id: number,
): Promise<Result<AlbumImageAddResponse, RequestError>> {
    const result = await handler(
        JSON.stringify({
            image_id,
        }),
    );
    if (!result.ok) return result;
    return Ok(result.value);
}

export const testing_AlbumImageAdder = add_image_to_album_with_handler;
