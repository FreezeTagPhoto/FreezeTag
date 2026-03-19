## Usage

This plugin utilizes the Google Gemini API to automatically generate descriptive tags for images. 

### Configuration

You must provide an API key to authenticate with Gemini. The plugin will automatically traverse up the directory tree to load the first `.env` file it finds.

1. Create a `.env` file in any parent directory.
2. Add your API key using the format below:

```env
GEMINI_API_KEY="your_api_key_here"
```
