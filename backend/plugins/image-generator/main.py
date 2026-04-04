import freezetag
from freezetag.hooks import image_batch, form_data, SendFormAction, AddImageAction
from diffusers import AutoPipelineForText2Image
import torch


@image_batch
def get_prompt_form(_ids):
    return SendFormAction(
        """<form><div style="display:flex; flex-direction:column; gap: 1rem; padding: 1rem; font-size: 1.5rem;">
        <p>Using Stable Diffusion v1.5</p>
        <label for="prompt">Prompt:</label><input style="font-size: 1.5rem;" type="text" name="prompt" id="prompt" required></input>
        <input style="padding: 1rem; font-size: 1.5rem;" type="submit" value="Make Image!"/>
        </div></form>"""
    )


@form_data
def process_prompt_form(data):
    prompt = data["prompt"]

    pipeline = AutoPipelineForText2Image.from_pretrained(
        "stable-diffusion-v1-5/stable-diffusion-v1-5",
        torch_dtype=torch.float16,
        variant="fp16",
        device_map="balanced",
    )

    image = image = pipeline(prompt).images[0]

    return AddImageAction("".join(prompt.split()), "png", image)


if __name__ == "__main__":
    freezetag.run()
