import { XMLParser, XMLValidator } from 'fast-xml-parser';
import { ChangeEvent, KeyboardEvent, useEffect, useMemo, useState } from 'react';
import Fuse from 'fuse.js';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import { Underline } from 'lucide-react';
export const ATTRIBUTE_PREFIX = '@_';
export const INDENT = '20px';
export const SPACE = '5px';

export function clean(xmlString: string): string {
  // Find the first '<' and the last '>' characters
  const start = xmlString.indexOf('<');
  const end = xmlString.lastIndexOf('>');

  if (start === -1 || end === -1 || end <= start) {
    throw new Error('Invalid XML Log');
  }

  // Extract the XML content between the first '<' and the last '>'
  const log = xmlString.substring(start, end + 1);

  // Remove non-printable characters (excluding \t, \n, \r)
  const cleanString = log.replace(/[^\x20-\x7E\t\n\r]/g, '');

  return cleanString;
}

export function parseXML(xml: string) {
  const validation = XMLValidator.validate(xml);
  if (validation !== true) {
    throw new Error('Invalid XML');
  }

  const parser = new XMLParser({
    ignoreAttributes: false,
    allowBooleanAttributes: false,
    attributeNamePrefix: ATTRIBUTE_PREFIX,
    alwaysCreateTextNode: true,
  });

  const parsedData = parser.parse(xml);

  // Clean the parsed XML data
  // cleanXMLText(parsedData);

  return parsedData;
}

interface CharSpan {
  start: number;
  end: number;
}

interface XPathLocation {
  xpath: string;
  attribute: string | null;
}

interface Location {
  local_char_span: CharSpan;
  value: string;
}

interface Prediction {
  label: string;
  location: Location;
}

interface XMLRendererProps {
  data: Record<string, any>;
  path: (string | number)[];
  choices: string[];
  predictions: Prediction[];
  onSelectionComplete: (selection: Selection) => void;
}

interface XMLAttributeRendererProps extends XMLRendererProps {
  attr: string;
}

interface XMLValueRendererProps extends XMLRendererProps {
  data: any;
  attr?: string;
}

interface XMLObjectRendererProps extends XMLRendererProps {
  tag: string;
}

interface TagSelectorProps {
  open: boolean;
  choices: string[];
  onSelect: (tag: string) => void;
  predictions: Prediction[];
}

interface Selection {
  start: number;
  end: number;
  xpath: string;
  tag: string;
  value: string;
}

export function TagSelector({ open, choices, onSelect, predictions }: TagSelectorProps) {
  const defaultOptions = useMemo(() => choices.map((label) => ({ label, new: false })), [choices]);

  const [fuse, setFuse] = useState(new Fuse(choices));
  const [options, setOptions] = useState<{ label: string; new: boolean }[]>(defaultOptions);
  const [selectedIndex, setSelectedIndex] = useState<number | null>(null);

  useEffect(() => {
    setFuse(new Fuse(choices));
    setOptions(defaultOptions);
  }, [choices]);

  const handleInputChange = (e: ChangeEvent<HTMLInputElement>) => {
    const query = e.target.value;
    const searchResults = query !== '' ? fuse.search(query).map((val) => val.item) : choices;
    let newOptions = searchResults.map((label) => ({ label, new: false }));
    if (query !== '' && !choices.includes(query)) {
      newOptions.unshift({ label: query, new: true });
    }
    setOptions(newOptions);
    setSelectedIndex(0);
  };

  const selectLabel = (label: string) => {
    onSelect(label);
    setOptions(defaultOptions);
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    e.stopPropagation();
    switch (e.key) {
      case 'ArrowUp':
        console.log('ArrowUp key pressed');
        setSelectedIndex((prev) =>
          prev === null ? options.length - 1 : Math.min(options.length - 1, prev - 1)
        );
        break;
      case 'ArrowDown':
        console.log('ArrowDown key pressed');
        setSelectedIndex((prev) => (prev === null ? 0 : Math.min(options.length, prev + 1)));
        break;
      case 'Enter':
        console.log('Enter key pressed');
        if (selectedIndex !== null) {
          selectLabel(options[selectedIndex].label);
        }
        break;
      default:
        break;
    }
  };

  const makeDropdownMenuItem = (
    index: number,
    label: string,
    isNew: boolean,
    isSelected: boolean
  ) => (
    <DropdownMenuItem
      className={`font-medium ${isSelected ? 'bg-accent' : ''}`}
      key={index}
      onMouseOver={() => setSelectedIndex(index)}
    >
      <button
        style={{ width: '100%', height: '100%', textAlign: 'left' }}
        onClick={() => selectLabel(label)}
      >
        {isNew && (
          <>
            <span
              className="bg-accent font-medium"
              style={{
                padding: '0 3px',
                marginRight: '5px',
                borderRadius: '2px',
                fontWeight: 'bold',
              }}
            >
              New{' '}
            </span>{' '}
          </>
        )}
        {label}
      </button>
    </DropdownMenuItem>
  );

  return (
    <div
      onMouseDown={(e) => {
        e.stopPropagation();
      }}
    >
      <DropdownMenu open={open} modal={false}>
        <DropdownMenuTrigger>
          <span />
        </DropdownMenuTrigger>
        <DropdownMenuContent>
          <Input
            autoFocus
            className="font-medium"
            onChange={handleInputChange}
            style={{ marginBottom: '5px' }}
            onKeyDown={handleKeyDown}
          />
          <div
            style={{ maxHeight: '30vh', overflowY: 'scroll' }}
            onMouseLeave={() => setSelectedIndex(null)}
          >
            {options.map((val, index) =>
              makeDropdownMenuItem(index, val.label, val.new, index === selectedIndex)
            )}
          </div>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}

function XMLAttributeRenderer({
  data,
  path,
  attr,
  choices,
  predictions,
  onSelectionComplete,
}: XMLAttributeRendererProps) {
  const key = attr.substring(ATTRIBUTE_PREFIX.length);
  let dataString = JSON.stringify(data);
  dataString = dataString.substring(0, dataString.length - 1);
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'row',
        gap: '',
        marginLeft: SPACE,
        flexWrap: 'nowrap',
      }}
    >
      {key}=&quot;
      <XMLValueRenderer
        data={dataString}
        path={path}
        attr={key}
        choices={choices}
        predictions={predictions}
        onSelectionComplete={onSelectionComplete}
      />
      &quot;
    </div>
  );
}

