tag_value_prompt = """You possess deep expertise in {domain_prompt}. Please generate unique {num_samples_per_tag} diverse values for the {tag} named entity. Below are some values of the {tag} entity:
{tag_example}

Description of the tag {tag}:
{tag_description}
Also Cover a wide variations in the format of the tag {tag} value.
Additionally, aim to cover a wide range of variations within the {tag} entity to ensure the data is as varied and representative as possible.

VERY IMPORTANT:
-  Ensure each value starts on a new line without any bullet points, prefixes, quotes, or emojis at the beginning of each sentence.
-  Make sure that the samples are relevant to the context of US citizen.
"""

template_prompt = """You have to generate {k} templatized sentences for the tags: {tags}

Description of the tags:
{tags_description}

Here are example templates for the tags [CARDHOLDER_NAME, EXPIRATION_DATE] on the domain of payment information.
-  [CARDHOLDER_NAME] and his friend john tried to dupe the credit card company by reporting their transaction with the card [PAN] as fradulent on 9th august.
-  In the month of december, the card was expired but the expiration date mentioned was [EXPIRATION_DATE].

Key Requirements:
-  Each template should start on a new line and with no bulleting, header/footer or any steps involved. 
-  Make sure to include at least two tags in each sentence.

** IMPORTANT POINT:
-  These templates would be filled later so make sure these templates would make sense after being filled. Here is one incorrect & correct templates for the tags [MEDICAL_INFO]
      Incorrect templates: My [MEDICAL_INFO] should remain confidential to protect my personal interest.
      Correct templates: My condition due to [MEDICAL_INFO] should remain confidential to protect my personal interest.
"""

dataset_generation_prompt = """The goal is to create a dataset for entity recognition. Please generate {num_to_generate} templates associated with given below tags for {domain_prompt}
            
Tags with their description and example:
{tags_info}

Following are some sample output format for generation. This is just for example and you should not mimic this pattern.

For example, here are some templates for the tags [CARDHOLDER_NAME, EXPIRATION_DATE] on the domain of payment information.
- [CARDHOLDER_NAME] and his friend john tried to dupe the credit card company by reporting their transaction with the card [PAN] as fradulent on 9th august.
- In the month of december, the card was expired but the expiration date mentioned was [EXPIRATION_DATE].

Key Requirements:
- Include words that could be interpreted as tag but are actually not, as depicted in the above examples.
- {value_requirements}

Output format:
-  Each template should be in a newline.
-  DO NOT include any bulleting, header/footer or enumeration. Do not include any quotes or emojis.

** IMPORTANT POINT:
-  These templates would be filled later so make sure these templates would make sense after being filled. Here is one incorrect & correct templates for the tags [MEDICAL_INFO]
      Incorrect templates: My [MEDICAL_INFO] should remain confidential to protect my personal interest.
      Correct templates: My condition due to [MEDICAL_INFO] should remain confidential to protect my personal interest.
"""
