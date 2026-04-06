import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method } from "@/api/common/apihandler";
import { Result, Err } from "@/common/result";

export type AlbumImagesResult = Result<
    number[],
    { status: number; message: string }
>;

export default async function AlbumImagesGetter(
    albumID: number,
): Promise<AlbumImagesResult> {
    const url = `${SERVER_ADDRESS}album/${albumID}/images`;
    const handler = ApiHandler<number[]>(url, false)(Method.GET);

    const result = await handler("");

    if (!result.ok) {
        const status = result.error.status_code;
        const bodyText = await result.error.response.text();
        let message = bodyText;

        if (status === 400 || status === 404) {
            try {
                const parsed = JSON.parse(bodyText);
                if (parsed.error) {
                    message = parsed.error;
                }
            } catch {
                // Keep raw response text if backend returned plain text
            }
        }

        return Err({
            status,
            message: message || "Failed to load album images.",
        });
    }

    return result;
}
