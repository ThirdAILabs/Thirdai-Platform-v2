import { ReactNode } from 'react';

interface TableProps {
  className?: string;
  children: ReactNode;
}

export function Table({ className, children }: TableProps) {
  return (
    <table className={`w-full caption-bottom text-sm ${className || ''}`}>
      {children}
    </table>
  );
}

interface TableHeaderProps {
  className?: string;
  children: ReactNode;
}

export function TableHeader({ className, children }: TableHeaderProps) {
  return (
    <thead className={`${className || ''}`}>
      {children}
    </thead>
  );
}

interface TableRowProps {
  className?: string;
  children: ReactNode;
}

export function TableRow({ className, children }: TableRowProps) {
  return (
    <tr className={`border-b transition-colors hover:bg-muted/50 ${className || ''}`}>
      {children}
    </tr>
  );
}

interface TableHeadProps {
  className?: string;
  children: ReactNode;
}

export function TableHead({ className, children }: TableHeadProps) {
  return (
    <th className={`h-12 px-4 text-left align-middle font-medium text-muted-foreground ${className || ''}`}>
      {children}
    </th>
  );
}

interface TableBodyProps {
  className?: string;
  children: ReactNode;
}

export function TableBody({ className, children }: TableBodyProps) {
  return (
    <tbody className={`${className || ''}`}>
      {children}
    </tbody>
  );
}

interface TableCellProps {
  className?: string;
  children: ReactNode;
  colSpan?: number;
  style?: React.CSSProperties;
}

export function TableCell({ className, children, colSpan, style }: TableCellProps) {
  return (
    <td className={`p-4 align-middle ${className || ''}`} colSpan={colSpan} style={style}>
      {children}
    </td>
  );
} 