import React, { useState, useMemo } from 'react';
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow
} from '@/components/ui/table';
import {
    Card,
    CardContent,
    CardHeader,
    CardTitle
} from '@/components/ui/card';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue
} from '@radix-ui/react-select';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Trash2, Edit2, Check, X } from 'lucide-react';

// Types for selections and predictions
interface Selection {
    start: number;
    end: number;
    xpath: string;
    tag?: string;
    text?: string;
}

interface Prediction {
    label: string;
    location: {
        local_char_span: { start: number; end: number };
        global_char_span: { start: number; end: number };
        xpath_location: { xpath: string; attribute: string | null };
        value: string;
    };
}

interface FeedbackDashboardProps {
    selections: Selection[];
    predictions: Prediction[];
    xmlText: string;
    onDeleteSelection?: (index: number) => void;
    onUpdateSelection?: (index: number, newTag: string) => void;
}

export function FeedbackDashboard({
    selections,
    predictions,
    xmlText,
    onDeleteSelection,
    onUpdateSelection
}: FeedbackDashboardProps) {
    // State for editing selections
    const [editingIndex, setEditingIndex] = useState<number | null>(null);
    const [editTag, setEditTag] = useState<string>('');

    // Compute selections with their corresponding text
    const enrichedSelections = useMemo(() => {
        return selections.map(selection => ({
            ...selection,
            text: xmlText.substring(selection.start, selection.end + 1)
        }));
    }, [selections, xmlText]);

    // Prepare tag options from predictions and current selections
    const tagOptions = useMemo(() => {
        const predictionTags = predictions.map(p => p.label);
        const selectionTags = [...new Set(selections.map(s => s.tag).filter(Boolean))];
        return Array.from(new Set([...predictionTags, ...selectionTags]));
    }, [predictions, selections]);

    // Handler for editing a selection
    const handleEditSelection = (index: number) => {
        setEditingIndex(index);
        setEditTag(enrichedSelections[index].tag || '');
    };

    // Handler for saving edited selection
    const handleSaveEdit = () => {
        if (editingIndex !== null && onUpdateSelection) {
            onUpdateSelection(editingIndex, editTag);
            setEditingIndex(null);
        }
    };

    // Handler for canceling edit
    const handleCancelEdit = () => {
        setEditingIndex(null);
    };

    return (
        <Card className="w-full max-w-4xl mx-auto">
            <CardHeader>
                <CardTitle>Feedback from this session</CardTitle>
            </CardHeader>
            <CardContent>
                <div className="grid gap-4">
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Text</TableHead>
                                <TableHead>Tag</TableHead>
                                <TableHead>Actions</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {enrichedSelections.map((selection, index) => (
                                <TableRow key={index}>
                                    {editingIndex === index ? (
                                        // Edit Mode
                                        <>
                                            <TableCell>{selection.text}</TableCell>
                                            <TableCell>
                                                <Select
                                                    value={editTag}
                                                    onValueChange={setEditTag}
                                                >
                                                    <SelectTrigger>
                                                        <SelectValue placeholder="Select tag" />
                                                    </SelectTrigger>
                                                    <SelectContent>
                                                        {tagOptions.map(tag => (
                                                            tag !== undefined && <SelectItem key={tag} value={tag}>
                                                                {tag}
                                                            </SelectItem>
                                                        ))}
                                                    </SelectContent>
                                                </Select>
                                            </TableCell>
                                            <TableCell>
                                                <div className="flex space-x-2">
                                                    <Button
                                                        variant="outline"
                                                        size="icon"
                                                        onClick={handleSaveEdit}
                                                    >
                                                        <Check className="h-4 w-4" />
                                                    </Button>
                                                    <Button
                                                        variant="outline"
                                                        size="icon"
                                                        onClick={handleCancelEdit}
                                                    >
                                                        <X className="h-4 w-4" />
                                                    </Button>
                                                </div>
                                            </TableCell>
                                        </>
                                    ) : (
                                        // View Mode
                                        <>
                                            <TableCell>{selection.text}</TableCell>
                                            <TableCell>
                                                <Badge variant="secondary">
                                                    {selection.tag || 'Untagged'}
                                                </Badge>
                                            </TableCell>
                                            <TableCell>
                                                <div className="flex space-x-2">
                                                    <Button
                                                        variant="outline"
                                                        size="icon"
                                                        onClick={() => handleEditSelection(index)}
                                                    >
                                                        <Edit2 className="h-4 w-4" />
                                                    </Button>
                                                    <Button
                                                        variant="destructive"
                                                        size="icon"
                                                        onClick={() => onDeleteSelection?.(index)}
                                                    >
                                                        <Trash2 className="h-4 w-4" />
                                                    </Button>
                                                </div>
                                            </TableCell>
                                        </>
                                    )}
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </div>
            </CardContent>
        </Card>
    );
}

export default FeedbackDashboard;