import { XMLParser, XMLValidator } from 'fast-xml-parser';
import {
  ChangeEvent,
  KeyboardEvent,
  ReactNode,
  useContext,
  useEffect,
  useMemo,
  useState
} from 'react';
import { ClickContext } from './clickContext';
import Fuse from 'fuse.js';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
// import { Feedback } from './useBackend';

export const ATTRIBUTE_PREFIX = '@_';
export const INDENT = '20px';
export const SPACE = '5px';

// export interface FrontendFeedback extends Feedback {
//   text: string;
// }

function replaceWhitespaceWithSpace(text: string): string {
  // Replace all whitespace characters with a single space
  return text.replace(/\s+/g, ' ').trim();
}

function cleanText(text: string): string {
  // Replace specific punctuation characters with a space
  return text.replace(/[:|"<>'\/\\,=%)(}{&]/g, ' ');
}

function cleanXMLText(obj: any): void {
  if (typeof obj === 'object' && obj !== null) {
    for (const key in obj) {
      if (key === '#text' && typeof obj[key] === 'string') {
        obj[key] = replaceWhitespaceWithSpace(cleanText(obj[key]));
      } else if (Array.isArray(obj[key])) {
        obj[key].forEach((item: any) => cleanXMLText(item));
      } else if (typeof obj[key] === 'object') {
        cleanXMLText(obj[key]);
      }
      // Optionally clean attribute values
      else if (key.startsWith(ATTRIBUTE_PREFIX) && typeof obj[key] === 'string') {
        obj[key] = cleanText(obj[key]);
      }
    }
  }
}

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
  cleanXMLText(parsedData);

  return parsedData;
}

interface XMLRendererProps {
  data: Record<string, any>;
  path: (string | number)[];
  choices: string[];
  // onFeedback: (feedback: FrontendFeedback) => void;
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
}

export function TagSelector({ open, choices, onSelect }: TagSelectorProps) {
  const defaultOptions = useMemo(
    () => choices.map((label) => ({ label, new: false })),
    [choices]
  );

  const [fuse, setFuse] = useState(new Fuse(choices));
  const [options, setOptions] =
    useState<{ label: string; new: boolean }[]>(defaultOptions);
  const [selectedIndex, setSelectedIndex] = useState<number | null>(null);

  useEffect(() => {
    setFuse(new Fuse(choices));
    setOptions(defaultOptions);
  }, [choices]);

  const handleInputChange = (e: ChangeEvent<HTMLInputElement>) => {
    const query = e.target.value;
    const searchResults =
      query !== '' ? fuse.search(query).map((val) => val.item) : choices;
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
          prev === null
            ? options.length - 1
            : Math.min(options.length - 1, prev - 1)
        );
        break;
      case 'ArrowDown':
        console.log('ArrowDown key pressed');
        setSelectedIndex((prev) =>
          prev === null ? 0 : Math.min(options.length, prev + 1)
        );
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
                fontWeight: 'bold'
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
              makeDropdownMenuItem(
                index,
                val.label,
                val.new,
                index === selectedIndex
              )
            )}
          </div>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}

export function ClickyThing(props: any) {
  const [hover, setHover] = useState(false);
  const background = props.selected
    ? 'rgba(153, 227, 181, 1.0)'
    : hover
      ? 'rgba(153, 227, 181, 0.5)'
      : '';
  return (
    <span
      style={{ cursor: 'pointer' }}
      onMouseDown={(e) => {
        e.stopPropagation();
        props.onMouseDown();
      }}
      onMouseOver={(e) => {
        setHover(true);
        e.stopPropagation();
        props.onMouseOver();
      }}
      onMouseUp={(e) => {
        e.stopPropagation();
        props.onMouseUp();
      }}
      onMouseLeave={(e) => {
        setHover(false);
        e.stopPropagation();
      }}
    >
      <span
        className={'bg-muted'}
        style={{
          flexWrap: 'nowrap',
          textWrap: 'nowrap',
          margin: '0 2px',
          padding: '2px 2px',
          borderRadius: '4px',
          background,
          userSelect: 'none',
          transition: '0.1s'
        }}
        onClick={props.onClick}
      >
        {props.children}
      </span>
    </span>
  );
}

function XMLAttributeRenderer({
  data,
  path,
  attr,
  choices,
  // onFeedback
}: XMLAttributeRendererProps) {
  const key = attr.substring(ATTRIBUTE_PREFIX.length);
  let dataString = JSON.stringify(data);
  dataString = dataString.substring(1, dataString.length - 1);
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'row',
        gap: '',
        marginLeft: SPACE,
        flexWrap: 'nowrap'
      }}
    >
      {key}=&quot;
      <XMLValueRenderer
        data={dataString}
        path={path}
        attr={key}
        choices={choices}
      // onFeedback={onFeedback}
      />
      &quot;
    </div>
  );
}

