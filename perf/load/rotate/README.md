Config-changer has a rotate script that does the following

1. Fetch cert and key from citadel
2. Generate short lived certificate
3. Rotate cert

The script requires `generate_cert` tool from the security repository.
The Dockerfile contained here builds the image.
