import { Result, Ok, Err } from "@/common/result";
import { install } from "undici";

install();

export enum Method {
    GET = "GET",
    POST = "POST",
    DELETE = "DELETE",
}

export type RequestError = { status_code: number; response: Response };

export function ApiHandler<T>(address: string, body_request: boolean = true) {
    return (method: Method) => {
        if (method === Method.POST && body_request) {
            return async (data: BodyInit): Promise<Result<T, RequestError>> => {
                let response;
                try {
                    response = await fetch(address, {
                        method: method,
                        body: data,
                    });
                } catch (error) {
                    return Err({
                        status_code: 0,
                        response: new Response(`Catastrophic Error: ${error}`),
                    });
                }

                if (response.ok) {
                    return Ok((await response.json()) as T);
                } else {
                    return Err({ status_code: response.status, response });
                }
            };
        } else {
            return async (data: BodyInit): Promise<Result<T, RequestError>> => {
                let response;
                try {
                    response = await fetch(address + data, {
                        method: method,
                    });
                } catch (error) {
                    return Err({
                        status_code: 0,
                        response: new Response(`Catastrophic Error: ${error}`),
                    });
                }

                if (response.ok) {
                    return Ok((await response.json()) as T);
                } else {
                    return Err({ status_code: response.status, response });
                }
            };
        }
    };
}
