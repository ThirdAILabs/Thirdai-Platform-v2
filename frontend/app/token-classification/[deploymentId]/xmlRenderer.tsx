// import React from 'react';

// interface XMLRendererProps {
//     xmlContent: string;
// }

// const XMLRenderer = ({ xmlContent }: XMLRendererProps) => {
//     const formatXML = (xml: string): string => {
//         let formatted = '';
//         let indent = 0;
//         const tab = '  '; // 2 spaces for indentation
//         const tokens = xml.trim().split(/(<\/?[^>]+>)/g);

//         tokens.forEach((token) => {
//             if (!token.trim()) return; // Skip empty tokens

//             // Check if it's a closing tag
//             if (token.startsWith('</')) {
//                 indent--;
//                 formatted += tab.repeat(Math.max(0, indent)) + token + '\n';
//             }
//             // Check if it's an opening tag
//             else if (token.startsWith('<') && !token.startsWith('<?') && !token.endsWith('/>')) {
//                 formatted += tab.repeat(indent) + token + '\n';
//                 indent++;
//             }
//             // Self-closing tag
//             else if (token.startsWith('<') && token.endsWith('/>')) {
//                 formatted += tab.repeat(indent) + token + '\n';
//             }
//             // Text content
//             else if (token.trim()) {
//                 formatted += tab.repeat(indent) + token.trim() + '\n';
//             }
//         });

//         return formatted.trim();
//     };

//     const highlightXML = (formattedXML: string): string => {
//         return formattedXML
//             .replace(/&/g, '&amp;')
//             .replace(/</g, '&lt;')
//             .replace(/>/g, '&gt;')
//             // Highlight tag names
//             .replace(/&lt;(\/?[\w:-]+)/g, '&lt;<span class="text-blue-500">$1</span>')
//             // Highlight attributes
//             .replace(/(\s+[\w:-]+)=/g, '<span class="text-purple-400">$1</span><span class="text-yellow-500">=</span>')
//             // Highlight attribute values
//             .replace(/="([^"]*?)"/g, '="<span class="text-green-500">$1</span>"')
//             // Highlight comments
//             .replace(/&lt;!--[\s\S]*?--&gt;/g, '<span class="text-gray-500">$&</span>');
//     };

//     return (
//         <div className="w-full max-w-4xl">
//             <pre className="p-4 rounded-lg overflow-x-auto">
//                 <code
//                     className="text-black block font-mono text-sm leading-6 whitespace-pre"
//                     dangerouslySetInnerHTML={{
//                         __html: highlightXML(formatXML(xmlContent))
//                     }}
//                 />
//             </pre>
//         </div>
//     );
// };

// export default XMLRenderer;


import React from 'react';

interface XMLRendererProps {
    xmlContent: string;
    predictions: Array<{
        label: string;
        location: {
            char_span: {
                start: number;
                end: number;
            };
            xpath_location: {
                xpath: string;
                attribute: string | null;
            };
            value: string;
        };
    }>;
}

const XMLRenderer = ({ xmlContent, predictions }: XMLRendererProps) => {


    const highlightXML = (formattedXML: string): string => {
        let highlightedXML = '';
        let currentIndex = 0;

        // predictions.sort((a, b) => a.location.char_span.start - b.location.char_span.start);

        predictions.forEach((prediction) => {
            const { start, end } = prediction.location.char_span;
            const value = prediction.location.value;
            const label = prediction.label;
            console.log("start and end index -> ", start, " ", end);
            // Add content before the prediction span
            highlightedXML += formattedXML.substring(currentIndex, start)
                .replace(/&/g, '&amp;')
                .replace(/</g, '&lt;')
                .replace(/>/g, '&gt;');

            // Validate and highlight the span
            const spanText = formattedXML.substring(start, end);
            console.log("spanText, -> ", spanText);
            if (spanText === value) {
                highlightedXML += `
          <span class="bg-yellow-200 relative">
            ${spanText.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')}
            <span class="absolute text-xs text-red-500 top-0 right-0">${label}</span>
          </span>
        `;
            } else {
                highlightedXML += spanText.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
            }

            currentIndex = end;
        });

        // Add remaining content
        highlightedXML += formattedXML.substring(currentIndex)
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;');

        return highlightedXML;
    };

    return (
        <div className="w-full max-w-4xl">
            <pre className="p-4 rounded-lg overflow-x-auto">
                <code
                    className="text-black block font-mono text-sm leading-6 whitespace-pre"
                    dangerouslySetInnerHTML={{
                        __html: highlightXML(xmlContent)
                    }}
                />
            </pre>
        </div>
    );
};

export default XMLRenderer;
