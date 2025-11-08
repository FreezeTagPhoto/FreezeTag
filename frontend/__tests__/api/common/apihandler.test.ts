/**
 * @jest-environment node
 */

import "@testing-library/jest-dom";
import { ApiHandler, Method } from "@/api/common/apihandler";
import { Err } from "@/common/result";

describe("API Handler", () => {
  it("can receive a response", async () => {
    const handler = ApiHandler(
      "http://api.github.com/users/SathyaTadinada/repos",
    )(Method.GET);
    const response = await handler("");

    expect(response.ok).toBeTruthy();
    if (response.ok) {
      expect(response.value).not.toBeNull();
    }
  });

  it("properly handles a 404", async () => {
    const handler = ApiHandler("http://google.com/free-ice-cream")(Method.GET);
    const response = await handler("");

    expect(response).toStrictEqual(Err(404));
  });

  it("properly handles a 405 response", async () => {
    const handler = ApiHandler("http://www.google.com")(Method.POST);
    const response = await handler("{status: 'good'}");

    expect(response).toStrictEqual(Err(405));
  });
});
