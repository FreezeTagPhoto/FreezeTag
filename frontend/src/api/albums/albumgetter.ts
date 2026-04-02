import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";

export type AlbumData = {
    id: number;
    name: string;
    owner_id: number;
    visibility_mode: number;
};

export type AlbumResult = Result<
    AlbumData,
    { status: number; message: string }
>;

export default async function AlbumGetter(
    albumID: number,
): Promise<AlbumResult> {
    const url = `${SERVER_ADDRESS}album/${albumID}`;
    const handler = ApiHandler<AlbumData>(url, false)(Method.GET);
    const result = await handler("");

    if (!result.ok) {
        return Err({
            status: result.error.status_code,
            message: "Failed to load album details.",
        });
    }
    return result;
}
