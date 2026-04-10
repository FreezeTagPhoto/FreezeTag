import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Ok, Err } from "@/common/result";

export type AlbumCreateResponse = {
    album_id: number;
};

export type AlbumCreateResult = Result<
    AlbumCreateResponse,
    { status: number; message: string }
>;

export default async function AlbumCreator(
    name: string,
    visibility_mode: number,
): Promise<AlbumCreateResult> {
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
): Promise<AlbumCreateResult> {
    const result = await handler(
        JSON.stringify({
            name,
            visibility_mode,
        }),
    );
    if (!result.ok) {
        const status = result.error.status_code;
        const bodyText = await result.error.response.text();
        let message = bodyText;
        try {
            const parsed = JSON.parse(bodyText);
            if (parsed.error) {
                message = parsed.error;
            }
        } catch {
            // Keep raw response text if backend returned plain text
        }
        return Err({
            status,
            message: message || "Failed to load album images.",
        });
    }
    return Ok(result.value);
}



export const testing_AlbumCreator = create_album_with_handler;
