import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";

export type AlbumRenameResult = Result<
    { message: string },
    { status: number; message: string }
>;

export default async function AlbumRenamer(
    albumId: number,
    newName: string,
): Promise<AlbumRenameResult> {
    const url = `${SERVER_ADDRESS}album/${albumId}/name`;
    const handler = ApiHandler<{ message: string }>(url, true)(Method.PATCH);
    const result = await handler(
        JSON.stringify({
            album_id: albumId,
            new_name: newName,
        }),
    );

    if (!result.ok) {
        const status = result.error.status_code;
        const bodyText = await result.error.response.text();
        let message = bodyText;

        try {
            const parsed = JSON.parse(bodyText);
            message = parsed.error || parsed.message || bodyText;
        } catch {
            /* use raw text */
        }

        return Err({ status, message });
    }

    return result;
}
