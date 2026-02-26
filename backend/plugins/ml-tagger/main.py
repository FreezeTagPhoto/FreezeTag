import freezetag
import spacy
from freezetag.hooks import single_image, init_func, AddTagsAction
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
    model = BlipForConditionalGeneration.from_pretrained("Salesforce/blip-image-captioning-large")
    log("finished loading models")

@single_image
def tag_image(img: Image.Image, id: int) -> AddTagsAction:
    global processor, model, nlp
    img.thumbnail((1024, 1024))
    input_image = img.convert("RGB")
    log("generating sentence...")
    inputs = processor(input_image, return_tensors="pt")
    out = model.generate(**inputs)
    sentence = processor.decode(out[0], skip_special_tokens=True)
    log(f"sentence = {sentence}")
    doc = nlp(sentence)
    nouns = [token.text for token in doc if token.pos_ == "NOUN" or token.pos_ == "PROPN"]
    # filter out unwanted words (arafe is a hallucination from BLIP-2)
    nouns = [noun for noun in nouns if noun not in ["arafe", "arafed", "araffe", "s"]]
    log(f"nouns = {nouns}")
    return AddTagsAction(id, nouns)

if __name__ == "__main__":
    freezetag.run()