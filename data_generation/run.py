import os
from dataclasses import asdict

from variables import DataCategory, GeneralVariables

# Load general variables from environment
general_variables: GeneralVariables = GeneralVariables.load_from_env()

def main():
    """
    Main function to initialize and generate the data based on environment variables.
    """
    if general_variables.data_category == DataCategory.text:
        from variables import TextGenerationVariables

        from data_generation.text_data_factory import TextDataFactory

        factory = TextDataFactory(api_key=general_variables.genai_key)
        args = TextGenerationVariables.load_from_env()

        factory.generate(
            task_prompt=args.task_prompt,
            samples_per_label=args.samples_per_label,
            target_labels=args.target_labels,
            user_vocab=args.user_vocab,
            examples=args.examples,
            user_prompts=args.user_prompts,
            labels_description=args.labels_description,
            batch_size=args.batch_size,
            vocab_per_sentence=args.vocab_per_sentence,
        )
    else:
        from variables import TokenGenerationVariables

        from data_generation.token_data_factory import TokenDataFactory

        factory = TokenDataFactory(api_key=general_variables.genai_key)
        args = TokenGenerationVariables.load_from_env()
        
        factory.generate(
            domain_prompt=args.domain_prompt,
            tags=args.tags,
            tag_examples=args.tag_examples,
            num_call_batches=args.num_call_batches,
            batch_size=args.batch_size,
            num_samples_per_tag=args.num_samples_per_tag,
        )


if __name__ == "__main__":
    main()
