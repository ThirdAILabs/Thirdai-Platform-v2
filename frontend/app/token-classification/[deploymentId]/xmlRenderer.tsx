import React from 'react';
import XMLViewer from 'react-xml-viewer';

interface CharSpan {
  start: number;
  end: number;
}

interface XPathLocation {
  xpath: string;
  attribute: string | null;
}

interface Location {
  char_span: CharSpan;
  xpath_location: XPathLocation;
  value: string;
}

interface Prediction {
  label: string;
  location: Location;
}

interface XMLHighlighterProps {
  xmlText: string;
  predictions: Prediction[];
}

interface Theme {
  tagColor: string;
  textColor: string;
  attributeKeyColor: string;
  attributeValueColor: string;
  separatorColor: string;
  commentColor: string;
  cdataColor: string;
  fontFamily: string;
}

const XMLRenderer: React.FC<XMLHighlighterProps> = ({ xmlText, predictions }) => {
  // Preprocess the XML to wrap highlights
  const preprocessXML = (text: string): string => {
    const sortedPredictions = [...predictions].sort(
      (a, b) => a.location.char_span.start - b.location.char_span.start
    );

    let processedText = '';
    let lastIndex = 0;

    sortedPredictions.forEach((prediction) => {
      const { start, end } = prediction.location.char_span;
      const label = prediction.label;

      // Append unmodified text before the highlight
      if (start > lastIndex) {
        processedText += text.slice(lastIndex, start);
      }

      // Append the highlighted text
      const textToHighlight = text.slice(start, end);
      processedText +=
        `<span style="background-color: yellow;">${textToHighlight}</span>` +
        `<span style="background-color: red; color: white; padding: 2px; border-radius: 3px; margin-left: 4px;">${label}</span>`;

      lastIndex = end;
    });

    // Append remaining unmodified text
    if (lastIndex < text.length) {
      processedText += text.slice(lastIndex);
    }

    return processedText;
  };

  // Apply preprocessing to XML
  const highlightedXML = preprocessXML(xmlText);
  // Custom theme for XMLViewer
  const customTheme: Theme = {
    tagColor: '#000bd4',
    textColor: '#000000',
    attributeKeyColor: '#2a7ab0',
    attributeValueColor: '#008000',
    separatorColor: '#333',
    commentColor: '#aaa',
    cdataColor: '#1d781d',
    fontFamily: 'monospace',
  };

  return (
    <div className="w-full">
      <div className="font-mono text-sm whitespace-pre-wrap">
        <XMLViewer
          xml={highlightedXML} // Keep the original XML for the viewer
          theme={{
            ...customTheme,
          }}
          indentSize={2}
          collapsible={true}
        />
      </div>

      {/* Higlighted text */}
      <div className="mt-4 p-4 border rounded">
        <h3 className="font-medium mb-2">Highlighted text only:</h3>
        <div className="flex flex-col gap-2">
          {predictions.map((prediction) => (
            <div key={prediction.location.char_span.start} className="flex items-center">
              <span className="rounded px-2 ml-2" style={{ backgroundColor: '#fff59d' }}>
                {prediction.location.value}
              </span>
              <span className="text-red-500">{prediction.label}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default XMLRenderer;