function XMLValueRenderer({
  data,
  path,
  choices,
  predictions,
  onSelectionComplete,
}: XMLValueRendererProps) {
  const [start, setStart] = useState<number | null>(null);
  const [end, setEnd] = useState<number | null>(null);
  const [predictionIndex, setIsPredictionIndex] = useState<number>(-1);
  const [isSelecting, setIsSelecting] = useState(false);
  const [tagSelectorOpen, setTagSelectorOpen] = useState(false);

  const xpath = useMemo(() => {
    if (path.length === 0) {
      return '#text'; // For normal text inputs
    }

    let xpathBuilder = '';
    path.forEach((step) => {
      if (typeof step === 'number') {
        xpathBuilder += `[${step + 1}]`;
      } else {
        if (xpathBuilder !== '') {
          xpathBuilder += '/';
        }
        xpathBuilder += step;
      }
    });
    return xpathBuilder;
  }, [path]);

  const charArray: string[] = data.toString().split('');

  useEffect(() => {
    for (let index = 0; index < predictions.length; index++) {
      const prediction = predictions[index];
      if (
        data
          .toString()
          .substring(
            prediction.location.local_char_span.start,
            prediction.location.local_char_span.end
          ) === prediction.location.value.trim()
      ) {
        setIsPredictionIndex(index);
      }
    }
  }, [data, predictions.length]);

  //Mouse event handlers for selection
  const handleMouseDown = (index: number) => {
    setStart(index);
    setEnd(null);
    setIsSelecting(true);
  };

  const handleMouseEnter = (index: number) => {
    if (isSelecting && start !== null) {
      setEnd(index);
    }
  };

  const handleMouseUp = () => {
    if (start !== null && end !== null) {
      setIsSelecting(false);
      setTagSelectorOpen(true);
    }
  };

  // Tag selection handler
  const handleTagSelect = (tag: string) => {
    if (start !== null && end !== null) {
      const normalizedStart = Math.min(start, end + 1);
      const normalizedEnd = Math.max(start, end + 1);
      const selectedText = data.toString().trim().substring(normalizedStart, normalizedEnd);
      onSelectionComplete({
        start: normalizedStart,
        end: normalizedEnd,
        xpath,
        tag,
        value: selectedText,
      });
    }
    setTagSelectorOpen(false);
    setStart(null);
    setEnd(null);
  };

  // Render logic with selection
  return (
    <div
      style={{
        width: 'fit-content',
        display: 'flex',
        flexDirection: 'row',
        justifyContent: 'start',
        flexWrap: 'wrap',
        userSelect: 'none',
      }}
    >
      {charArray.map((token, index) => {
        const isSelected =
          start !== null &&
          end !== null &&
          index >= Math.min(start, end) &&
          index <= Math.max(start, end);

        // Prediction highlighting
        const predictionIndexSpan =
          predictionIndex !== -1 &&
          predictions[predictionIndex] !== undefined &&
          index >= predictions[predictionIndex].location.local_char_span.start &&
          index <= predictions[predictionIndex].location.local_char_span.end - 1;

        return (
          <>
            <span
              key={index}
              onMouseDown={() => handleMouseDown(index)}
              onMouseEnter={() => handleMouseEnter(index)}
              onMouseUp={handleMouseUp}
              style={{
                backgroundColor: isSelected
                  ? 'rgba(153, 227, 181, 0.5)'
                  : predictionIndexSpan
                    ? 'rgba(255, 255, 0, 0.3)'
                    : 'transparent',
                cursor: 'text',
              }}
            >
              {token === ' ' ? '\u00A0' : token}
            </span>
            {predictionIndex !== -1 &&
              predictions[predictionIndex] &&
              index === predictions[predictionIndex].location.local_char_span.end - 1 && (
                <span
                  className="font-semibold text-red-500"
                  style={{ backgroundColor: 'rgba(255, 255, 0, 0.3)' }}
                >
                  {'\u00A0'}
                  {'[' + predictions[predictionIndex].label + ']'}
                </span>
              )}
          </>
        );
      })}
      {tagSelectorOpen && (
        <TagSelector
          open={tagSelectorOpen}
          choices={choices}
          onSelect={handleTagSelect}
          predictions={predictions}
        />
      )}
    </div>
  );
}

