import freezetag
from freezetag.hooks import single_image, form_data, SendFormAction, MultipartAction, AddRawImageWithTagsAction, NoAction, DeleteImageAction, Error
from freezetag.message import get_image_file, read_config, get_metadata, get_image_tags, log

from PIL import Image
import piexif, io

from datetime import datetime
from pathlib import Path
from jinja2 import Environment, FileSystemLoader, select_autoescape
from wand.image import Image as WandImage

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
    ext = Path(meta["fileName"]).suffix
    if ext != ".jpg" and ext != ".jpeg":
        return SendFormAction("<form><div style=\"display:flex; flex-direction:column; gap: 1rem; padding: 1rem; font-size: 1.5rem;\"><p>This feature doesn't work with images that aren't JPEG formatted. Use the conversion feature to convert them to JPEG before using this feature!</p></div></form>")
    date_taken = datetime.fromtimestamp(meta["dateTaken"]).strftime("%Y-%m-%dT%H:%M") if meta["dateTaken"] is not None else ""
    latitude = meta["latitude"] if meta["latitude"] is not None else ""
    longitude = meta["longitude"] if meta["longitude"] is not None else ""
    make = meta["cameraMake"] if meta["cameraMake"] is not None else ""
    model = meta["cameraModel"] if meta["cameraModel"] is not None else ""
    template = env.get_template("metadata.html")
    return SendFormAction(
        template.render(id=id, date_taken=date_taken, latitude=latitude, longitude=longitude, camera_make=make, camera_model=model)
    )

def _decimal_to_dms(value: float):
    abs_val = abs(value)
    degrees = int(abs_val)
    minutes_float = (abs_val - degrees) * 60
    minutes = int(minutes_float)
    seconds_num = int(round((minutes_float - minutes) * 60 * 10000))
    return ((degrees, 1), (minutes, 1), (seconds_num, 10000))

def _load_exif_safe(raw: bytes) -> dict:
    try:
        return piexif.load(raw)
    except Exception:
        return {"0th": {}, "Exif": {}, "GPS": {}, "1st": {}}

STRING_FIELDS = {
    "make": (piexif.ImageIFD.Make, "0th"),
    "model": (piexif.ImageIFD.Model, "0th"),
}

@form_data
def process_file_form(data):
    id = int(data["id"])
    action = data["action"]
    duplicate = data.get("duplicate", "off") == "on"
    match action:
        case "format":
            format = data["format_select"]
            meta = get_metadata(id)
            filename = Path(meta["fileName"]).stem
            tags = get_image_tags(id)
            tags.append(f"converted:{format}")

            with WandImage(blob=get_image_file(id)) as img:
                img_bin = img.make_blob(format)
                actions = [AddRawImageWithTagsAction(f"conv-{filename}", format, img_bin, tags)]
                if not duplicate:
                    actions.append(DeleteImageAction(id))
                return MultipartAction(*actions)
        case "metadata":
            meta = get_metadata(id)
            filename = Path(meta["fileName"]).stem
            tags = get_image_tags(id)
            tags.append(f"converted:metadata")
            raw = get_image_file(id)
            exif_dict = _load_exif_safe(raw) if raw else {"0th": {}, "Exif": {}, "GPS": {}, "1st": {}}

            exif_dict = _load_exif_safe(raw)

            for field, (tag, ifd) in STRING_FIELDS.items():
                val = data.get(field, "").strip()
                if val:
                    exif_dict[ifd][tag] = val.encode("utf-8")
                else:
                    exif_dict[ifd].pop(tag, None)
            
            dt_val = data.get("date_taken", "").strip()
            if dt_val and len(dt_val) >= 16:
                try:
                    dt_exif = dt_val[:10].replace("-", ":") + " " + dt_val[11:16] + ":00"
                    dt_bytes = dt_exif.encode("utf-8")
                    exif_dict["Exif"][piexif.ExifIFD.DateTimeOriginal] = dt_bytes
                    exif_dict["0th"][piexif.ImageIFD.DateTime] = dt_bytes
                except Exception as e:
                    log(f"[exif-editor] Could not parse datetime '{dt_val}': {e}")
            else:
                exif_dict["Exif"].pop(piexif.ExifIFD.DateTimeOriginal, None)
                exif_dict["0th"].pop(piexif.ImageIFD.DateTime, None)
            
            lat_str = data.get("latitude", "").strip()
            lon_str = data.get("longitude", "").strip()
            alt_str = data.get("altitude", "").strip()
            if lat_str and lon_str:
                try:
                    lat = float(lat_str)
                    lon = float(lon_str)
                    alt = float(alt_str) if alt_str else 0.0
                    alt_rational = (int(abs(alt) * 10), 10)
                    exif_dict["GPS"] = {
                        piexif.GPSIFD.GPSVersionID:    (2, 3, 0, 0),
                        piexif.GPSIFD.GPSLatitudeRef:  b"N" if lat >= 0 else b"S",
                        piexif.GPSIFD.GPSLatitude:     _decimal_to_dms(lat),
                        piexif.GPSIFD.GPSLongitudeRef: b"E" if lon >= 0 else b"W",
                        piexif.GPSIFD.GPSLongitude:    _decimal_to_dms(lon),
                        piexif.GPSIFD.GPSAltitudeRef:  b"\x00" if alt >= 0 else b"\x01",
                        piexif.GPSIFD.GPSAltitude:     alt_rational,
                    }
                except Exception as e:
                    log(f"[exif-editor] Could not encode GPS: {e}")
            else:
                exif_dict["GPS"] = {}
            
            try:
                exif_bytes = piexif.dump(exif_dict)

                src_img = Image.open(io.BytesIO(raw))
                out = io.BytesIO()
                if src_img.mode not in ("RGB", "L"):
                    src_img = src_img.convert("RGB")
                src_img.save(out, format="JPEG", exif=exif_bytes, quality=95, subsampling=0)
                modified_bytes = out.getvalue()
            except Exception as e:
                log(f"[exif-editor] Failed to write EXIF: {e}")
                return NoAction()

            actions = [AddRawImageWithTagsAction(filename, "jpeg", modified_bytes, tags)]
            if not duplicate:
                actions.append(DeleteImageAction(id))
            return MultipartAction(*actions)
        case _:
            return Error(f"unsupported action: {action}")


if __name__ == "__main__":
    freezetag.run()