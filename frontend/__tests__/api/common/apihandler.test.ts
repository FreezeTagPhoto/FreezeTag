/**
 * @jest-environment node
 */

import "@testing-library/jest-dom";
import { ApiHandler, Method } from "@/api/common/apihandler";

describe("API Handler", () => {
  it("can receive a response", async () => {
    const handler = new ApiHandler("http://www.google.com", Method.GET);
    const response = await handler.send_request("search?q=university of utah");

    if (response.ok) {
      expect(response.value).not.toBeNull();
    } else {
      fail();
    }
  });

  it("properly handles a 405 response", async () => {
    const handler = new ApiHandler("http://www.google.com", Method.POST);
    const response = await handler.send_request("{status: 'good'}");

    if (response.ok) {
      fail();
    } else {
      expect(response.error).toBe(405);
    }
  });
});
