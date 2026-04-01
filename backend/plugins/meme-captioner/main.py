import freezetag
from freezetag.hooks import single_image, form_data, SendFormAction, AddImageAction
from freezetag.message import get_image

from PIL import ImageDraw, ImageFont


@single_image
def make_caption_form(_img, id):
    return SendFormAction(
        f"""<form><div style="display:flex; flex-direction:column; gap: 1rem; padding: 1rem; font-size: 1.5rem;">
        <input type="hidden" value="{id}" name="id"></input>
        <label for="top_text">Top Text:</label><input style="font-size: 1.5rem;" type="text" name="top_text" id="top_text" required></input>
        <label for="bottom_text">Bottom Text:</label><input style="font-size: 1.5rem;" type="text" name="bottom_text" id="bottom_text" required></input>
        <input style="padding: 1rem; font-size: 1.5rem;" type="submit" value="Caption Meme!"/>
        </div></form>"""
    )


@form_data
def process_caption_form(data):
    id = data["id"]
    top = data["top_text"].upper()
    bottom = data["bottom_text"].upper()

    image = get_image(int(id))
    width, height = image.size
    font_size = height * 0.15

    draw = ImageDraw.Draw(image)
    font = ImageFont.truetype("impact.ttf", font_size)

    while True:
        len_top = font.getlength(top)
        len_bottom = font.getlength(bottom)
        if (len_top > width) or (len_bottom > width):
            font_size *= 0.9
            font = font.font_variant(size=font_size)
        else:
            break

    stroke_width = font_size * 0.05
    top_padding = (width - len_top) / 2
    bottom_padding = (width - len_bottom) / 2
    bottom_bbox = font.getbbox(bottom, stroke_width=stroke_width)
    bottom_gap = bottom_bbox[3] - bottom_bbox[1]

    draw.text(
        (top_padding, 0),
        top,
        (255, 255, 255),
        font=font,
        stroke_width=stroke_width,
        stroke_fill=(0, 0, 0),
    )
    draw.text(
        (bottom_padding, height - bottom_gap * 1.3),
        bottom,
        (255, 255, 255),
        font=font,
        stroke_width=stroke_width,
        stroke_fill=(0, 0, 0),
    )

    return AddImageAction(f"{id}_meme", "png", image)


if __name__ == "__main__":
    freezetag.run()
