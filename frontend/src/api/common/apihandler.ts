import { Result, Ok, Err } from "@/common/result";
import { fetch, Response, BodyInit } from "undici";

export enum Method {
  GET = "GET",
  POST = "POST",
}
export type RequestBody = BodyInit;
export type RequestResult = Result<RequestBody, number>;

export class ApiHandler {
  private address: string;
  private method: Method;

  constructor(address: string, method: Method) {
    this.address = address;
    if (address.charAt(address.length - 1) != "/") {
      this.address = this.address + "/";
    }
    this.method = method;
  }

  public async send_request(data: string): Promise<RequestResult> {
    let result: Promise<Response>;
    if (this.method === Method.GET) {
      result = fetch(this.address + data, {
        method: this.method,
      });
    } else {
      result = fetch(this.address, {
        method: this.method,
        body: data,
      });
    }

    const response = await result;

    if (response.ok) {
      return Ok(response.body);
    } else {
      return Err(response.status);
    }
  }
}
