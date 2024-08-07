attribute_dimension_prompt = """Which attribute dimensions do you consider most vital in determining the topic of the task: ```{domain_prompt}```?

VERY IMPORTANT POINTS:
   1. DO NOT include any bulleting, enumeration, header/footer in any sentences. Do not include any quotes or emojis.
   2. Give only the attributes in output and make sure each attribute should start on a new line. Do not include any extra new line.

List atmost 5 attributes.

Eg. If domain is news, topics can be
Sports
Healthcare
Politics
Stocks
Weather
"""

attribute_value_prompt = """Given your extensive expertise in {domain_prompt}, please provide a range of realistic values for {attribute}. Ensure these estimates are diverse and applicable to real-world scenarios. 
For attributes known to have well-defined values, provide specific, practical estimates; for all other attributes, offer generalized yet realistic ranges.

VERY IMPORTANT POINTS:
   1. DO NOT include any bulleting, enumeration, header/footer in any sentences. Do not include any quotes or emojis.
   2. Give only the attributes in output and make sure each attribute should start on a new line. Do not include any extra new line.
"""

# To generate realistic tag examples for the given user tag
tag_value_prompt = """Please generate {num_samples_per_tag} diverse samples for the {tag} named entity. Below are some examples of the {tag} entity:
{tag_example}

VERY IMPORTANT POINTS:
        - Ensure each sample starts on a new line without any bullet points, prefixes, quotes, or emojis or any header/footer.

VERY IMPORTANT: Only include the attributes themselves in the output. Each attribute should appear on a new line without any additional new lines.

Additionally, aim to cover a wide range of variations within the {tag} entity to ensure the data is as varied and representative as possible.
"""

template_prompt = """You have to generate 2 templatized sentences for the tags: {tags}

As an example here are two sentences for the tags [GENDER,ENTHNICITY,DISABILITY,SSN]

She is [GENDER] and is of [ENTHNICITY] DESCENT 
After getting diagnosed with [DISABILITY] John went home. His social security number is [SSN].

Each sentence should start on a new line and with no bulleting or any header/footer.
"""