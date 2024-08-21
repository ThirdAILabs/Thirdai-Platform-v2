import os
from dataclasses import asdict

from utils import save_dict
from variables import DataCategory, GeneralVariables
from urllib.parse import urljoin
import requests
import io
import json

# Load general variables from environment
general_variables: GeneralVariables = GeneralVariables.load_from_env()

def launch_train_job(file_location: str, train_args: str):
    try:
        api_url = general_variables.model_bazaar_endpoint
        headers = {"User-Agent": "Datagen job"}
        url = urljoin(api_url, "api/train/udt-impl")
        empty_file = io.BytesIO(b"")
        empty_file.name = file_location
        data = {
            "args_json": train_args,
            "file_details_list": json.dumps({"file_details": [{"mode": "supervised", "location": "nfs", "is_folder": False}]})
        }
        files = {
            'files': (file_location, empty_file, 'text/plain')
        }
        response = requests.request("post", url, headers=headers, data=data, files=files)
        response.raise_for_status()
        return response.json()
    except requests.exceptions.RequestException as exception:
        print(exception)
        raise exception


def main():
    """
    Main function to initialize and generate the data based on environment variables.
    """
    if general_variables.data_category == DataCategory.text:
        from text_data_factory import TextDataFactory
        from variables import TextGenerationVariables

        factory = TextDataFactory()
        args = TextGenerationVariables.load_from_env()

    else:
        from token_data_factory import TokenDataFactory
        from variables import TokenGenerationVariables

        factory = TokenDataFactory()
        args = TokenGenerationVariables.load_from_env()

    # Saving the args first
    save_dict(
        factory.generation_args_location,
        **{"data_id": general_variables.data_id, **asdict(args)}
    )
    dataset_config = factory.generate_data(**asdict(args))
    train_args_dict = json.loads(general_variables.train_args)
    train_args_dict['extra_options']['source_column'] = dataset_config['input_feature']
    train_args_dict['extra_options']['target_column'] = dataset_config['target_feature']
    train_args_dict['extra_options']['target_labels'] = dataset_config['target_labels']
    train_args = json.dumps(train_args_dict)
    launch_train_job(dataset_config['filepath'], train_args)
    # launch_train_job("/path/to/file", json.dumps({"data_id": "123456"}))


if __name__ == "__main__":
    main()
