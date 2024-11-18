import pytest
from platform_common.pii.logtypes.xml.position_tracker import parse_xml_with_positions

pytestmark = [pytest.mark.unit]


def test_text_extraction():
    xml_string = "<root><parent><child>Content</child></parent></root>"
    char_spans = parse_xml_with_positions(xml_string)

    expected = {
        ("/root[1]/parent[1]/child[1]", None): {
            "value": "Content",
            "start": 21,
            "end": 28,
        },
    }
    assert char_spans == expected


def test_attribute_extraction():
    xml_string = "<root><element attr1='value1' attr2='value2'></element></root>"
    char_spans = parse_xml_with_positions(xml_string)

    expected = {
        ("/root[1]/element[1]", "attr1"): {"value": "value1", "start": 22, "end": 28},
        ("/root[1]/element[1]", "attr2"): {"value": "value2", "start": 37, "end": 43},
    }
    assert char_spans == expected


def test_self_closing_tags():
    xml_string = "<root><empty-element attr='value'/></root>"
    char_spans = parse_xml_with_positions(xml_string)

    expected = {
        ("/root[1]/empty-element[1]", "attr"): {
            "value": "value",
            "start": 27,
            "end": 32,
        },
    }
    assert char_spans == expected


def test_comments_handling():
    xml_string = "<root><!-- This is a comment --><child>Text</child></root>"
    char_spans = parse_xml_with_positions(xml_string)

    expected_positions = {
        ("/root[1]/child[1]", None): {"start": 39, "end": 43, "value": "Text"}
    }
    assert char_spans == expected_positions


def test_text_content_extraction():
    # we ignore the .tail text for the elements
    xml_string = "<root>Leading text<child>Child text</child>Trailing text</root>"
    char_spans = parse_xml_with_positions(xml_string)

    expected = {
        ("/root[1]", None): {"start": 6, "end": 18, "value": "Leading text"},
        ("/root[1]/child[1]", None): {"start": 25, "end": 35, "value": "Child text"},
    }
    assert char_spans == expected


def test_whitespace_handling():
    xml_string = (
        "<root>\n    <child>\n        Text with whitespace\n    </child>\n</root>"
    )
    char_spans = parse_xml_with_positions(xml_string)

    expected = {
        ("/root[1]/child[1]", None): {
            "value": "\n        Text with whitespace\n    ",
            "start": 18,
            "end": 52,
        },
    }
    assert char_spans == expected


def test_namespaces_ignored():
    xml_string = (
        "<ns:root xmlns:ns='http://example.com/ns'><ns:child>Text</ns:child></ns:root>"
    )
    char_spans = parse_xml_with_positions(xml_string)

    expected = {
        ("/root[1]", "xmlns:ns"): {
            "start": 19,
            "end": 40,
            "value": "http://example.com/ns",
        },
        ("/root[1]/child[1]", None): {"start": 52, "end": 56, "value": "Text"},
    }
    assert char_spans == expected


def test_entity_references():
    xml_string = "<root>Value &amp; more text</root>"
    char_spans = parse_xml_with_positions(xml_string)

    expected = {
        ("/root[1]", None): {"value": "Value & more text", "start": 6, "end": 23},
    }
    assert char_spans == expected


def test_malformed_xml_recovery():
    xml_string = "<root><child>Text</child><unclosed-root>"
    char_spans = parse_xml_with_positions(xml_string)

    # Should recover and parse up to the error
    expected = {
        ("/root[1]/child[1]", None): {"value": "Text", "start": 13, "end": 17},
    }
    assert char_spans == expected


def test_duplicate_tags():
    xml_string = "<root><item>First</item><item>Second</item></root>"
    char_spans = parse_xml_with_positions(xml_string)

    expected = {
        ("/root[1]/item[1]", None): {"value": "First", "start": 12, "end": 17},
        ("/root[1]/item[2]", None): {"value": "Second", "start": 30, "end": 36},
    }
    assert char_spans == expected


def test_complex_xml():
    xml_string = """
    <Event>
        <System>
            <Data name="first_name">John</Data>
            <Data name="last_name">Doe</Data>
        </System>
        <Message>
            Hello, this is a test message.
        </Message>
        <!-- Comment -->
        <EmptyElement attr="value"/>
    </Event>
    """
    char_spans = parse_xml_with_positions(xml_string)

    expected = {
        ("/Event[1]/System[1]/Data[1]", "name"): {
            "value": "first_name",
            "start": 58,
            "end": 68,
        },
        ("/Event[1]/System[1]/Data[1]", None): {
            "value": "John",
            "start": 70,
            "end": 74,
        },
        ("/Event[1]/System[1]/Data[2]", "name"): {
            "value": "last_name",
            "start": 102,
            "end": 111,
        },
        ("/Event[1]/System[1]/Data[2]", None): {
            "value": "Doe",
            "start": 113,
            "end": 116,
        },
        ("/Event[1]/Message[1]", None): {
            "value": "\n            Hello, this is a test message.\n        ",
            "start": 140,
            "end": 189,
        },
        ("/Event[1]/EmptyElement[1]", "attr"): {
            "value": "value",
            "start": 223,
            "end": 230,
        },
    }
    assert char_spans == expected
