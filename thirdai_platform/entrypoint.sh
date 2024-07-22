$EXPORT_IMAGE_NAMES_COMMAND

# Finally, run the Uvicorn server
uvicorn models_app:app --reload --host 0.0.0.0 --port 80