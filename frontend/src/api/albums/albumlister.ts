import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method, RequestError } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";

export type AlbumItem = { 
    id: number;
    name: string;
    owner_id: number;
}

export type AlbumListResult = Result<
    AlbumItem[],
    { status: number; message: string }
>;

type AlbumListResponse = AlbumItem[];


export default async function AlbumLister(): Promise<AlbumListResult> {
    return list_albums_with_handler(
        ApiHandler<AlbumListResponse>(SERVER_ADDRESS + "/album")(Method.GET),
    );
}

async function list_albums_with_handler(
    handler: (data: BodyInit) => Promise<Result<AlbumListResponse, RequestError>>,
): Promise<AlbumListResult> {
    const result = await handler("");

    if (!result.ok) {
        const status = result.error.status_code;
        const bodyText = await result.error.response.text();
        return Err({
            status,
            message: bodyText || "Failed to load albums.",
        });
    }

    return result;
}
