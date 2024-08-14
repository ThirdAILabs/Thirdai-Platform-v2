'use client'

import { Button } from "@/components/ui/button"
import { updateModel } from "@/lib/backend";
import { useRouter, useSearchParams } from "next/navigation"

export default function UpdateButton() {
    const params = useSearchParams();
    
    function handleUpdate() {
        updateModel(params.get("id") as string);
    }
    
    return <div style={{display: "flex", justifyContent: "center", marginTop: "20px", marginBottom: "20vh"}}>
        <Button onClick={handleUpdate}>Update model with feedback</Button>
    </div>
}