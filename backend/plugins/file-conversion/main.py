import freezetag
from freezetag.hooks import single_image, form_data, SendFormAction, AddRawImageWithTagsAction, Error
from freezetag.message import get_image_file, read_config, get_metadata, get_image_tags

from datetime import datetime
from pathlib import Path
from jinja2 import Environment, FileSystemLoader, select_autoescape
from wand.image import Image
import exif
from exif import DATETIME_STR_FORMAT

@single_image
def make_file_form(_img, id):
    config = read_config("config.toml")
    formats = [item.strip() for item in config["format_options"].split(",")]
    env = Environment(
        loader=FileSystemLoader("templates"),
        autoescape=select_autoescape()
    )
    template = env.get_template("format.html")
    return SendFormAction(
        template.render(id=id, formats=formats)
    )


@single_image
def change_metadata_form(_img, id):
    env = Environment(
        loader=FileSystemLoader("templates"),
        autoescape=select_autoescape()
    )
    meta = get_metadata(id)
    date_taken = datetime.fromtimestamp(meta["dateTaken"]).strftime("%Y-%m-%dT%H:%M") if meta["dateTaken"] is not None else ""
    latitude = meta["latitude"] if meta["latitude"] is not None else ""
    longitude = meta["longitude"] if meta["longitude"] is not None else ""
    make = meta["cameraMake"] if meta["cameraMake"] is not None else ""
    model = meta["cameraModel"] if meta["cameraModel"] is not None else ""
    template = env.get_template("metadata.html")
    return SendFormAction(
        template.render(id=id, date_taken=date_taken, latitude=latitude, longitude=longitude, camera_make=make, camera_model=model)
    )

def decimal_to_exif_dms(decimal_degree):
    """Converts decimal degrees to (degrees, minutes, seconds) for EXIF."""
    degree = int(abs(decimal_degree))
    minute = int((abs(decimal_degree) - degree) * 60)
    second = (abs(decimal_degree) * 60 - degree * 60 - minute) * 60
    
    return (degree, minute, second)

@form_data
def process_file_form(data):
    id = int(data["id"])
    action = data["action"]
    match action:
        case "format":
            format = data["format_select"]
            meta = get_metadata(id)
            filename = Path(meta["fileName"]).stem
            tags = get_image_tags(id)
            tags.append(f"converted:{format}")

            with Image(blob=get_image_file(id)) as img:
                img_bin = img.make_blob(format)
                return AddRawImageWithTagsAction(f"conv-{filename}", format, img_bin, tags)
        case "metadata":
            meta = get_metadata(id)
            filename = Path(meta["fileName"]).stem
            ext = Path(meta["fileName"]).suffix
            tags = get_image_tags(id)
            tags.append(f"converted:metadata")
            img = exif.Image(get_image_file(id))
            if not img.has_exif:
                return Error(f"image doesn't have EXIF fields")
            date_taken = datetime.fromisoformat(data["date_taken"]) if data["date_taken"] != "" else None
            latitude = float(data["latitude"]) if data["latitude"] != "" else None
            longitude = float(data["longitude"]) if data["longitude"] != "" else None
            make = data["camera_make"]
            model = data["camera_model"]
            if date_taken is not None:
                img.datetime_original = date_taken.strftime(DATETIME_STR_FORMAT)
            if latitude is not None:
                img.gps_latitude = decimal_to_exif_dms(latitude)
                img.gps_latitude_ref = "S" if latitude < 0 else "N"
            if longitude is not None:
                img.gps_longitude = decimal_to_exif_dms(longitude)
                img.gps_longitude_ref = "W" if longitude < 0 else "E"
            img.make = make
            img.model = model
            return AddRawImageWithTagsAction(f"conv-{filename}", ext, img.get_file(), tags)
        case _:
            return Error(f"unsupported action: {action}")


if __name__ == "__main__":
    freezetag.run()