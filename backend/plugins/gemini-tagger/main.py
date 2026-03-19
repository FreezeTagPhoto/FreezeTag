import os
import time
from PIL import Image

import freezetag
from freezetag.hooks import single_image, init_func, AddTagsAction
from freezetag.message import log
from google import genai
from dotenv import load_dotenv, find_dotenv
from google.genai import types

client = None
image_count = 0

prompt = """
Provide a comma-separated list of descriptive tags for this image.
create tags that a human would reasonably use to describe the content of the image to another human. 
Output only the tags"""

@init_func
def init():
    global client
    global model, device, transform
    log("Initializing Gemini Tagger plugin...")
    try:
        load_dotenv(find_dotenv())
        api_token = os.getenv("GEMINI-API-KEY")
        if not api_token:
            log("GEMINI-API-KEY not found in environment variables.")
            return
        client = genai.Client(api_key=api_token)
    except Exception as e:
        log(f"Error initializing Gemini Tagger plugin: {e}")
        return
        

@single_image
def tag_image(img: Image.Image, id: int) -> AddTagsAction:
    global client, image_count
    image_count += 1
    if image_count % 15 == 0:
        log("Sleeping for a bit to avoid hitting Gemini API rate limits...")
        time.sleep(60)
    assert client is not None, "Gemini client not initialized"
    try:
        
        img = _compress_image(img)

        response = client.models.generate_content(
            model="gemini-3.1-flash-lite-preview", 
            contents=[prompt, img],
            config=types.GenerateContentConfig(
                temperature=0,
                candidate_count=1
            ),
        )
        
        raw_text = response.text.strip()
        tags_list = [tag.strip() for tag in raw_text.split(",") if tag.strip()]
        
        log(f"Generated {len(tags_list)} tags for image {id}")
        return AddTagsAction(id, tags_list)
    except Exception as e:
        log(f"Error processing image {id}: {e}")
        return AddTagsAction(id, [])

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