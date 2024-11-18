import pytest
from platform_common.pii.logtypes.xml.logtype import XMLTokenClassificationLog

pytestmark = [pytest.mark.unit]

xml_string = """
<Event>
  <System>
    <Data name="first_name">
      Shubh
    </Data>
    <Data name="last_name">
      Gupta
    </Data>
  </System>
  <Message>
    I work at {ThirdAI Corp}.
  </Message>
</Event>
"""


def test_xml_logtype():
    log = XMLTokenClassificationLog(log=xml_string)

    inference_sample = log.inference_sample

    assert inference_sample == {
        "source": "<System> <Data> name first_name <System> <Data> Shubh <System> <Data> name last_name <System> <Data> Gupta <Event> <Message> I work at ThirdAI Corp ."
    }

    predictions = [[("O", 1.0)] for token in inference_sample["source"].split()]
    predictions[6] = [("FIRST_NAME", 1.0)]
    predictions[13] = [("LAST_NAME", 1.0)]
    predictions[19] = [("COMPANY", 1.0)]
    predictions[20] = [("COMPANY", 1.0)]

    predictions = log.process_prediction(model_predictions=predictions)

    assert predictions.log_type == "xml"
    assert (
        predictions.query_text
        == '<Event>\n  <System>\n    <Data name="first_name">\n      Shubh\n    </Data>\n    <Data name="last_name">\n      Gupta\n    </Data>\n  </System>\n  <Message>\n    I work at {ThirdAI Corp}.\n  </Message>\n</Event>'
    )

    actual_predictions = [
        {
            "label": "FIRST_NAME",
            "location": {
                "char_span": {"start": 54, "end": 59},
                "xpath_location": {"xpath": "/Event/System/Data[1]", "attribute": None},
                "value": "Shubh",
            },
        },
        {
            "label": "LAST_NAME",
            "location": {
                "char_span": {"start": 106, "end": 111},
                "xpath_location": {"xpath": "/Event/System/Data[2]", "attribute": None},
                "value": "Gupta",
            },
        },
        {
            "label": "COMPANY",
            "location": {
                "char_span": {"start": 163, "end": 175},
                "xpath_location": {"xpath": "/Event/Message", "attribute": None},
                "value": "ThirdAI Corp",
            },
        },
    ]

    assert len(predictions.predictions) == len(actual_predictions)
    for pred, actual_pred in zip(predictions.predictions, actual_predictions):
        assert pred.label == actual_pred["label"]
        assert pred.location.char_span == actual_pred["location"]["char_span"]
        assert pred.location.xpath_location == actual_pred["location"]["xpath_location"]
        assert pred.location.value == actual_pred["location"]["value"]
