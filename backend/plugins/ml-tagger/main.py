import freezetag
import spacy
from freezetag.hooks import process_func, init_func, TagAction, SkipAction
from freezetag.message import log
from PIL import Image
from transformers import BlipProcessor, BlipForConditionalGeneration
from spacy.cli import download

processor = None
model = None
nlp = None

@init_func
def init():
    global processor, model, nlp
    log("loading models...")
    nlp = spacy.load("en_core_web_sm")
    processor = BlipProcessor.from_pretrained("Salesforce/blip-image-captioning-large")
    model = BlipForConditionalGeneration.from_pretrained("Salesforce/blip-image-captioning-large").to("cuda")
    log("finished loading models")

@process_func
def tag_image(img: Image.Image, id: int) -> TagAction:
    global processor, model, nlp
    img.thumbnail((1024, 1024))
    input_image = img.convert("RGB")
    log("generating sentence...")
    inputs = processor(input_image, return_tensors="pt").to("cuda")
    out = model.generate(**inputs)
    sentence = processor.decode(out[0], skip_special_tokens=True)
    log(f"sentence = {sentence}")
    doc = nlp(sentence)
    nouns = [token.text for token in doc if token.pos_ == "NOUN" or token.pos_ == "PROPN"]
    log(f"nouns = {nouns}")
    return TagAction(id, nouns)

if __name__ == "__main__":
    freezetag.run()