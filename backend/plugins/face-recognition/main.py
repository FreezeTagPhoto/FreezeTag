import freezetag
from freezetag.hooks import *
from freezetag.message import *
import face_recognition
import numpy as np
from io import BytesIO

known_names = []
known_encodings = []

@init_func
def init():
    global known_names, known_encodings
    faces = search_images(tagsLike=["person:"])
    for face_img in faces:
        tags = get_image_tags(face_img)
        name = ""
        for tag in tags:
            if tag.startswith("person:"):
                name = tag.split(":")[1]
                break
        if name == "":
            log("image tagged 'person:' but had no people")
            continue
        log(f"encoding person '{name}'")
        img = get_image(face_img)
        if img is None:
            log(f"couldn't load image for '{name}', skipping")
            continue
        img.thumbnail((1024, 1024))
        img_file = BytesIO()
        img.save(img_file, format='WEBP')
        face_img = face_recognition.load_image_file(img_file)
        person_encoding = face_recognition.face_encodings(face_img, face_recognition.face_locations(face_img))[0]
        known_names.append(name)
        known_encodings.append(person_encoding)
    log("finished setting up")

@single_image
def detect_faces(img, id):
    global known_names, known_encodings
    img.thumbnail((1024, 1024))
    img_file = BytesIO()
    img.save(img_file, format='WEBP')
    face_img = face_recognition.load_image_file(img_file)
    names = []
    for face in face_recognition.face_encodings(face_img, face_recognition.face_locations(face_img)):
        matches = face_recognition.compare_faces(known_encodings, face)
        if True in matches:
            names.append(known_names[matches.index(True)])
    if len(names) == 0:
        log("no known faces in image")
        return NoAction()
    log(f"tagging image with names {names}")
    return AddTagsAction(id, names)

if __name__ == "__main__":
    freezetag.run()