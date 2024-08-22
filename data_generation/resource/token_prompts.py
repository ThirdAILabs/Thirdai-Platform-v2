tag_value_prompt = """You possess deep expertise in {domain_prompt}. Please generate {num_samples_per_tag} diverse samples for the {tag} named entity. Below are some examples of the {tag} entity:
{tag_example}

Description of the tag {tag}:
{tag_description}

Ensure each sample starts on a new line without any bullet points, prefixes, quotes, or emojis at the beginning of each sentence.

VERY IMPORTANT: Only include the attributes themselves in the output. Each attribute should appear on a new line without any additional new lines.

Additionally, aim to cover a wide range of variations within the {tag} entity to ensure the data is as varied and representative as possible.
"""

template_prompt = """You have to generate {k} templatized sentences for the tags: {tags}

Description of the tags:
{tags_description}

As an example here are two sentences for the tags [GENDER,ENTHNICITY,DISABILITY,SSN]

She is [GENDER] and is of [ENTHNICITY] DESCENT 
After getting diagnosed with [DISABILITY] John went home. His social security number is [SSN].

Key Requirements:
-   Each sentence should start on a new line and with no bulleting, header/footer or any steps involved. 
-   Make sure to include at least two entities in each samples.

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
-  The entities should strictly belong to one of {tags}. Make sure to include at least two entities in each samples. Do not include anything apart from entities in square brackets
-  Give only the generated samples in output and make sure each sample should start on a new line. Do not include any extra new line. 
-  DO NOT include any bulleting or header/footer with any samples. Do not include any quotes or emojis.
-  Give equal weightage to all the tags.
-  {rnd_prompts_str}

** IMPORTANT POINT:
-  These Entities would be filled later so make sure these samples would make sense after being filled. Here are some incorrect and correct samples for the tags [SEXUAL_ORIENTATION, GENDER]
      Incorrect sample: My [SEXUAL_ORIENTATION] and [GENDER] should remain confidential to protect my financial interests.
      Correct Sample: My sexuality [SEXUAL_ORIENTATION] and gender [GENDER] should remain confidential to protect my financial interests.
"""
