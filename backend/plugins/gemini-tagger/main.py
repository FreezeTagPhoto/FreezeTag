import os
import time
from PIL import Image

import freezetag
from freezetag.hooks import single_image, init_func, AddTagsAction, Error
from freezetag.message import log, read_config
from google import genai
from google.genai import types
from pydantic import BaseModel, Field
from enum import Enum

class Models(Enum):
    GEMINI_3_1_FLASH_LITE_PREVIEW = "gemini-3.1-flash-lite-preview"
    GEMMA_3_4B_IT = "gemma-3-4b-it"


client = None
image_count = 0

prompt = """
Provide a comma-separated list of descriptive tags for this image.
create tags that a human would reasonably use to describe the content of the image to another human
ensure the tags are specific and descriptive, avoid generic tags like "photo" or "image"
aim for a diverse set of tags that cover different aspects of the image
if text is present in the image, include key words from the text as tags, but do not include the entire text as a tag
avoid creating tags that are too similar to each other, such as "cat" and "cats", or "red shirt" and "red shirts"
avoid ambiguous tags
Output only the tags
"""




class Tag(BaseModel):
    name: str = Field(description="The name of the tag")

class TagResponse(BaseModel):
    tags: list[Tag] = Field(description="A list of tags generated for the image")

@init_func
def init():
    global client
    global model, device, transform
    log("Initializing Gemini Tagger plugin...")
    try:
        api_token = read_config("config.toml")["gemini_key"]
        if not api_token or api_token == "":
            return Error("'gemini_key' not set in config.toml")
        client = genai.Client(api_key=api_token)
    except Exception as e:
        log(f"Error initializing Gemini client: {e}")
        return Error(f"Error initializing Gemini client: {e}")
    log("Gemini Tagger plugin initialized successfully.")
        

@single_image
def tag_image(img: Image.Image, id: int):
    global client, image_count
    image_count += 1
    if image_count % 15 == 0:
        log("Sleeping for a bit to avoid hitting Gemini API rate limits...")
        time.sleep(60)
    assert client is not None, "Gemini client not initialized"
    try:
        img = _compress_image(img)
        response = client.models.generate_content(
        model=Models.GEMINI_3_1_FLASH_LITE_PREVIEW.value,
        contents=[prompt, img],
            config=types.GenerateContentConfig(
                temperature=0,
                response_mime_type="application/json",
                response_schema=TagResponse,
            ),
        )
        log(f"Gemini response for image {id}: {response.text}")
        tags_data = response.parsed
        tags_list = [tag.name for tag in tags_data.tags]
        log(f"Generated {len(tags_list)} tags for image {id}")
        return AddTagsAction(id, tags_list)
    except Exception as e:
        log(f"Error processing image {id}: {e}")
        return Error(f"Error processing image {id}: {e}")

def _compress_image(img: Image.Image) -> types.Part:
    img.thumbnail((512, 512))
    img = img.convert("RGB")
    from io import BytesIO
    buffer = BytesIO()
    img.save(buffer, format="JPEG", optimize=True, quality=20)
    return types.Part.from_bytes(
        data=buffer.getvalue(),
        mime_type="image/jpeg"
    )

if __name__ == "__main__":
    freezetag.run()