function XMLObjectRenderer({
  data,
  path,
  tag,
  choices,
  predictions,
  onSelectionComplete,
}: XMLObjectRendererProps) {
  console.log('Paht in ObjectRenderer: ', path);
  const attrs = Object.keys(data).filter((key) => key.startsWith(ATTRIBUTE_PREFIX));
  const numKeys = Object.keys(data).length;
  const emptyChild = numKeys === 0 || (numKeys - attrs.length === 1 && data['#text'] === '');
  const hasChild = numKeys === 0 || (numKeys > attrs.length && !emptyChild);
  const childInSameLine = !(typeof data === 'object') || data['#text'] !== undefined;
  const flexDirection = childInSameLine ? 'row' : 'column';
  const marginLeft = childInSameLine ? '0' : '20px';
  return (
    <div
      style={{
        display: 'flex',
        flexDirection,
        flexWrap: 'wrap',
        justifyContent: 'start',
      }}
    >
      <div style={{ display: 'flex', flexDirection: 'row' }}>
        <span>{`<${tag} `}</span>
        {attrs.map((key) => (
          <XMLAttributeRenderer
            key={key}
            attr={key}
            data={data[key]}
            path={path}
            choices={choices}
            predictions={predictions}
            onSelectionComplete={onSelectionComplete}
          />
        ))}
        {hasChild ? <span>{`>`}</span> : <span style={{ marginLeft: SPACE }}>{`/>`}</span>}
      </div>
      {hasChild && (
        <>
          <div style={{ marginLeft }}>
            <XMLRenderer
              data={data}
              path={path}
              choices={choices}
              predictions={predictions}
              onSelectionComplete={onSelectionComplete}
            />
          </div>
          <span style={{ width: 'fit-content' }}>{`</${tag}>`}</span>
        </>
      )}
    </div>
  );
}

export function XMLRenderer({
  data,
  path,
  choices,
  predictions,
  onSelectionComplete,
}: XMLRendererProps) {
  if (typeof data === 'string') {
    // Data is a string, render it directly
    return (
      <XMLValueRenderer
        data={data}
        path={path}
        choices={choices}
        predictions={predictions}
        onSelectionComplete={onSelectionComplete}
      />
    );
  }

  if (data['#text']) {
    // Data has a #text key, render the text
    return (
      <XMLValueRenderer
        data={data['#text']}
        path={path}
        choices={choices}
        predictions={predictions}
        onSelectionComplete={onSelectionComplete}
      />
    );
  }

  if (!(typeof data === 'object')) {
    return (
      <XMLValueRenderer
        data={data}
        path={path}
        choices={choices}
        predictions={predictions}
        onSelectionComplete={onSelectionComplete}
      />
    );
  }

  const childKeys = Object.keys(data).filter((key) => !key.startsWith(ATTRIBUTE_PREFIX));

  return (
    <div style={{ display: 'flex', flexDirection: 'column', width: 'fit-content' }}>
      {childKeys.map((key, index) => {
        if (key == '#text') {
          return (
            <XMLValueRenderer
              key={index}
              data={data[key]}
              path={path}
              choices={choices}
              predictions={predictions}
              onSelectionComplete={onSelectionComplete}
            />
          );
        }

        if (Array.isArray(data[key])) {
          return (
            <>
              {data[key].map((child, index) => (
                <XMLObjectRenderer
                  key={index}
                  data={child}
                  path={[...path, key, index]}
                  tag={key}
                  choices={choices}
                  predictions={predictions}
                  onSelectionComplete={onSelectionComplete}
                />
              ))}
            </>
          );
        }

        return (
          <XMLObjectRenderer
            key={index}
            data={data[key]}
            path={[...path, key]}
            tag={key}
            choices={choices}
            predictions={predictions}
            onSelectionComplete={onSelectionComplete}
          />
        );
      })}
    </div>
  );
}
