from PIL import Image

import freezetag
from freezetag.hooks import single_image, init_func, AddTagsAction, Error
from freezetag.message import log, read_config
import requests
import io

def identify_plant(image: Image.Image):
    api_token = read_config("config.toml")["plantnet_key"]
    url = f"https://my-api.plantnet.org/v2/identify/all?api-key={api_token}"
    
        
    buffer = io.BytesIO()
    image.convert("RGB").save(buffer, format="JPEG", quality=90)
    buffer.seek(0)
    files = [('images', ('image.jpg', buffer, 'image/jpeg'))]
    data = {'organs': ['auto']}

    
    response = requests.post(url, files=files, data=data)
    
    if response.status_code == 200:
        json_data = response.json()
        results = json_data.get('results', [])
        
        if results:
            species = results[0]['species']
            # try and get a common name, default if it doesn't exist to the scientific name
            return species.get('commonNames', [species['scientificNameWithoutAuthor']])[0]
            
    elif response.status_code == 404:
        log(f"Plant not identified for image: {response.text}")
    else:
        log(f"API Error {response.status_code}: {response.text}")
        
    return None

@single_image
def tag_image(img: Image.Image, id: int):
    try:
        plant_name = identify_plant(img)
        log(f"Identified plant for image {id}: {plant_name}")
        if plant_name:
            return AddTagsAction(id, [plant_name])
        else:
            return AddTagsAction(id, ["Unknown Plant"])
    except Exception as e:
        log(f"Error processing image {id}: {e}")
        return Error(f"Error processing image {id}: {e}")

if __name__ == "__main__":
    freezetag.run()
