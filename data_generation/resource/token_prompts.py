tag_value_prompt = """You possess deep expertise in {domain_prompt}. Please generate unique {num_samples_per_tag} diverse samples for the {tag} named entity. Below are some examples of the {tag} entity:
{tag_example}

Description of the tag {tag}:
{tag_description}

Ensure each sample starts on a new line without any bullet points, prefixes, quotes, or emojis at the beginning of each sentence.

VERY IMPORTANT:
-  Ensure each sample starts on a new line without any bullet points, prefixes, quotes, or emojis at the beginning of each sentence.
-  Only include the attributes themselves in the output. Each attribute should appear on a new line without any additional new lines.
-  Make sure that the samples are relevant to the context of US citizen.

Additionally, aim to cover a wide range of variations within the {tag} entity to ensure the data is as varied and representative as possible.
"""

template_prompt = """You have to generate {k} templatized sentences for the tags: {tags}

Description of the tags:
{tags_description}

As an example here is an example sentences for the tags [NAME, AGE]

[NAME] was born in chicago and is of age [AGE] currently.
where,
 - [NAME] could be 'Chelsey Lobo'
 - [AGE] could be 39.

Key Requirements:
-   Each sentence should start on a new line and with no bulleting, header/footer or any steps involved. 
-   Make sure to include at least two entities in each samples.
-   Make sure to include all the given entity in the templatized sentences.

** IMPORTANT POINT:
-  These Entities would be filled later so make sure these samples would make sense after being filled.
"""

dataset_generation_prompt = """You possess deep expertise in {domain_prompt}. Please generate {num_to_generate} templates of synthetic sentences and associated tags for {domain_prompt}
            
VERY IMPORTANT: MAKE SURE identify all named entities occurred that belong to one of the following entity types: 
{tags}

Description of the tags:
{tag_description}

When generating the output for the NER task, adhere to the following strict format guidelines to ensure data consistency and format correctness

Following are some sample output format for generation. This is just for example and you should not mimic this pattern.

{templatized_sentences_examples}

Key Requirements:
-  Mask only the Entities in square brackets.
-  The entities should strictly belong to one of {tags}. Do not include anything apart from entities in square brackets
-  Make sure that all the tags are being used in each templates.
-  Give only the generated samples in output and make sure each sample should start on a new line. Do not include any extra new line. 
-  DO NOT include any bulleting or header/footer with any samples. Do not include any quotes or emojis.
-  Give equal weightage to all the tags.
-  {rnd_prompts_str}
"""
