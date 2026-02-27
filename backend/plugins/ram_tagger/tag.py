import os
import torch
import numpy as np
from PIL import Image
from huggingface_hub import hf_hub_download

import freezetag
from freezetag.hooks import single_image, init_func, AddTagsAction
from freezetag.message import log
from ram.models.ram_plus import ram_plus
from ram import inference_ram as inference
from ram import get_transform

model = None
device = None
transform = None

@init_func
def init():
    global model, device, transform
    log("RAM++: Checking for pretrained weights in ./pretrained...")
    
    try:
        # this uses a pretty large model, so it needs to be downloaded from huggingface
        model_path = hf_hub_download(
            repo_id="xinyu1205/recognize-anything-plus-model",
            filename="ram_plus_swin_large_14m.pth",
            local_dir="./pretrained",
            local_dir_use_symlinks=False
        )
        
        device = torch.device('cpu')
        log(f"RAM++: Loading model onto {device}...")
        
        # Initialize model (vit='swin_l' matches the 14m checkpoint)
        model = ram_plus(pretrained=model_path, image_size=384, vit='swin_l')
        model.eval()
        model = model.to(device)
        
        transform = get_transform(image_size=384)
        log("RAM++: Finished loading model")
        
    except Exception as e:
        log(f"RAM++: Failed to initialize: {e}")

@single_image
def tag_image(img: Image.Image, id: int) -> AddTagsAction:
    global model, device, transform
    assert transform is not None, "Model not initialized properly"
    try:
        input_image = img.convert("RGB")
        image_tensor = transform(input_image).unsqueeze(0).to(device)
        tags_list = inference(image_tensor, model)
    
        

        log(f"Found tags for image {id}: {tags_list}")
        return AddTagsAction(id, tags_list)

    except Exception as e:
        log(f"Error processing image {id}: {e}")
        return AddTagsAction(id, [])

if __name__ == "__main__":
    freezetag.run()