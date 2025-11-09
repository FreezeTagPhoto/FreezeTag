import SERVER_ADDRESS from "@/api/common/serveraddress";
import { ApiHandler, Method } from "@/api/common/apihandler";
import { Result, Err, Ok } from "@/common/result";

// Maps from image path to its ID on success, or to an error message on failure.
export type UploadResult = Result<Map<string, Result<number, string>>, number>;
type UploadResponse = {
  uploaded: { id: number; filename: string }[];
  errors: { error: string; filename: string }[];
};

export async function ImageUploader(event: FormData): Promise<UploadResult> {
  return image_upload_with_handler(
    ApiHandler<UploadResponse>(SERVER_ADDRESS + "upload/")(Method.POST),
    event,
  );
}

async function image_upload_with_handler(
  handler: (data: BodyInit) => Promise<Result<UploadResponse, number>>,
  event: FormData,
): Promise<UploadResult> {
  const request_result = await handler(event);

  if (!request_result.ok) {
    return Err(request_result.error);
  }

  const body = request_result.value;

  const image_map = new Map();
  for (const image of body.uploaded) {
    image_map.set(image.filename, Ok(image.id));
  }
  for (const error of body.errors) {
    image_map.set(error.filename, Err(error.error));
  }

  return Ok(image_map);
}

export const testing_ImageUploader = image_upload_with_handler;
export type testing_UploadResponse = UploadResponse;
