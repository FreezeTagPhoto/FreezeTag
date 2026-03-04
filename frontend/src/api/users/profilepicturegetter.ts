import SERVER_ADDRESS from "@/api/common/serveraddress";
import { Result, Err, Ok } from "@/common/result";

export type ProfilePictureGetResult = Result<
    Blob,
    { status: number; message: string }
>;

/**
 * GET /users/profile-picture/{id}
 * Returns raw image bytes (e.g. image/webp) as a Blob.
 */
export default async function ProfilePictureGetter(
    user_id: number,
): Promise<ProfilePictureGetResult> {
    const url = SERVER_ADDRESS + `users/profile-picture/${user_id}`;

    let resp: Response;
    try {
        resp = await fetch(url, {
            method: "GET",
            credentials: "include",
            headers: {
                Accept: "image/*",
            },
        });
    } catch (e) {
        return Err({
            status: 0,
            message:
                e instanceof Error ? e.message : "Network error fetching image",
        });
    }

    if (!resp.ok) {
        const status = resp.status;

        // If your backend returns JSON { error: string } on failures (like other endpoints),
        // try to parse it; otherwise fallback to text.
        const contentType = resp.headers.get("content-type") ?? "";
        if (contentType.includes("application/json")) {
            try {
                const data = (await resp.json()) as { error?: string };
                return Err({
                    status,
                    message: data.error ?? "Failed to fetch profile picture",
                });
            } catch {
                return Err({
                    status,
                    message: "Failed to fetch profile picture",
                });
            }
        }

        const text = await resp.text().catch(() => "");
        return Err({
            status,
            message: text || "Failed to fetch profile picture",
        });
    }

    const blob = await resp.blob();
    return Ok(blob);
}
