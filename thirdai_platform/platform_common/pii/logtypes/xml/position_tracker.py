import json
import re

from lxml import etree


class PositionTrackingTarget:
    def __init__(self, xml_string):
        self.positions = []
        self.xml_string = xml_string
        self.element_stack_elements = []
        self.total_offset = 0  # Total character offset
        self.pos = 0  # Current position in xml_string
        self.tag_counter_stack = []  # Stack to track tag counts at each level

    def _update_position(self, length):
        self.total_offset += length
        self.pos += length

    def start(self, tag, attrib):
        # Record the start position
        start_offset = self.total_offset

        # Ignore namespaces and use only local names
        namespace_uri, local_name = self._split_namespace(tag)
        tag_name = local_name

        # Find the exact start tag in the xml_string
        tag_end_match = re.search(r">", self.xml_string[self.pos :])
        if not tag_end_match:
            raise ValueError("Cannot find end of start tag")
        tag_end_pos = self.pos + tag_end_match.end()
        tag_text = self.xml_string[self.pos : tag_end_pos]

        # Determine if self-closing tag
        is_self_closing = tag_text.strip().endswith("/>")

        # Update total_offset and pos
        self._update_position(len(tag_text))

        # Update tag counters
        if not self.tag_counter_stack:
            # We are at root level
            self.tag_counter_stack.append({})
        current_tag_counts = self.tag_counter_stack[-1]

        count = current_tag_counts.get(tag_name, 0) + 1
        current_tag_counts[tag_name] = count

        # Build current xpath
        xpath_parts = []
        for elem in self.element_stack_elements:
            xpath_parts.append(f"{elem['name']}[{elem['count']}]")
        xpath_parts.append(f"{tag_name}[{count}]")
        current_xpath = "/" + "/".join(xpath_parts)

        # Now, extract attributes with positions
        attrib_positions = {}

        # Exclude the tag name at the beginning
        tag_name_match = re.match(r"<\s*" + re.escape(tag_name), tag_text)
        if tag_name_match:
            attr_text = tag_text[tag_name_match.end() : -1]  # Exclude '>' at the end
            if is_self_closing:
                attr_text = attr_text.rstrip("/")  # Exclude '/' for self-closing tags
            attr_offset = start_offset + tag_name_match.end()
        else:
            attr_text = tag_text[1:-1]  # Exclude '<' and '>' at the ends
            attr_offset = start_offset + 1

        # Now find attributes in attr_text
        attr_pattern = re.compile(r'\s+(\S+?)\s*=\s*(["\'])(.*?)\2', re.DOTALL)
        for match in attr_pattern.finditer(attr_text):
            full_attr_name = match.group(1)
            attr_value = match.group(3)
            # Positions in attr_text for the attribute value
            value_start_in_attr_text = match.start(3)
            value_end_in_attr_text = match.end(3)
            value_start = attr_offset + value_start_in_attr_text
            value_end = attr_offset + value_end_in_attr_text

            # Ignore attribute namespaces as well
            _, attr_name = self._split_namespace(full_attr_name)

            attrib_positions[attr_name] = {
                "start": value_start,
                "end": value_end,
                "value": attr_value,
            }

        # Initialize attrib[None] for text node
        attrib_positions[None] = None  # Will be updated in 'data' method

        # Create element entry
        element_entry = {
            "type": "element",
            "name": tag_name,
            "count": count,
            "attrib": attrib_positions,
            "start_offset": start_offset,
            "end_offset": None,  # To be updated in 'end' method
            "xpath": current_xpath,
            "is_self_closing": is_self_closing,
            "text_start": None,  # Start position of text content
            "text_end": None,  # End position of text content
        }

        # Append to positions
        self.positions.append(element_entry)

        # Push onto element stack
        self.element_stack_elements.append(element_entry)

        # Append a new dict to tag_counter_stack for the next level
        self.tag_counter_stack.append({})

    def end(self, tag):
        # Pop the dict from tag_counter_stack
        self.tag_counter_stack.pop()

        # Pop the element entry from the stack
        element_entry = self.element_stack_elements.pop()
        is_self_closing = element_entry["is_self_closing"]

        # Use local name for tag
        namespace_uri, local_name = self._split_namespace(tag)
        tag_name = local_name

        if not is_self_closing:
            # Find the exact end tag in the xml_string
            end_tag_pattern = r"</\s*" + re.escape(tag_name) + r"\s*>"
            pattern = re.compile(end_tag_pattern, re.DOTALL)
            match = pattern.match(self.xml_string, self.pos)
            if match:
                tag_text = match.group(0)
                # Update total_offset and pos
                self._update_position(len(tag_text))
            else:
                # Fallback: find the next '>' character
                gt_pos = self.xml_string.find(">", self.pos)
                if gt_pos == -1:
                    raise ValueError("Cannot find end of end tag")
                tag_text = self.xml_string[self.pos : gt_pos + 1]
                self._update_position(len(tag_text))
        else:
            # For self-closing tags, end_offset is already set
            pass

        # Set the end_offset
        element_entry["end_offset"] = self.total_offset

        # If text positions are recorded, update attrib[None]
        if element_entry["text_start"] is not None:
            element_entry["attrib"][None] = {
                "start": element_entry["text_start"],
                "end": element_entry["text_end"],
                "value": self.xml_string[
                    element_entry["text_start"] : element_entry["text_end"]
                ],
            }

    def data(self, data):
        data_length = len(data)
        if data.strip():
            start_offset = self.total_offset
            self._update_position(data_length)
            end_offset = self.total_offset
            # Get the current element to add the text node
            if self.element_stack_elements:
                current_element = self.element_stack_elements[-1]
                if current_element["text_start"] is None:
                    # First text node within this element
                    current_element["text_start"] = start_offset
                    current_element["text_end"] = end_offset
                else:
                    # Concatenate text nodes
                    current_element["text_end"] = end_offset
            else:
                # Text outside of any element (unlikely in well-formed XML)
                pass
        else:
            self._update_position(data_length)

    def comment(self, text):
        comment_text = f"<!--{text}-->"
        start_offset = self.total_offset
        self._update_position(len(comment_text))
        end_offset = self.total_offset
        self.positions.append(
            {
                "type": "comment",
                "content": text,
                "start_offset": start_offset,
                "end_offset": end_offset,
                "xpath": self.current_xpath,
            }
        )

    def close(self):
        char_spans = {}
        for position in self.positions:
            for attr, details in position["attrib"].items():
                if details is not None:
                    char_spans[(position["xpath"], attr)] = details

        for (xpath, attr), details in char_spans.items():
            value_in_xml = self.xml_string[details["start"] : details["end"]]
            assert (
                value_in_xml == details["value"]
            ), f"parsed value for {xpath=}, {attr=} does not match the value found using char spans in the xml"
        return char_spans

    def _split_namespace(self, tag):
        if "}" in tag:
            namespace_uri, local_name = tag[1:].split("}")
            return None, local_name  # Ignore the namespace URI
        else:
            return None, tag


def parse_xml_with_positions(xml_string):
    parser = etree.XMLParser(
        target=PositionTrackingTarget(xml_string),
        recover=True,  # Continue parsing even if the XML is malformed
        remove_blank_text=False,  # Keep blank text nodes
    )
    try:
        etree.fromstring(xml_string.encode("utf-8"), parser)
    except etree.XMLSyntaxError as e:
        print(f"XML Syntax Error: {e}")
    return parser.target.close()
