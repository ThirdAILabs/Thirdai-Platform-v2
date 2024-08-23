import { XMLParser } from "fast-xml-parser";
import { useState } from "react";

export const ATTRIBUTE_PREFIX = '@_';
export const INDENT = "20px";
export const SPACE = "5px";

export function parseXML(xml: string) {
  const parser = new XMLParser({
    ignoreAttributes: false,
    allowBooleanAttributes: false,
    attributeNamePrefix: ATTRIBUTE_PREFIX,
    // preserveOrder: true,
    alwaysCreateTextNode: true,
  });
  return parser.parse(xml);
}

interface XMLRendererProps {
  data: Record<string, any>;
  path: (string | number)[];
  onClickNode: (node: Record<string, any>) => void;
}

interface XMLAttributeRendererProps extends XMLRendererProps {
  attributeKey: string;
};

interface XMLValueRendererProps extends XMLRendererProps {
  data: any;
};

interface XMLObjectRendererProps extends XMLRendererProps {
  tag: string;
};

function ClickyThing(props: any) {
  const [hover, setHover] = useState(false);
  return <span
    style={{cursor: "pointer"}}
    onMouseOver={(e) => {
      setHover(true);
    }}
    onMouseLeave={(e) => {
      setHover(false);
    }}>
    <span
      className={"bg-muted"}
      style={{flexWrap: "nowrap", textWrap: "nowrap", margin: "0 2px", padding: "2px 2px", borderRadius: "4px", background: hover ? "#99E3B5" : "", transition: "0.1s"}}
      onClick={props.onClick}>
      {props.children}
    </span>
  </span> 
    
}

function XMLAttributeRenderer({ data, path, attributeKey, onClickNode }: XMLAttributeRendererProps) {
  const key = attributeKey.substring(ATTRIBUTE_PREFIX.length);
  const handleClick = () => {
    onClickNode({
      path,
      key,
      data,
    });
  }
  return <div style={{display: "flex", flexDirection: "row", gap: "", marginLeft: SPACE, flexWrap: "nowrap"}}>
    {key}=<ClickyThing onClick={handleClick}>{JSON.stringify(data)}</ClickyThing>
  </div>
}

function XMLValueRenderer({ data, path, onClickNode }: XMLValueRendererProps) {
  const tokens = typeof data === "string" ? data.split(/\s+/) : [data.toString()];
  const handleClick = (token: string, index: number) => () => {
    onClickNode({
      path,
      token,
      index,
    });
  }
  return <div style={{width: "fit-content", display: "flex", flexDirection: "row", justifyContent: "start", flexWrap: "wrap"}}>
    {
      tokens.map((token, index) => <ClickyThing key={index} onClick={handleClick(token, index)}>{token}</ClickyThing>)
    }
  </div>;
}

function XMLObjectRenderer({ data, path, tag, onClickNode }: XMLObjectRendererProps) {
  if (tag === "EventID") {
    console.log(data);
  }
  const attributeKeys = Object.keys(data).filter(key => key.startsWith(ATTRIBUTE_PREFIX));
  const numKeys = Object.keys(data).length;
  const hasChild = numKeys === 0 || (numKeys > attributeKeys.length);
  const childInSameLine = (!(typeof data === 'object') || data["#text"] !== undefined);
  const flexDirection = childInSameLine ? "row" : "column";
  const marginLeft = childInSameLine ? "0" : "20px";
  return <div style={{display: "flex", flexDirection, flexWrap: "wrap", justifyContent: "start"}}>
    <div style={{display: "flex", flexDirection: "row"}}>
      <span>{`<${tag} `}</span>
      {attributeKeys.map(key => <XMLAttributeRenderer key={key} attributeKey={key} data={data[key]} path={path} onClickNode={onClickNode} />)}
      {hasChild ? <span>{`>`}</span> : <span>{`/>`}</span>}
    </div>
    {hasChild && <>
        <div style={{marginLeft}}>
          <XMLRenderer data={data} path={path} onClickNode={onClickNode} />
        </div>
        <span style={{width: "fit-content"}}>{`</${tag}>`}</span>
      </>
    }
  </div>
}

export function XMLRenderer({ data, path, onClickNode }: XMLRendererProps) {
  if (!(typeof data === 'object')) {
    return (
      <XMLValueRenderer data={data} path={path} onClickNode={onClickNode} />
    );
  }
  
  const childKeys = Object.keys(data).filter(key => !key.startsWith(ATTRIBUTE_PREFIX));
  return <div style={{display: "flex", flexDirection: "column", width: "fit-content"}}>
    {
      childKeys.map((key) => {
        if (key == "#text") {
          return <XMLValueRenderer key={key} data={data[key]} path={path} onClickNode={onClickNode} />
        }

        if (Array.isArray(data[key])) {
          return <>
            {data[key].map((child, index) => (
              <XMLObjectRenderer key={index} data={child} path={[...path, key, index]} tag={key} onClickNode={onClickNode} />
            ))}
          </>
        }

        return <XMLObjectRenderer key={key} data={data[key]} path={[...path, key]} tag={key} onClickNode={onClickNode} />
      })
    }
  </div>
}

