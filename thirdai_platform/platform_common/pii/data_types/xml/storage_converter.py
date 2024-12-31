from typing import List, Tuple

from platform_common.pii.data_types.pydantic_models import CharSpan, XMLUserFeedback
from platform_common.pii.data_types.xml.parser import XMLParser
from platform_common.pii.data_types.xml.utils import remove_special_characters
from platform_common.thirdai_storage.data_types import XMLFeedbackData, XMLLogData


def convert_charspan_to_token_span(charspan: CharSpan, string: str) -> Tuple[int, int]:
    cleaned_string = remove_special_characters(string)
    tokens = cleaned_string.split()

    # Keep track of character positions for each token
    char_pos = 0
    token_starts = []
    token_ends = []  # Add this to track where each token ends

    # Build mapping of character positions to token positions
    for token in tokens:
        # Skip whitespace
        while char_pos < len(cleaned_string) and cleaned_string[char_pos].isspace():
            char_pos += 1
        token_starts.append(char_pos)
        char_pos += len(token)
        token_ends.append(char_pos)  # Store where each token ends

    # Find start token
    start_token = 0
    for i, pos in enumerate(token_starts):
        if pos <= charspan.start < pos + len(tokens[i]):
            start_token = i
            break

    # Find end token
    end_token = start_token
    for i, pos in enumerate(token_starts[start_token:], start_token):
        if charspan.end <= pos:  # If span ends before this token starts
            break
        if token_ends[i] <= charspan.end:  # If span includes this whole token
            end_token = i + 1
        elif pos <= charspan.end < token_ends[i]:  # If span ends within this token
            end_token = i + 1
            break

    return (start_token, end_token)


def convert_xml_feedback_to_storage_format(
    user_feedback: XMLUserFeedback,
) -> Tuple[XMLLogData, List[XMLFeedbackData]]:
    xml_string = user_feedback.xml_string
    feedbacks = user_feedback.feedbacks

    parsed_xml = XMLParser(xml_string, remove_delimiters=False)

    elements = parsed_xml.find_all_elements()

    feedback_data = []
    for feedback in feedbacks:
        xpath = feedback.location.xpath
        attribute = feedback.location.attribute

        xml_reference = parsed_xml.tree.xpath(xpath)

        assert len(xml_reference) == 1

        xml_reference = xml_reference[0]

        if attribute:
            search_string = xml_reference.attrib[attribute]
        else:
            search_string = xml_reference.text

        token_span = convert_charspan_to_token_span(feedback.charspan, search_string)

        feedback_data.append(
            XMLFeedbackData(
                xpath=xpath,
                attribute=attribute,
                token_start=token_span[0],
                token_end=token_span[1],
                n_tokens=len(search_string.split()),
                label=feedback.label,
                user_provided=True,
            )
        )
    log = XMLLogData(xml_string=xml_string, elements=elements)

    return log, feedback_data
