export function isXML(input: string): boolean {
  try {
    const parser = new DOMParser();
    const parsedDoc = parser.parseFromString(input, 'application/xml');

    // Check if the parsed document has parser errors
    const parserError = parsedDoc.getElementsByTagName('parsererror');
    return parserError.length === 0;
  } catch (error) {
    // If an error occurs, it's not valid XML
    return false;
  }
}
