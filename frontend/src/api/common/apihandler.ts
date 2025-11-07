import { Result, Ok, Err } from "@/common/result";
import { fetch, Response, BodyInit } from "undici";

export enum Method {
  GET = "GET",
  POST = "POST",
  DELETE = "DELETE",
}
export type RequestBody = BodyInit;
export type RequestResult = Result<RequestBody, number>;

export function ApiHandler(address: string) {
  if (!address.endsWith("/")) address += "/";
  return (method: Method) => {
    return async (data: string): Promise<RequestResult> => {
      let result: Promise<Response>;
      if (method === Method.POST) {
        result = fetch(address, {
          method: method,
          body: data,
        });
      } else {
        result = fetch(address + data, {
          method: method,
        });
      }
      const response = await result;

      if (response.ok) {
        return Ok(response.body);
      } else {
        return Err(response.status);
      }
    };
  };
}