function XMLValueRenderer({
  data,
  path,
  attr,
  choices,
  // onFeedback
}: XMLValueRendererProps) {
  const [start, setStart] = useState<number | null>(null);
  const [end, setEnd] = useState<number | null>(null);
  const [range, setRange] = useState<[number, number] | null>(null);

  const click = useContext(ClickContext);

  const tokens = useMemo(
    () => (typeof data === 'string' ? data.split(/\s+/) : [data.toString()]),
    [data]
  );

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

  const clickKey = useMemo(() => `${xpath}:${attr}`, [xpath, attr]);

  const finalizeSelection = () => {
    if (start !== null && end !== null) {
      setRange([Math.min(start, end), Math.max(start, end)]);
      setStart(null);
      setEnd(null);
    }
  };

  const selected = (index: number) => {
    const withinStartEnd =
      start !== null &&
      end !== null &&
      index >= Math.min(start, end) &&
      index <= Math.max(start, end);
    const withinRange =
      range !== null && index >= range[0] && index <= range[1];
    return withinStartEnd || withinRange;
  };

  function arrayFromRange(x: number, y: number): number[] {
    return Array.from({ length: y - x + 1 }, (_, i) => x + i);
  }

  const submit = (newLabel: string) => {
    if (!range) {
      return;
    }
    // onFeedback({
    //   xpath,
    //   attr: attr || null,
    //   indices: arrayFromRange(range[0], range[1]),
    //   ntokens: tokens.length,
    //   label: newLabel,
    //   text: tokens.slice(range[0], range[1] + 1).join(' '),
    //   user_provided: true
    // });
    setStart(null);
    setEnd(null);
    setRange(null);
  };

  return (
    <div
      style={{
        width: 'fit-content',
        display: 'flex',
        flexDirection: 'row',
        justifyContent: 'start',
        flexWrap: 'wrap'
      }}
    >
      {tokens.map((token, index) => (
        <>
          <ClickyThing
            key={`clicky${index}`}
            selected={selected(index) && clickKey === click.key}
            onMouseDown={() => {
              click.register(clickKey);
              setStart(index);
              setRange(null);
            }}
            onMouseOver={() => {
              setEnd(index);
            }}
            onMouseUp={finalizeSelection}
          >
            {token}
          </ClickyThing>
          {
            <TagSelector
              key={`token${index}`}
              open={
                selected(index) &&
                clickKey === click.key &&
                range !== null &&
                index === range[1]
              }
              onSelect={(newLabel: string) => submit(newLabel)}
              choices={choices}
            />
          }
        </>
      ))}
    </div>
  );
}

function XMLObjectRenderer({
  data,
  path,
  tag,
  choices,
  // onFeedback
}: XMLObjectRendererProps) {
  const attrs = Object.keys(data).filter((key) =>
    key.startsWith(ATTRIBUTE_PREFIX)
  );
  const numKeys = Object.keys(data).length;
  const emptyChild =
    numKeys === 0 || (numKeys - attrs.length === 1 && data['#text'] === '');
  const hasChild = numKeys === 0 || (numKeys > attrs.length && !emptyChild);
  const childInSameLine =
    !(typeof data === 'object') || data['#text'] !== undefined;
  const flexDirection = childInSameLine ? 'row' : 'column';
  const marginLeft = childInSameLine ? '0' : '20px';
  return (
    <div
      style={{
        display: 'flex',
        flexDirection,
        flexWrap: 'wrap',
        justifyContent: 'start'
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
          // onFeedback={onFeedback}
          />
        ))}
        {hasChild ? (
          <span>{`>`}</span>
        ) : (
          <span style={{ marginLeft: SPACE }}>{`/>`}</span>
        )}
      </div>
      {hasChild && (
        <>
          <div style={{ marginLeft }}>
            <XMLRenderer
              data={data}
              path={path}
              choices={choices}
            // onFeedback={onFeedback}
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
  // onFeedback
}: XMLRendererProps) {
  console.log("print inside xml renderer");
  console.log(data);
  console.log(path);
  console.log(choices);

  if (typeof data === 'string') {
    // Data is a string, render it directly
    return (
      <XMLValueRenderer
        data={data}
        path={path}
        choices={choices}
      // onFeedback={onFeedback}
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
      // onFeedback={onFeedback}
      />
    );
  }

  if (!(typeof data === 'object')) {
    return (
      <XMLValueRenderer
        data={data}
        path={path}
        choices={choices}
      // onFeedback={onFeedback}
      />
    );
  }

  const childKeys = Object.keys(data).filter(
    (key) => !key.startsWith(ATTRIBUTE_PREFIX)
  );
  return (
    <div
      style={{ display: 'flex', flexDirection: 'column', width: 'fit-content' }}
    >
      {childKeys.map((key, index) => {
        if (key == '#text') {
          return (
            <XMLValueRenderer
              key={index}
              data={data[key]}
              path={path}
              choices={choices}
            // onFeedback={onFeedback}
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
                // onFeedback={onFeedback}
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
          // onFeedback={onFeedback}
          />
        );
      })}
    </div>
  );
}
