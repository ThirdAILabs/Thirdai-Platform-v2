'use client'

import { Button } from "@/components/ui/button";
import { trainTokenClassifier } from "@/lib/backend";

export default function Page() {
    return <div>
        <Button onClick={() => trainTokenClassifier(`name-${Date.now()}`, "Name detection", [{name: "NAME", example: "Benito Geordie", description: ""}])}>
            Train token classification model
        </Button>
    </div>
}