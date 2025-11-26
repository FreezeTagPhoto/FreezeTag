/**
 * @jest-environment node
 */

import { RequestError } from "@/api/common/apihandler";
import { testing_TagGetResponse, testing_TagGetter } from "@/api/tags/taggetter";

import { Result, Ok, Err } from "@/common/result";

type HandlerReturnType = Promise<Result<testing_TagGetResponse, RequestError>>;

describe("Tag Getter", () => {
  it("should not include a leading slash without an image id", () => {
    const handler = async (query: BodyInit): HandlerReturnType => {
      expect(query).toBe("");
      return Ok([]);
    };

    testing_TagGetter(handler);
  });
});
