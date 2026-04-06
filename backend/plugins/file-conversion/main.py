import freezetag
from freezetag.hooks import single_image, form_data, SendFormAction, AddRawImageAction
from freezetag.message import get_image_file, read_config, get_metadata

from pathlib import Path
from jinja2 import Environment, FileSystemLoader, select_autoescape
from wand.image import Image

@single_image
def make_file_form(_img, id):
    config = read_config("config.toml")
    formats = [item.strip() for item in config["format_options"].split(",")]
    env = Environment(
        loader=FileSystemLoader("templates"),
        autoescape=select_autoescape()
    )
    template = env.get_template("form.html")
    return SendFormAction(
        template.render(id=id, formats=formats)
    )


@form_data
def process_file_form(data):
    id = int(data["id"])
    format = data["format_select"]
    meta = get_metadata(id)
    filename = Path(meta["fileName"]).stem

    with Image(blob=get_image_file(id)) as img:
        img_bin = img.make_blob(format)
        return AddRawImageAction(f"conv-{filename}", format, img_bin)


if __name__ == "__main__":
    freezetag.run()