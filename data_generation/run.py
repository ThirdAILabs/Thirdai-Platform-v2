import os
from dataclasses import asdict

from variables import DataCategory, GeneralVariables, TokenGenerationVariables

# Load general variables from environment
general_variables: GeneralVariables = GeneralVariables.load_from_env()


def main():
    """
    Main function to initialize and generate the data based on environment variables.
    """
    if general_variables.data_category == DataCategory.text:
        from text_data_generator import TextDataFactory
        from variables import TextGenerationVariables

        factory = TextDataFactory(api_key=general_variables.genai_key)
        args = TextGenerationVariables.load_from_env()
    else:
        from token_data_generator import TokenDataFactory
        from variables import TokenGenerationVariables

        factory = TokenDataFactory(api_key=general_variables.genai_key)
        args = TokenGenerationVariables.load_from_env()

    factory.generate(general_variables.save_dir, asdict(args))

    factory


if __name__ == "__main__":
    main()
